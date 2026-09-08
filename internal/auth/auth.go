package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Role string

const (
	RoleAdmin     Role = "admin"
	RoleRegistrar Role = "registrar"
	RoleDoctor    Role = "doctor"
)

type Session struct {
	Login string `json:"login"`
	Role  Role   `json:"role"`
	Exp   int64  `json:"exp"`
}

type Manager struct {
	secret []byte
	dbURL  string
	secure bool
}

func NewManager(secret, databaseURL string, secureCookie bool) *Manager {
	return &Manager{secret: []byte(secret), dbURL: databaseURL, secure: secureCookie}
}

func RoleFromLogin(login string) (Role, bool) {
	switch login {
	case "u_admin":
		return RoleAdmin, true
	case "u_registrar":
		return RoleRegistrar, true
	case "u_doctor":
		return RoleDoctor, true
	default:
		return "", false
	}
}

func (m *Manager) VerifyLogin(ctx context.Context, login, password string) (Session, error) {
	role, ok := RoleFromLogin(login)
	if !ok {
		return Session{}, errors.New("неизвестный логин")
	}
	u, err := url.Parse(m.dbURL)
	if err != nil {
		return Session{}, fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	u.User = url.UserPassword(login, password)
	db, err := sql.Open("pgx", u.String())
	if err != nil {
		return Session{}, err
	}
	defer db.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		return Session{}, errors.New("неверный логин или пароль")
	}
	return Session{
		Login: login,
		Role:  role,
		Exp:   time.Now().Add(12 * time.Hour).Unix(),
	}, nil
}

func (m *Manager) SetSession(w http.ResponseWriter, s Session) error {
	raw, err := json.Marshal(s)
	if err != nil {
		return err
	}
	sig := m.sign(raw)
	val := base64.RawURLEncoding.EncodeToString(raw) + "." + base64.RawURLEncoding.EncodeToString(sig)
	http.SetCookie(w, &http.Cookie{
		Name:     "clinic_session",
		Value:    val,
		Path:     "/",
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(s.Exp, 0),
	})
	return nil
}

func (m *Manager) ClearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "clinic_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   m.secure,
		MaxAge:   -1,
	})
}

func (m *Manager) SessionFromRequest(r *http.Request) (Session, error) {
	c, err := r.Cookie("clinic_session")
	if err != nil {
		return Session{}, err
	}
	parts := strings.Split(c.Value, ".")
	if len(parts) != 2 {
		return Session{}, errors.New("bad cookie")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Session{}, err
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Session{}, err
	}
	if !hmac.Equal(sig, m.sign(raw)) {
		return Session{}, errors.New("bad signature")
	}
	var s Session
	if err := json.Unmarshal(raw, &s); err != nil {
		return Session{}, err
	}
	if time.Now().Unix() > s.Exp {
		return Session{}, errors.New("expired")
	}
	if _, ok := RoleFromLogin(s.Login); !ok || s.Role == "" {
		return Session{}, errors.New("invalid session")
	}
	return s, nil
}

func (m *Manager) sign(raw []byte) []byte {
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write(raw)
	return mac.Sum(nil)
}

type ctxKey int

const sessionKey ctxKey = 1

func WithSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, sessionKey, s)
}

func SessionFromContext(ctx context.Context) (Session, bool) {
	s, ok := ctx.Value(sessionKey).(Session)
	return s, ok
}

func (m *Manager) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s, err := m.SessionFromRequest(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r.WithContext(WithSession(r.Context(), s)))
	})
}

func RequireRoles(roles ...Role) func(http.Handler) http.Handler {
	allowed := make(map[Role]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s, ok := SessionFromContext(r.Context())
			if !ok {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			if _, ok := allowed[s.Role]; !ok {
				http.Error(w, "Недостаточно прав", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func HomePath(role Role) string {
	switch role {
	case RoleAdmin:
		return "/admin"
	case RoleRegistrar:
		return "/registrar/patients"
	case RoleDoctor:
		return "/doctor/day"
	default:
		return "/login"
	}
}
