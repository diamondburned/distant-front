package tmplutil

import (
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/phogolabs/parcello"
)

// Templater describes the template information to be constructed.
type Templater struct {
	Includes  map[string]string // name -> path
	Functions template.FuncMap

	tmpl template.Template
	once sync.Once
}

// Register registers a subtemplate.
func (tmpler *Templater) Register(name, path string) *Subtemplate {
	tmpler.Includes[name] = path
	return &Subtemplate{tmpler, name}
}

// Execute executes any subtemplate.
func (tmpler *Templater) Execute(w io.Writer, tmpl string, v interface{}) error {
	tmpler.Preload()
	return tmpler.tmpl.ExecuteTemplate(w, tmpl, v)
}

// Func registers a function.
func (tmpler *Templater) Func(name string, fn interface{}) {
	if _, ok := tmpler.Functions[name]; ok {
		panic("Duplicate function with name " + name)
	}
	tmpler.Functions[name] = fn
}

// Preload preloads the templates once. If the templates are already
// preloaded, then it does nothing.
func (tmpler *Templater) Preload() {
	tmpler.once.Do(func() {
		tmpl := template.New("")
		tmpl = tmpl.Funcs(tmpler.Functions)
		for name, incl := range tmpler.Includes {
			tmpl = template.Must(tmpl.New(name).Parse(readFile(incl)))
		}

		tmpler.tmpl = *tmpl
		tmpler.Includes = nil
	})
}

// Subtemplate describes a subtemplate that belongs to some parent template.
type Subtemplate struct {
	tmpl *Templater
	name string
}

// Execute executes the subtemplate.
func (sub *Subtemplate) Execute(w io.Writer, v interface{}) error {
	sub.tmpl.Preload()
	return sub.tmpl.tmpl.ExecuteTemplate(w, sub.name, v)
}

func readFile(filePath string) string {
	f, err := parcello.Open(filePath)
	if err != nil {
		log.Fatalln("Failed to open file:", err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln("Failed to read file:", err)
	}

	return string(b)
}

// AlwaysFlush is the middleware to always flush after a write.
func AlwaysFlush(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(flushWriter{w, flusher}, r)
	})
}

type flushWriter struct {
	http.ResponseWriter
	flusher http.Flusher
}

func (f flushWriter) Write(b []byte) (int, error) {
	n, err := f.ResponseWriter.Write(b)
	if err != nil {
		return n, err
	}

	f.flusher.Flush()
	return n, nil
}
