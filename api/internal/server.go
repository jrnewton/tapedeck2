package app

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"tapedeck/internal/database"
	"tapedeck/internal/database/tape"
	"tapedeck/internal/database/user"
	"tapedeck/internal/lazy"
)

const logHeaders = false

// Typically i define return codes
// based on the line number which is a
// cheap way to ensure uniqueness.
// I like return unique values as it provides
// a quick pointer to where things went
// off the rails.
type ReturnCode int

func (rc ReturnCode) Code() int {
	return int(rc)
}

const (
	RcOkay ReturnCode = iota
	_                 // skip 1 thru 4 which are warning levels.
	_
	_
	_
	RcConfigFile
	RcWebDir
	RcUserDir
	RcListenAddress
	RcTemplatesDir
	RcTemplateEngine
	RcStaticDir
	RcListenAndServe
)

type contextKey string

const userKey contextKey = "user"

type ServerConfig struct {
	WebDir           string `json:"webDir"`
	UserDir          string `json:"userDir"`
	ServerListenAddr string `json:"serverListenAddr"`
	DbFile           string `json:"dbFile"`
	ProductionMode   bool   `jons:"productionMode"`
}

// checkDir will join the parentDir to dirName and check that the new dir exists.
func checkDir(parentDir string, dirName string) (fullDir string, err error) {
	log.Println("enter checkDir", parentDir, dirName)
	defer log.Println("exit checkDir")

	fullDir = filepath.Join(parentDir, dirName)
	_, err = os.Stat(fullDir)
	if err != nil {
		err = fmt.Errorf("failed to verify %s in %s: %w", dirName, parentDir, err)
	}
	return
}

func readConfig(jsonFilePath string) (ServerConfig, error) {
	log.Println("enter readConfig", jsonFilePath)
	defer log.Println("exit readConfig")

	var config ServerConfig

	bytes, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return config, err
	}

	config.WebDir = filepath.Join(config.WebDir)
	config.UserDir = filepath.Join(config.UserDir)

	return config, nil
}

func RunServer(jsonConfigPath string) (rc ReturnCode, err error) {
	log.Println("enter RunServer", jsonConfigPath)
	defer log.Println("exit RunServer")

	config, configErr := readConfig(jsonConfigPath)
	if configErr != nil {
		return RcConfigFile, configErr
	}

	// config validation
	log.Println("validate web directory")
	_, dirErr := os.Stat(config.WebDir)
	if dirErr != nil {
		return RcWebDir, dirErr
	}

	log.Println("validate user directory")
	_, dirErr2 := os.Stat(config.UserDir)
	if dirErr2 != nil {
		if !config.ProductionMode {
			log.Println("user directory not found, run `make user` to create it")
		}
		return RcUserDir, dirErr2
	}

	log.Println("validate server listen address")
	if strings.Count(config.ServerListenAddr, ":") != 1 {
		return RcListenAddress, fmt.Errorf("invalid server listen address: %v", config.ServerListenAddr)
	}

	templateDir, tmplErr := checkDir(config.WebDir, "templates")
	if tmplErr != nil {
		return RcTemplatesDir, tmplErr
	}

	cache := config.ProductionMode
	log.Println("using template directory", templateDir, "and cache", cache)
	tmplEngine := newTemplateEngine(templateDir, cache)
	initErr := tmplEngine.init()
	if initErr != nil {
		return RcTemplateEngine, fmt.Errorf("template engine init failed: %w", initErr)
	}

	staticDir, staticErr := checkDir(config.WebDir, "static")
	if staticErr != nil {
		return RcStaticDir, staticErr
	}
	log.Println("using static directory", staticDir)

	log.Println("using db file", config.DbFile)
	db := database.New(config.DbFile)

	log.Println("open database")
	db.Open()
	defer db.Close()

	log.Println("server verification complete")

	// Open routes
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	http.HandleFunc("/", chain(makeLogger, makeRootHandler(tmplEngine)))

	// Secure routes
	http.HandleFunc("/s/list", chain(makeLogger, makeUserLookup(db), makeListHandler(db, tmplEngine)))
	http.HandleFunc("/s/playback", chain(makeLogger, makeUserLookup(db), makePlaybackHandler(db, tmplEngine)))
	http.HandleFunc("/s/record", chain(makeLogger, makeUserLookup(db), makeRecordHandler(db, tmplEngine)))

	log.Println("server starting on", config.ServerListenAddr)
	err = http.ListenAndServe(config.ServerListenAddr, nil)
	if err != nil {
		return RcListenAndServe, err
	} else {
		return RcOkay, nil
	}
}

