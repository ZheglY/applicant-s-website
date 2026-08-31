package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"github.com/yarik/unik/internal/auth"
	"github.com/yarik/unik/internal/config"
	"github.com/yarik/unik/internal/domain"
	"github.com/yarik/unik/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already exists")
)

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

type AuthService struct {
	repo   *repository.Repository
	tokens *auth.Manager
	cfg    config.Config
}

func NewAuthService(repo *repository.Repository, tokens *auth.Manager, cfg config.Config) *AuthService {
	return &AuthService{repo: repo, tokens: tokens, cfg: cfg}
}

type RegisterInput struct {
	FullName        string   `json:"fullname"`
	Password        string   `json:"password"`
	PasswordConfirm string   `json:"password_confirm"`
	BirthDate       string   `json:"birthdate"`
	Phone           string   `json:"phone"`
	Email           string   `json:"email"`
	Telegram        string   `json:"telegram"`
	School          string   `json:"school"`
	Achievements    string   `json:"achievements"`
	Priorities      []string `json:"priorities"`
	Agreement       bool     `json:"agreement"`
	EGEScores       ScoreMap `json:"ege_scores"`
	EGEScoresAlias  ScoreMap `json:"egeScores,omitempty"`
}

type ScoreMap map[string]int

func (scores *ScoreMap) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	result := make(ScoreMap, len(raw))
	for subject, value := range raw {
		var score int
		if err := json.Unmarshal(value, &score); err != nil {
			var text string
			if stringErr := json.Unmarshal(value, &text); stringErr != nil || strings.TrimSpace(text) == "" {
				return fmt.Errorf("invalid EGE score for %s", subject)
			}
			var scanErr error
			score, scanErr = strconv.Atoi(strings.TrimSpace(text))
			if scanErr != nil {
				return fmt.Errorf("invalid EGE score for %s", subject)
			}
		}
		result[strings.TrimSpace(subject)] = score
	}
	*scores = result
	return nil
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (domain.Applicant, error) {
	if len(input.EGEScores) == 0 && len(input.EGEScoresAlias) > 0 {
		input.EGEScores = input.EGEScoresAlias
	}
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.FullName = strings.TrimSpace(input.FullName)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Telegram = strings.TrimSpace(input.Telegram)
	input.School = strings.TrimSpace(input.School)
	if err := validateRegistration(input); err != nil {
		return domain.Applicant{}, err
	}

	directions, err := s.repo.ListDirections(ctx)
	if err != nil {
		return domain.Applicant{}, err
	}
	byName := make(map[string]domain.Direction, len(directions))
	for _, direction := range directions {
		byName[direction.DirectionName] = direction
	}
	selected := make([]domain.Direction, 0, len(input.Priorities))
	maxScore := 0
	for _, name := range input.Priorities {
		direction, ok := byName[strings.TrimSpace(name)]
		if !ok {
			return domain.Applicant{}, &ValidationError{Message: "Direction not found: " + name}
		}
		total := 0
		missing := make([]string, 0)
		for _, subject := range direction.Subjects {
			score, exists := input.EGEScores[subject]
			if !exists {
				missing = append(missing, subject)
				continue
			}
			total += score
		}
		if len(missing) > 0 {
			return domain.Applicant{}, &ValidationError{
				Message: fmt.Sprintf("Missing EGE scores for %s: %s", direction.DirectionName, strings.Join(missing, ", ")),
			}
		}
		if total > maxScore {
			maxScore = total
		}
		selected = append(selected, direction)
	}

	achievements := parseAchievements(input.Achievements)
	for _, item := range achievements {
		maxScore += item.Points
	}
	parts := strings.Fields(input.FullName)
	middleName := ""
	if len(parts) > 2 {
		middleName = strings.Join(parts[2:], " ")
	}
	hash, err := hashPassword(input.Password)
	if err != nil {
		return domain.Applicant{}, fmt.Errorf("hash password: %w", err)
	}

	applicant, err := s.repo.RegisterApplicant(ctx, repository.RegisterApplicantParams{
		LastName:     parts[0],
		FirstName:    parts[1],
		MiddleName:   middleName,
		Email:        input.Email,
		PasswordHash: hash,
		Phone:        input.Phone,
		Telegram:     input.Telegram,
		BirthDate:    input.BirthDate,
		TotalScore:   maxScore,
		Achievements: achievements,
		School:       input.School,
		Sex:          true,
		EGEScores:    map[string]int(input.EGEScores),
		Directions:   selected,
	})
	if errors.Is(err, repository.ErrDuplicate) {
		return domain.Applicant{}, ErrEmailExists
	}
	return applicant, err
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (domain.Applicant, string, error) {
	user, err := s.repo.GetApplicantByLogin(ctx, strings.TrimSpace(input.Username))
	if errors.Is(err, repository.ErrNotFound) {
		return domain.Applicant{}, "", ErrInvalidCredentials
	}
	if err != nil {
		return domain.Applicant{}, "", err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), passwordBytes(input.Password)) != nil {
		return domain.Applicant{}, "", ErrInvalidCredentials
	}
	token, err := s.tokens.Create(user.ID, user.Role)
	if err != nil {
		return domain.Applicant{}, "", fmt.Errorf("create access token: %w", err)
	}
	return user, token, nil
}

