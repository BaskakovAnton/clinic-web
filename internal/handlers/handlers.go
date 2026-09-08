package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"clinic/internal/auth"
	apptemplates "clinic/internal/templates"
	"clinic/internal/store"
)

type Server struct {
	Auth  *auth.Manager
	Store *store.Store
	Tmpl  *apptemplates.Renderer
	DB    *sql.DB
}

type pageData struct {
	Title         string
	Active        string
	Session       auth.Session
	Flash         string
	FlashError    string
	Query         string
	Date          string
	DoctorID      int
	Patients      []store.Patient
	Patient       store.Patient
	Slots         []store.FreeSlot
	Appointments  []store.AppointmentRow
	Appointment   store.AppointmentRow
	Specialties   []string
	Specialty     string
	CardRows      []store.CardRow
	Doctors       []store.Doctor
	StaffList     []store.Staff
	Staff         store.Staff
	Visit         store.Visit
	Orders        []store.VisitOrder
	Workload      []store.WorkloadRow
	Preferential  []store.PreferentialRow
	Stats         store.DashboardStats
	Form          map[string]string
}

func (s *Server) render(w http.ResponseWriter, name string, data pageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.Tmpl.Render(w, name, data); err != nil {
		log.Printf("template %s: %v", name, err)
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}

func urlQuery(s string) string { return url.QueryEscape(s) }

func flashOK(r *http.Request) string  { return r.URL.Query().Get("ok") }
func flashErr(r *http.Request) string { return r.URL.Query().Get("error") }

func redirectOK(w http.ResponseWriter, r *http.Request, path, msg string) {
	http.Redirect(w, r, path+"?ok="+urlQuery(msg), http.StatusSeeOther)
}

func redirectErr(w http.ResponseWriter, r *http.Request, path, msg string) {
	http.Redirect(w, r, path+"?error="+urlQuery(msg), http.StatusSeeOther)
}

func parseDate(s string) (time.Time, error) {
	if s == "" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	}
	return time.ParseInLocation("2006-01-02", s, time.Local)
}

func doctorCookie(r *http.Request) int {
	c, err := r.Cookie("clinic_doctor_id")
	if err != nil {
		return 0
	}
	id, _ := strconv.Atoi(c.Value)
	return id
}

func setDoctorCookie(w http.ResponseWriter, id int) {
	secure := os.Getenv("APP_ENV") == "production" || os.Getenv("SECURE_COOKIE") == "1" || os.Getenv("SECURE_COOKIE") == "true"
	http.SetCookie(w, &http.Cookie{
		Name:     "clinic_doctor_id",
		Value:    strconv.Itoa(id),
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Server) Healthz(w http.ResponseWriter, r *http.Request) {
	if err := s.DB.PingContext(r.Context()); err != nil {
		http.Error(w, "db down", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) Home(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	http.Redirect(w, r, auth.HomePath(sess.Role), http.StatusSeeOther)
}

func (s *Server) LoginGet(w http.ResponseWriter, r *http.Request) {
	if sess, err := s.Auth.SessionFromRequest(r); err == nil {
		http.Redirect(w, r, auth.HomePath(sess.Role), http.StatusSeeOther)
		return
	}
	s.render(w, "login.html", pageData{Title: "Вход", FlashError: flashErr(r)})
}

func (s *Server) LoginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		redirectErr(w, r, "/login", "Ошибка формы")
		return
	}
	sess, err := s.Auth.VerifyLogin(r.Context(), strings.TrimSpace(r.FormValue("login")), r.FormValue("password"))
	if err != nil {
		redirectErr(w, r, "/login", err.Error())
		return
	}
	if err := s.Auth.SetSession(w, sess); err != nil {
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, auth.HomePath(sess.Role), http.StatusSeeOther)
}

func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	s.Auth.ClearSession(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
