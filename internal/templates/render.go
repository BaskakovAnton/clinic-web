package templates

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"
	"time"

	"clinic/internal/store"
)

type Renderer struct {
	t *template.Template
}

func doctorAvatarURL(d store.Doctor) string {
	return avatarFrom(d.ImageURL, d.Gender)
}

func staffAvatarURL(st store.Staff) string {
	return avatarFrom(st.ImageURL, st.Gender)
}

func avatarFrom(imageURL string, gender sql.NullString) string {
	if strings.TrimSpace(imageURL) != "" {
		return imageURL
	}
	if gender.Valid {
		switch gender.String {
		case "f":
			return "/static/img/doctor-placeholder-f.jpg"
		case "m":
			return "/static/img/doctor-placeholder-m.jpg"
		}
	}
	return "/static/img/doctor-placeholder.jpg"
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
		"doctorAvatar": doctorAvatarURL,
		"staffAvatar":  staffAvatarURL,
		"orderKind": func(kind string) string {
			switch kind {
			case "medication":
				return "лекарство"
			case "procedure":
				return "процедура"
			case "test":
				return "анализ"
			default:
				return kind
			}
		},
		"requestStatus": func(status string) string {
			switch status {
			case "new":
				return "новая"
			case "called":
				return "перезвонили"
			case "done":
				return "закрыта"
			case "cancelled":
				return "отменена"
			default:
				return status
			}
		},
		"priceFmt": func(v any) string {
			switch t := v.(type) {
			case int64:
				return fmt.Sprintf("%d", t)
			case int:
				return fmt.Sprintf("%d", t)
			default:
				return ""
			}
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
