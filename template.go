package tapedeck

import (
	"fmt"
	"html/template"
	"log"
	"os"
)

// TemplateEngine is built upon [template.Template]. It is used to cache and evaluate templates.
// Call [NewTemplateEngine] to create a new instance and call [TemplateEngine.Init] before first use.
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

func NewTemplateEngine(templateDir string, cache bool) *TemplateEngine {
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
	bw := new(LazyBytesWriter)

	log.Println("template execute")
	err := tmpl.Execute(bw, data)
	if err != nil {
		return nil, fmt.Errorf("template execute error: %w", err)
	}

	return bw.Bytes()
}
