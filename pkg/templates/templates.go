package templates

import (
	"fmt"
	"html/template"
	"io"
	"path"
	"sync"
	"time"

	"github.com/spraints/temps/pkg/reloadabletemplate"
	"github.com/spraints/temps/pkg/types"
)

type Template interface {
	Execute(wr io.Writer, data interface{}) error
	ExecuteTemplate(wr io.Writer, name string, data interface{}) error
}

type Templates struct {
	load  func(path string) Template
	funcs map[string]interface{}

	lock      sync.Mutex
	templates map[string]Template
}

func New(basePath string, reloadable bool, assetTag string) *Templates {
	return &Templates{
		load:      loader(basePath, reloadable, mkfact(assetTag)),
		templates: map[string]Template{},
	}
}

func mkfact(assetTag string) func(name string) *template.Template {
	tz, err := time.LoadLocation("America/Indiana/Indianapolis")
	if err != nil {
		panic(err)
	}

	funcs := map[string]interface{}{
		"c":  func(t types.Temperature) string { return fmt.Sprintf("%3.0f", t.Celsius()) },
		"f":  func(t types.Temperature) string { return fmt.Sprintf("%3.0f", t.Fahrenheit()) },
		"t":  func(t time.Time) string { return t.In(tz).Format("15:04 2-Jan-2006 MST") },
		"ts": func(t time.Time) string { return fmt.Sprint(t.Unix()) },
		"at": func(path string) string { return path + "?" + assetTag },
	}

	return func(name string) *template.Template {
		return template.New(name).Funcs(funcs)
	}
}

func (t *Templates) Get(path string) Template {
	t.lock.Lock()
	defer t.lock.Unlock()

	if res, ok := t.templates[path]; ok {
		return res
	}

	res := t.load(path)
	t.templates[path] = res
	return res
}

func loader(basePath string, reloadable bool, factory func(name string) *template.Template) func(path string) Template {
	if reloadable {
		return func(name string) Template {
			return reloadabletemplate.New(path.Join(basePath, name), func() *template.Template { return factory(name) })
		}
	}

	return func(name string) Template {
		t, err := factory(name).ParseFiles(path.Join(basePath, name))
		if err != nil {
			return errTemplate{err}
		}
		return t
	}
}

type errTemplate struct {
	err error
}

func (et errTemplate) Execute(wr io.Writer, data interface{}) error {
	return et.err
}

func (et errTemplate) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	return et.err
}
