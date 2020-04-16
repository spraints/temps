package reloadabletemplate

import (
	"html/template"
	"io"
	"os"
	"sync"
	"time"
)

type ReloadableTemplate struct {
	path    string
	factory func() *template.Template

	lock        sync.Mutex
	loadedAt    time.Time
	template    *template.Template
	templateErr error
}

func New(path string, factory func() *template.Template) *ReloadableTemplate {
	return &ReloadableTemplate{path: path, factory: factory}
}

func (r *ReloadableTemplate) Execute(wr io.Writer, data interface{}) error {
	t, err := r.load()
	if err != nil {
		return err
	}
	return t.Execute(wr, data)
}

func (r *ReloadableTemplate) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	t, err := r.load()
	if err != nil {
		return err
	}
	return t.ExecuteTemplate(wr, name, data)
}

func (r *ReloadableTemplate) load() (*template.Template, error) {
	var (
		stat os.FileInfo
		err  error
	)

	r.lock.Lock()
	defer r.lock.Unlock()

	if r.loadedAt.IsZero() {
		goto load
	}

	stat, err = os.Stat(r.path)
	if err != nil {
		return nil, err
	}

	if stat.ModTime().Before(r.loadedAt) {
		goto done
	}

load:
	r.loadedAt = time.Now()
	r.template, r.templateErr = r.factory().ParseFiles(r.path)

done:
	return r.template, r.templateErr
}
