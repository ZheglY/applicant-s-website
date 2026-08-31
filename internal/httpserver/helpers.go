package httpserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/yarik/unik/internal/domain"
	"github.com/yarik/unik/internal/repository"
)

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, detail string) {
	writeJSON(w, status, map[string]string{"detail": detail})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "JSON body must contain a single object")
		return false
	}
	return true
}

func pathID(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "Invalid identifier")
		return 0, false
	}
	return id, true
}

func (s *Server) optionalUser(r *http.Request) (domain.Applicant, bool) {
	cookie, err := r.Cookie(accessCookieName)
	if err != nil || cookie.Value == "" {
		return domain.Applicant{}, false
	}
	claims, err := s.tokens.Verify(cookie.Value)
	if err != nil {
		return domain.Applicant{}, false
	}
	user, err := s.repo.GetApplicantByID(r.Context(), claims.UserID)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			s.logger.Warn("load current user", zap.Error(err))
		}
		return domain.Applicant{}, false
	}
	return user, true
}

func (s *Server) requireUserAPI(w http.ResponseWriter, r *http.Request) (domain.Applicant, bool) {
	user, ok := s.optionalUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required")
	}
	return user, ok
}

func (s *Server) requireUserPage(w http.ResponseWriter, r *http.Request) (domain.Applicant, bool) {
	user, ok := s.optionalUser(r)
	if !ok {
		http.Redirect(w, r, "/auth/enter", http.StatusFound)
	}
	return user, ok
}
