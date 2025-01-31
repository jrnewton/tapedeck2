package app

import (
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
	tape "tapedeck/internal/database/tape"
	"tapedeck/internal/lazy"
)

type ServerConfig struct {
	ServerDir        string `json:"serverDir"`
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

	config.ServerDir = filepath.Join(config.ServerDir)
	config.UserDir = filepath.Join(config.UserDir)

	return config, nil
}

func RunServer(jsonConfigPath string) (rc int, err error) {
	log.Println("enter RunServer", jsonConfigPath)
	defer log.Println("exit RunServer")

	config, configErr := readConfig(jsonConfigPath)
	if configErr != nil {
		return 209, configErr
	}

	// config validation
	log.Println("validate server directory")
	_, dirErr := os.Stat(config.ServerDir)
	if dirErr != nil {
		return 148, dirErr
	}

	log.Println("validate user directory")
	_, dirErr2 := os.Stat(config.UserDir)
	if dirErr2 != nil {
		return 95, dirErr2
	}

	log.Println("validate server listen address")
	if strings.Count(config.ServerListenAddr, ":") != 1 {
		return 92, fmt.Errorf("invalid server listen address: %v", config.ServerListenAddr)
	}

	templateDir, tmplErr := checkDir(config.ServerDir, "templates")
	if tmplErr != nil {
		return 210, tmplErr
	}

	cache := config.ProductionMode
	log.Println("using template directory", templateDir, "and cache", cache)
	tmplEngine := newTemplateEngine(templateDir, cache)
	initErr := tmplEngine.Init()
	if initErr != nil {
		return 217, fmt.Errorf("template engine init failed: %w", initErr)
	}

	staticDir, staticErr := checkDir(config.ServerDir, "static")
	if staticErr != nil {
		return 222, staticErr
	}
	log.Println("using static directory", staticDir)

	log.Println("using db file", config.DbFile)
	db := database.New(config.DbFile)

	log.Println("open database")
	db.Open(true)
	defer db.Close(false)

	log.Println("server verification complete")

	// Open routes
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	http.HandleFunc("/", makeRootHandler(tmplEngine))

	// Secure routes
	http.HandleFunc("/s/list", makeListHandler(db, tmplEngine))
	http.HandleFunc("/s/playback", makePlaybackHandler(db, tmplEngine))
	http.HandleFunc("/s/record", makeRecordHandler(db, tmplEngine))

	//TODO: define middelware to dump headers based on log level.
	//TODO: extract `X-Email` header value and maybe `X-User`.
	//X-Email:[rocketnewton@gmail.com]
	//X-User:[112165920196384629909]

	log.Println("server starting on", config.ServerListenAddr)
	err = http.ListenAndServe(config.ServerListenAddr, nil)
	if err != nil {
		return 286, err
	} else {
		return 0, nil
	}
}

// CheckerUser extracts the user from the request headers
// and determines if the user is allowed to proceed
// func CheckUser(db *database.Database, r *http.Request) (*userpkg.User, error) {
// 	email := r.Header.Get("X-EMAIL")

// 	if email == "" {
// 		return nil, fmt.Errorf("user not authenticated? X-EMAIL header not found")
// 	}

// 	return userpkg.GetUserByEmail(db, email)
// }

type handler func(http.ResponseWriter, *http.Request)

func makeRootHandler(tmplEngine *TemplateEngine) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("enter rootHandler", r.URL.String())
		defer log.Println("exit rootHandler")
		//http.Redirect(w, r, "/testcases", http.StatusTemporaryRedirect)

		defer func() {
			if r := recover(); r != nil {
				msg := fmt.Sprintf("panic in handleRoot: %v\n%v", r, string(debug.Stack()))
				log.Println(msg)
				http.Error(w, msg, 500)
			}
		}()

		bytes, evalErr := tmplEngine.Eval("index.html", "")
		if evalErr != nil {
			http.Error(w, evalErr.Error(), 500)
			return
		}

		log.Println("write bytes to response")
		w.Write(bytes)
	}
}

func makeListHandler(db *database.Database, tmplEngine *TemplateEngine) handler {
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

		tapes, getErr := tape.GetAllTapes(db)
		log.Println("GetAllTapes returned items: ", len(tapes))
		for i, v := range tapes {
			log.Println(i, v)
		}

		if getErr != nil {
			http.Error(w, getErr.Error(), 500)
			return
		}

		bytes, evalErr := tmplEngine.Eval("list.html", tapes)
		if evalErr != nil {
			http.Error(w, evalErr.Error(), 500)
			return
		}

		log.Println("write bytes to response")
		w.Write(bytes)
	}
}

func makePlaybackHandler(db *database.Database, tmplEngine *TemplateEngine) handler {
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

		t, getErr := tape.GetTape(db, id)
		if getErr != nil {
			http.Error(w, getErr.Error(), 500)
			return
		}

		bytes, evalErr := tmplEngine.Eval("playback.html", t)
		if evalErr != nil {
			http.Error(w, evalErr.Error(), 500)
			return
		}

		log.Println("write bytes to response")
		w.Write(bytes)
	}
}

func makeRecordHandler(db *database.Database, tmplEngine *TemplateEngine) handler {
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

		bytes, evalErr := tmplEngine.Eval("record.html", nil)
		if evalErr != nil {
			http.Error(w, evalErr.Error(), 500)
			return
		}

		log.Println("write bytes to response")
		w.Write(bytes)
	}
}

// TemplateEngine is built upon [template.Template]. It is used to cache and evaluate templates.
// Call [New] to create a new instance and call [TemplateEngine.Init] before first use.
// Failure call [TemplateEngine.Init] before first use will result in a panic.
type TemplateEngine struct {
	dir         string
	root        *template.Template
	cache       bool
	initialized bool
}

func (t TemplateEngine) String() string {
	var rootName = "nil"
	if t.root != nil {
		rootName = t.root.Name()
	}

	return fmt.Sprintf("template engine %v %v %v", t.dir, t.cache, rootName)
}

func (t *TemplateEngine) lookup(name string) (*template.Template, error) {
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

func newTemplateEngine(templateDir string, cache bool) *TemplateEngine {
	log.Println("enter NewTemplateEngine", templateDir, cache)
	defer log.Println("exit NewTemplateEngine")

	t := new(TemplateEngine)
	t.dir = templateDir
	t.cache = cache
	log.Println("template engine", t)
	return t
}

func (engine *TemplateEngine) Init() (err error) {
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

func (engine *TemplateEngine) cacheAll() error {
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

// Eval evaluates the template with the given fileName using the given data.
func (engine *TemplateEngine) Eval(fileName string, data any) ([]byte, error) {
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
