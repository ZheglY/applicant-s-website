package httpserver

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/yarik/unik/internal/cache"
	"github.com/yarik/unik/internal/service"
)

type registerPageData struct {
	Directions any
	Subjects   []string
}

func (s *Server) handleRegisterPage(w http.ResponseWriter, r *http.Request) {
	directions, err := s.cachedDirections(r.Context())
	if err != nil {
		s.internalError(w, "list registration directions", err)
		return
	}
	set := make(map[string]struct{})
	for _, direction := range directions {
		for _, subject := range direction.Subjects {
			set[subject] = struct{}{}
		}
	}
	subjects := make([]string, 0, len(set))
	for subject := range set {
		subjects = append(subjects, subject)
	}
	sort.Strings(subjects)
	s.render(w, http.StatusOK, "regist.html", registerPageData{Directions: directions, Subjects: subjects})
}

func (s *Server) handleLoginPage(w http.ResponseWriter, _ *http.Request) {
	s.render(w, http.StatusOK, "enter.html", nil)
}

func (s *Server) handleLoginRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/auth/enter", http.StatusFound)
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if !s.enforceRateLimit(w, r, "register", clientIP(r), s.cfg.RegistrationRateLimit, s.cfg.RegistrationRateWindow) {
		return
	}
	var input service.RegisterInput
	if !decodeJSON(w, r, &input) {
		return
	}
	applicant, err := s.authService.Register(r.Context(), input)
	var validationErr *service.ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeError(w, http.StatusBadRequest, validationErr.Error())
	case errors.Is(err, service.ErrEmailExists):
		writeError(w, http.StatusBadRequest, "User with this email already exists")
	case err != nil:
		s.internalError(w, "register user", err)
	default:
		s.invalidateCache(r.Context(), cache.KeySummary)
		writeJSON(w, http.StatusOK, map[string]any{
			"message": "Регистрация успешна", "id": applicant.ID, "email": applicant.Email,
		})
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var input service.LoginInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if !s.enforceRateLimit(w, r, "login", clientIP(r)+":"+strings.ToLower(input.Username), s.cfg.LoginRateLimit, s.cfg.LoginRateWindow) {
		return
	}
	user, token, err := s.authService.Login(r.Context(), input)
	if errors.Is(err, service.ErrInvalidCredentials) {
		writeError(w, http.StatusUnauthorized, "Неверный логин или пароль")
		return
	}
	if err != nil {
		s.internalError(w, "authenticate user", err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: accessCookieName, Value: token, Path: "/", HttpOnly: true,
		Secure: s.cfg.CookieSecure, SameSite: http.SameSiteLaxMode,
		MaxAge: int(s.cfg.AccessTokenTTL.Seconds()), Expires: time.Now().Add(s.cfg.AccessTokenTTL),
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"message": "Вход выполнен", "role": user.Role, "user_id": user.ID, "full_name": user.FullName(),
	})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name: accessCookieName, Value: "", Path: "/", HttpOnly: true,
		Secure: s.cfg.CookieSecure, SameSite: http.SameSiteLaxMode,
		MaxAge: -1, Expires: time.Unix(1, 0),
	})
	http.Redirect(w, r, "/auth/enter", http.StatusFound)
}

func (s *Server) internalError(w http.ResponseWriter, action string, err error) {
	s.logger.Error(action, zap.Error(err))
	writeError(w, http.StatusInternalServerError, "Internal server error")
}