// middleware is a [http.HandlerFunc] that can be chained together with other
// functions to build so-called "middleware"
type middleware func(http.HandlerFunc) http.HandlerFunc

// chain takes middleware functions
func chain(m ...middleware) http.HandlerFunc {
	if len(m) == 0 {
		panic("at least one middleware function required")
	}

	wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//log.Printf("final\n")
	})

	for i := len(m) - 1; i >= 0; i-- {
		wrapped = m[i](wrapped)
	}

	return wrapped
}

func getUserFromRequest(w http.ResponseWriter, r *http.Request) *user.User {
	u, ok := r.Context().Value(userKey).(*user.User)
	if !ok {
		http.Error(w, "user not found in context", http.StatusInternalServerError)
		return nil
	} else {
		log.Printf("found user %v\n", u)
		return u
	}
}

// makeUserLookup creates a [middleware] function that will retrieve the authenticated user
// and make it available in the request context.
func makeUserLookup(db *database.Database) middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			u := r.Context().Value(userKey)

			if u == nil {
				userEmail := r.Header.Get("X-EMAIL")
				userId := r.Header.Get("X-USER")
				log.Printf("X-EMAIL is %s\n", userEmail)
				log.Printf("X-USER is %s\n", userId)

				if userEmail == "" {
					log.Printf("X-EMAIL header not set\n")
					w.WriteHeader(http.StatusNotFound)
					return
				}

				u, err := user.GetByEmail(db, userEmail)

				if err != nil {
					log.Printf("error getting user object: %v\n", err)
					w.WriteHeader(http.StatusNotFound)
					return
				}

				if u == nil {
					log.Printf("user object not found for email %q\n", userEmail)
					w.WriteHeader(http.StatusNotFound)
					return
				}

				ctx := context.WithValue(r.Context(), userKey, u)
				newRequest := r.WithContext(ctx)

				log.Printf("user added to request context %v\n", ctx)
				next.ServeHTTP(w, newRequest)
			} else {
				log.Printf("user found in request context %v\n", u)
				next.ServeHTTP(w, r)
			}
		}
	}
}

// makeLogger is a [middleware] function that logs all header values.
func makeLogger(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if logHeaders {
			log.Printf("-- HEADERS ---------------")
			for key, val := range r.Header {
				for _, v := range val {
					// user agent is roughly 130
					if len(v) > 130 {
						v = v[:130] + "..."
					}
					log.Printf("  %s=%s", key, v)
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func makeRootHandler(tmplEngine *templateEngine) middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log.Println("enter rootHandler", r.URL.String())
			defer log.Println("exit rootHandler")

			defer func() {
				if r := recover(); r != nil {
					msg := fmt.Sprintf("panic in handleRoot: %v\n%v", r, string(debug.Stack()))
					log.Println(msg)
					http.Error(w, msg, 500)
				}
			}()

			bytes, evalErr := tmplEngine.eval("index.html", "")
			if evalErr != nil {
				http.Error(w, evalErr.Error(), 500)
				return
			}

			log.Println("write bytes to response")
			w.Write(bytes)

			next.ServeHTTP(w, r)
		}
	}
}

func makeListHandler(db *database.Database, tmplEngine *templateEngine) middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log.Println("enter MakeListHandler", r.URL.String())
			defer log.Println("exit MakeListHandler")
			defer func() {
				if r := recover(); r != nil {
					msg := fmt.Sprintf("panic in MakeListHandler: %v\n%v", r, string(debug.Stack()))
					log.Println(msg)
					http.Error(w, msg, 500)
				}
			}()

			u := getUserFromRequest(w, r)
			if u == nil {
				return
			}

			tapes, getErr := tape.GetTapesForUser(u.Id, db)
			log.Println("GetAllTapes returned items: ", len(tapes))
			for i, v := range tapes {
				log.Println(i, v)
			}

			if getErr != nil {
				http.Error(w, getErr.Error(), 500)
				return
			}

			bytes, evalErr := tmplEngine.eval("list.html", tapes)
			if evalErr != nil {
				http.Error(w, evalErr.Error(), 500)
				return
			}

			log.Println("write bytes to response")
			w.Write(bytes)
		}
	}
}

