package tapedeck

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
)

type handler func(http.ResponseWriter, *http.Request)

func MakeRootHandler(tmplEngine *TemplateEngine) handler {
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

func MakeListHandler(db *Database, tmplEngine *TemplateEngine) handler {
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

		tapes, getErr := GetAllTapes(db)
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

func MakePlaybackHandler(db *Database, tmplEngine *TemplateEngine) handler {
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
		id, parseErr := strconv.Atoi(params.Get("id"))
		if parseErr != nil {
			http.Error(w, fmt.Errorf("tape id failed to parse as number: %w", parseErr).Error(), 500)
			return
		}

		tape, getErr := GetTape(db, id)
		if getErr != nil {
			http.Error(w, getErr.Error(), 500)
			return
		}

		bytes, evalErr := tmplEngine.Eval("playback.html", tape)
		if evalErr != nil {
			http.Error(w, evalErr.Error(), 500)
			return
		}

		log.Println("write bytes to response")
		w.Write(bytes)
	}
}

func MakeRecordHandler(db *Database, tmplEngine *TemplateEngine) handler {
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
