package templates

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"time"
)

type Renderer struct {
	t *template.Template
}

func New(dir string) (*Renderer, error) {
	funcs := template.FuncMap{
		"roleLabel": func(role string) string {
			switch role {
			case "admin":
				return "админ"
			case "registrar":
				return "регистратура"
			case "doctor":
				return "врач"
			default:
				return role
			}
		},
		"shiftDate": func(date string, days int) string {
			t, err := time.Parse("2006-01-02", date)
			if err != nil {
				return date
			}
			return t.AddDate(0, 0, days).Format("2006-01-02")
		},
	}
	pattern := filepath.Join(dir, "*.html")
	t, err := template.New("").Funcs(funcs).ParseGlob(pattern)
	if err != nil {
		return nil, fmt.Errorf("parse templates %s: %w", pattern, err)
	}
	return &Renderer{t: t}, nil
}

func (r *Renderer) Render(w io.Writer, name string, data any) error {
	return r.t.ExecuteTemplate(w, name, data)
}