func (s *AuthService) Bootstrap(ctx context.Context) error {
	if err := s.repo.SeedDirections(ctx, DefaultDirections()); err != nil {
		return err
	}
	staff := []struct {
		login, password, role, last, first, middle string
	}{
		{s.cfg.DefaultAdmissionsLogin, s.cfg.DefaultAdmissionsPassword, domain.RoleAdmissions, "Петрова", "Ирина", "Сергеевна"},
		{s.cfg.DefaultAnalystLogin, s.cfg.DefaultAnalystPassword, domain.RoleAnalyst, "Соколов", "Алексей", "Викторович"},
	}
	for _, item := range staff {
		hash, err := s.staffPasswordHash(ctx, item.login, item.password)
		if err != nil {
			return err
		}
		if err = s.repo.EnsureStaffUser(ctx, item.login, hash, item.role, item.last, item.first, item.middle); err != nil {
			return err
		}
	}
	return nil
}

func (s *AuthService) staffPasswordHash(ctx context.Context, login, password string) (string, error) {
	existing, err := s.repo.GetApplicantByLogin(ctx, login)
	if err == nil && bcrypt.CompareHashAndPassword([]byte(existing.PasswordHash), passwordBytes(password)) == nil {
		return existing.PasswordHash, nil
	}
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return "", err
	}
	return hashPassword(password)
}

func validateRegistration(input RegisterInput) error {
	if len(strings.Fields(input.FullName)) < 2 {
		return &ValidationError{Message: "Full name is required"}
	}
	if utf8.RuneCountInString(input.Password) < 6 {
		return &ValidationError{Message: "Password must be at least 6 characters"}
	}
	if input.Password != input.PasswordConfirm {
		return &ValidationError{Message: "Passwords do not match"}
	}
	if !emailPattern.MatchString(input.Email) {
		return &ValidationError{Message: "Invalid email format"}
	}
	if input.BirthDate == "" || !datePattern.MatchString(input.BirthDate) {
		return &ValidationError{Message: "Invalid birth date"}
	}
	if _, err := time.Parse("2006-01-02", input.BirthDate); err != nil {
		return &ValidationError{Message: "Invalid birth date"}
	}
	if strings.TrimSpace(input.School) == "" {
		return &ValidationError{Message: "School is required"}
	}
	if !input.Agreement {
		return &ValidationError{Message: "Agreement is required"}
	}
	if len(input.Priorities) == 0 || len(input.Priorities) > 3 {
		return &ValidationError{Message: "Select from one to three directions"}
	}
	seen := make(map[string]struct{}, len(input.Priorities))
	for _, name := range input.Priorities {
		name = strings.TrimSpace(name)
		if name == "" {
			return &ValidationError{Message: "Direction is required"}
		}
		if _, exists := seen[name]; exists {
			return &ValidationError{Message: "Directions must not repeat"}
		}
		seen[name] = struct{}{}
	}
	if len(input.EGEScores) == 0 {
		return &ValidationError{Message: "EGE scores are required"}
	}
	for subject, score := range input.EGEScores {
		if strings.TrimSpace(subject) == "" || score < 0 || score > 100 {
			return &ValidationError{Message: "EGE score must be between 0 and 100"}
		}
	}
	return nil
}