func makePlaybackHandler(db *database.Database, tmplEngine *templateEngine) middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log.Println("enter MakePlaybackHandler", r.URL.String())
			defer log.Println("exit MakePlaybackHandler")

			defer func() {
				if r := recover(); r != nil {
					msg := fmt.Sprintf("panic in MakePlaybackHandler: %v\n%v", r, string(debug.Stack()))
					log.Println(msg)
					http.Error(w, msg, 500)
				}
			}()

			params := r.URL.Query()
			id, parseErr := strconv.ParseInt(params.Get("id"), 10, 64)
			if parseErr != nil {
				http.Error(w, fmt.Errorf("tape id failed to parse as number: %w", parseErr).Error(), 500)
				return
			}

			t, getErr := tape.GetTape(id, db)
			if getErr != nil {
				http.Error(w, getErr.Error(), 500)
				return
			}

			bytes, evalErr := tmplEngine.eval("playback.html", t)
			if evalErr != nil {
				http.Error(w, evalErr.Error(), 500)
				return
			}

			log.Println("write bytes to response")
			w.Write(bytes)
		}
	}
}

func makeRecordHandler(_ *database.Database, tmplEngine *templateEngine) middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log.Println("enter MakeRecordHandler", r.URL.String())
			defer log.Println("exit MakeRecordHandler")

			defer func() {
				if r := recover(); r != nil {
					msg := fmt.Sprintf("panic in MakeRecordHandler: %v\n%v", r, string(debug.Stack()))
					log.Println(msg)
					http.Error(w, msg, 500)
				}
			}()

			u := getUserFromRequest(w, r)
			if u == nil {
				return
			}

			bytes, evalErr := tmplEngine.eval("record.html", nil)
			if evalErr != nil {
				http.Error(w, evalErr.Error(), 500)
				return
			}

			log.Println("write bytes to response")
			w.Write(bytes)
		}
	}
}

// templateEngine is built upon [template.Template]. It is used to cache and evaluate templates.
// Call [newTemplateEngine] to create a new instance and call [templateEngine.Init] before first use.
// Failure call [templateEngine.Init] before first use will result in a panic.
type templateEngine struct {
	dir         string
	root        *template.Template
	cache       bool
	initialized bool
}

func (t templateEngine) String() string {
	var rootName = "nil"
	if t.root != nil {
		rootName = t.root.Name()
	}

	return fmt.Sprintf("template engine %q %v %q", t.dir, t.cache, rootName)
}

func (t *templateEngine) lookup(name string) (*template.Template, error) {
	if !t.initialized {
		panic(fmt.Errorf("template engine not initialized"))
	}

	// When cache not enabled, build it on each lookup.
	if !t.cache {
		err := t.cacheAll()
		if err != nil {
			return nil, err
		}
	}

	return t.root.Lookup(name), nil
}

func newTemplateEngine(templateDir string, cache bool) *templateEngine {
	log.Println("enter NewTemplateEngine", templateDir, cache)
	defer log.Println("exit NewTemplateEngine")

	t := new(templateEngine)
	t.dir = templateDir
	t.cache = cache
	log.Println("template engine", t)
	return t
}

func (engine *templateEngine) init() (err error) {
	log.Println("enter Init", engine.dir, engine.cache)
	defer log.Println("exit Init")

	if engine.cache {
		err = engine.cacheAll()
	}

	if err == nil {
		engine.initialized = true
	}

	return
}

func (engine *templateEngine) cacheAll() error {
	log.Println("enter cacheAll", engine.dir, engine.cache)
	defer log.Println("exit cacheAll")

	glob := engine.dir + string(os.PathSeparator) + "*.html"
	log.Println("parsing templates with glob", glob)

	t, err := template.ParseGlob(glob)

	if err != nil {
		log.Println("ParseGlob gave error", err)
		return err
	}

	if t == nil {
		// should not happen
		panic(fmt.Errorf("ParseGlob gave nil template"))
	}

	engine.root = t
	return nil
}

// eval evaluates the template with the given fileName using the given data.
func (engine *templateEngine) eval(fileName string, data any) ([]byte, error) {
	log.Println("enter Eval", fileName)
	defer log.Println("exit Eval", fileName)

	tmpl, lookupErr := engine.lookup(fileName)
	if lookupErr != nil {
		return nil, fmt.Errorf("template lookup error: %w", lookupErr)
	}

	// Resolve the template to a temp buffer in order to deal with any
	// runtime errors during template processing.
	lw := new(lazy.Writer)

	log.Println("template execute")
	err := tmpl.Execute(lw, data)
	if err != nil {
		return nil, fmt.Errorf("template execute error: %w", err)
	}

	return lw.Bytes()
}