func parseAchievements(value string) []domain.Achievement {
	parts := strings.FieldsFunc(value, func(r rune) bool { return r == ';' || r == '\n' || r == ',' })
	result := make([]domain.Achievement, 0, len(parts))
	for _, part := range parts {
		if text := strings.TrimSpace(part); text != "" {
			result = append(result, domain.Achievement{Text: text, Points: 1})
		}
	}
	return result
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(passwordBytes(password), bcrypt.DefaultCost)
	return string(hash), err
}

func passwordBytes(password string) []byte {
	value := []byte(password)
	if len(value) > 72 {
		value = value[:72]
	}
	return value
}

var (
	emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	datePattern  = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

func DefaultDirections() []domain.Direction {
	items := []domain.Direction{
		{FacultyName: "Университет", DirectionCode: "ПИ", DirectionName: "Программная инженерия", BudgetPlaces: 12, PaidPlaces: 12, Subjects: []string{"Математика", "Информатика", "Русский язык"}},
		{FacultyName: "Университет", DirectionCode: "ИИ", DirectionName: "Искусственный интеллект", BudgetPlaces: 15, PaidPlaces: 16, Subjects: []string{"Математика", "Информатика", "Русский язык"}},
		{FacultyName: "Университет", DirectionCode: "БА", DirectionName: "Бизнес-аналитика", BudgetPlaces: 8, PaidPlaces: 10, Subjects: []string{"Математика", "Обществознание", "Русский язык"}},
		{FacultyName: "Университет", DirectionCode: "ПМ", DirectionName: "Прикладная математика", BudgetPlaces: 10, PaidPlaces: 12, Subjects: []string{"Математика", "Физика", "Русский язык"}},
		{FacultyName: "Университет", DirectionCode: "ИБ", DirectionName: "Информационная безопасность", BudgetPlaces: 13, PaidPlaces: 14, Subjects: []string{"Математика", "Информатика", "Русский язык"}},
		{FacultyName: "Университет", DirectionCode: "ВЕБ", DirectionName: "Web-разработка", BudgetPlaces: 9, PaidPlaces: 10, Subjects: []string{"Математика", "Информатика", "Русский язык"}},
		{FacultyName: "Университет", DirectionCode: "СА", DirectionName: "Системное администрирование", BudgetPlaces: 7, PaidPlaces: 9, Subjects: []string{"Математика", "Информатика", "Русский язык"}},
		{FacultyName: "Университет", DirectionCode: "РОБ", DirectionName: "Робототехника", BudgetPlaces: 6, PaidPlaces: 8, Subjects: []string{"Математика", "Физика", "Русский язык"}},
		{FacultyName: "Университет", DirectionCode: "ДИЗ", DirectionName: "Дизайн", BudgetPlaces: 5, PaidPlaces: 10, Subjects: []string{"Рисунок", "Композиция", "Русский язык"}},
		{FacultyName: "Университет", DirectionCode: "ЭК", DirectionName: "Экономика", BudgetPlaces: 8, PaidPlaces: 12, Subjects: []string{"Математика", "Обществознание", "Русский язык"}},
	}
	return items
}
