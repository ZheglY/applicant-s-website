package httpserver

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/yarik/unik/internal/cache"
	"github.com/yarik/unik/internal/domain"
	"github.com/yarik/unik/internal/repository"
)

var (
	statusLabels = map[string]string{
		domain.StatusAccepted: "Зачислен",
		domain.StatusPending:  "На рассмотрении",
		domain.StatusRejected: "Отклонён",
	}
	statusClasses = map[string]string{
		domain.StatusAccepted: "status_accepted",
		domain.StatusPending:  "status_pending",
		domain.StatusRejected: "status_rejected",
	}
	profileStatusClasses = map[string]string{
		domain.StatusAccepted: "accepted",
		domain.StatusPending:  "pending",
		domain.StatusRejected: "rejected",
	}
	subjectAliases = map[string][]string{
		"Math":           {"Math", "Mathematics", "Mathematics (profile)", "Математика", "Математика (проф.)"},
		"Russian":        {"Russian", "Русский язык"},
		"Informatics":    {"Informatics", "Информатика"},
		"Physics":        {"Physics", "Физика"},
		"Social Studies": {"Social Studies", "Обществознание"},
		"Web Design":     {"Web Design", "Веб-дизайн", "Дизайн"},
		"Networks":       {"Networks", "Сети"},
		"Linux":          {"Linux"},
		"Drawing":        {"Drawing", "Рисунок"},
		"Composition":    {"Composition", "Композиция"},
	}
)

type newsPageData struct {
	CurrentUser   domain.Applicant
	CanManageNews bool
}

type newsResponse struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Text     string `json:"text"`
	Image    string `json:"image,omitempty"`
	Date     string `json:"date"`
}

func (s *Server) handleNewsPage(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUserPage(w, r)
	if !ok {
		return
	}
	s.render(w, http.StatusOK, "main.html", newsPageData{
		CurrentUser: user, CanManageNews: user.Role == domain.RoleAdmissions,
	})
}

func (s *Server) handleNewsData(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireUserAPI(w, r); !ok {
		return
	}
	items, err := s.cachedNews(r.Context())
	if err != nil {
		s.internalError(w, "list news", err)
		return
	}
	response := make([]newsResponse, 0, len(items))
	for _, item := range items {
		response = append(response, newsToResponse(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": response})
}

type createNewsRequest struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Text     string `json:"text"`
	Image    string `json:"image"`
}

func (s *Server) handleCreateNews(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUserAPI(w, r)
	if !ok {
		return
	}
	if user.Role != domain.RoleAdmissions {
		writeError(w, http.StatusForbidden, "Admissions only")
		return
	}
	var input createNewsRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	input.Title, input.Subtitle, input.Text, input.Image = cleanText(input.Title), cleanText(input.Subtitle), cleanText(input.Text), cleanText(input.Image)
	if input.Title == "" || input.Subtitle == "" || input.Text == "" {
		writeError(w, http.StatusBadRequest, "Field is required")
		return
	}
	created, err := s.repo.CreateNews(r.Context(), domain.News{
		Title: input.Title, Subtitle: input.Subtitle, Text: input.Text, ImageURL: input.Image, AuthorID: &user.ID,
	})
	if err != nil {
		s.internalError(w, "create news", err)
		return
	}
	s.invalidateCache(r.Context(), cache.KeyNews)
	writeJSON(w, http.StatusOK, newsToResponse(created))
}

func (s *Server) handleDeleteNews(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUserAPI(w, r)
	if !ok {
		return
	}
	if user.Role != domain.RoleAdmissions {
		writeError(w, http.StatusForbidden, "Admissions only")
		return
	}
	id, ok := pathID(w, r, "newsID")
	if !ok {
		return
	}
	if err := s.repo.DeleteNews(r.Context(), id); errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "News not found")
	} else if err != nil {
		s.internalError(w, "delete news", err)
	} else {
		s.invalidateCache(r.Context(), cache.KeyNews)
		writeJSON(w, http.StatusOK, map[string]string{"message": "Deleted"})
	}
}

func newsToResponse(item domain.News) newsResponse {
	return newsResponse{
		ID: item.ID, Title: item.Title, Subtitle: item.Subtitle, Text: item.Text,
		Image: item.ImageURL, Date: item.CreatedAt.Format("02.01.2006"),
	}
}

type listPageData struct {
	CurrentUser domain.Applicant
	Directions  []directionView
}

type directionView struct {
	ID           int64
	Name         string
	BudgetPlaces int
	PaidPlaces   int
	Subjects     []string
	Applicants   []rankedApplicantView
}

type rankedApplicantView struct {
	ID          int64
	FullName    string
	Age         string
	Scores      []string
	TotalScore  int
	Status      string
	StatusLabel string
	StatusClass string
}

func (s *Server) handleListPage(w http.ResponseWriter, r *http.Request) {
	user, ok := s.optionalUser(r)
	if !ok {
		user = domain.Applicant{Role: domain.RoleGuest, FirstName: "Гость"}
	}
	rankings, err := s.repo.ListRankings(r.Context())
	if err != nil {
		s.internalError(w, "list rankings", err)
		return
	}
	views := make([]directionView, 0, len(rankings))
	for _, ranking := range rankings {
		view := directionView{
			ID: ranking.Direction.ID, Name: ranking.Direction.DirectionName,
			BudgetPlaces: ranking.Direction.BudgetPlaces, PaidPlaces: ranking.Direction.PaidPlaces,
			Subjects: ranking.Direction.Subjects,
		}
		for _, item := range ranking.Applicants {
			scores := make([]string, 0, len(ranking.Direction.Subjects))
			for _, subject := range ranking.Direction.Subjects {
				if score, exists := scoreForSubject(item.Applicant.EGEScores, subject); exists {
					scores = append(scores, fmt.Sprintf("%d", score))
				} else {
					scores = append(scores, "-")
				}
			}
			status := normalizedStatus(item.Status)
			view.Applicants = append(view.Applicants, rankedApplicantView{
				ID: item.Applicant.ID, FullName: item.Applicant.FullName(), Age: ageText(item.Applicant.BirthDate),
				Scores: scores, TotalScore: competitiveScore(item.Applicant, ranking.Direction.Subjects),
				Status: status, StatusLabel: statusLabels[status], StatusClass: statusClasses[status],
			})
		}
		sort.SliceStable(view.Applicants, func(i, j int) bool {
			if view.Applicants[i].TotalScore == view.Applicants[j].TotalScore {
				return view.Applicants[i].ID < view.Applicants[j].ID
			}
			return view.Applicants[i].TotalScore > view.Applicants[j].TotalScore
		})
		views = append(views, view)
	}
	s.render(w, http.StatusOK, "list.html", listPageData{CurrentUser: user, Directions: views})
}

type statsPageData struct {
	CurrentUser       domain.Applicant
	TotalApplicants   int
	TotalApplications int
	BudgetPlaces      int
	PaidPlaces        int
	AverageScore      float64
	Competition       float64
	PlanTotal         int
	PlanPercent       int
	Priority1         int
	Priority2         int
	Priority3         int
	AcceptedCount     int
	PendingCount      int
	RejectedCount     int
	PopularDirections []domain.PopularDirection
}

func (s *Server) handleStatsPage(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUserPage(w, r)
	if !ok {
		return
	}
	if user.Role != domain.RoleAdmissions && user.Role != domain.RoleAnalyst {
		writeError(w, http.StatusForbidden, "Staff access required")
		return
	}
	summary, err := s.cachedSummary(r.Context())
	if err != nil {
		s.internalError(w, "load statistics", err)
		return
	}
	totalPlaces := summary.BudgetPlaces + summary.PaidPlaces
	competition := 0.0
	if summary.BudgetPlaces > 0 {
		competition = math.Round(float64(summary.TotalApplications)/float64(summary.BudgetPlaces)*10) / 10
	}
	planPercent := 0
	if totalPlaces > 0 {
		planPercent = min(100, int(math.Round(float64(summary.TotalApplications)/float64(totalPlaces)*100)))
	}
	s.render(w, http.StatusOK, "stats.html", statsPageData{
		CurrentUser: user, TotalApplicants: summary.TotalApplicants, TotalApplications: summary.TotalApplications,
		BudgetPlaces: summary.BudgetPlaces, PaidPlaces: summary.PaidPlaces,
		AverageScore: math.Round(summary.AverageScore*10) / 10, Competition: competition,
		PlanTotal: totalPlaces, PlanPercent: planPercent,
		Priority1: summary.PriorityCounts[1], Priority2: summary.PriorityCounts[2], Priority3: summary.PriorityCounts[3],
		AcceptedCount: summary.StatusCounts[domain.StatusAccepted], PendingCount: summary.StatusCounts[domain.StatusPending],
		RejectedCount: summary.StatusCounts[domain.StatusRejected], PopularDirections: summary.PopularDirections,
	})
}

func (s *Server) handleApplicantList(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUserAPI(w, r)
	if !ok {
		return
	}
	if user.Role != domain.RoleAdmissions && user.Role != domain.RoleAnalyst {
		writeError(w, http.StatusForbidden, "Access denied")
		return
	}
	applicants, err := s.repo.ListStudentApplicants(r.Context())
	if err != nil {
		s.internalError(w, "list applicants", err)
		return
	}
	items := make([]map[string]any, 0, len(applicants))
	for _, applicant := range applicants {
		items = append(items, map[string]any{"id": applicant.ID, "full_name": applicant.FullName(), "total_score": applicant.TotalScore})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

type profilePageData struct {
	CurrentUser   domain.Applicant
	Applicant     domain.Applicant
	AgeText       string
	EGEScores     []scoreView
	TotalScore    int
	Achievements  []domain.Achievement
	BonusPoints   int
	Directions    []profileDirectionView
	CanEditStatus bool
}

type scoreView struct {
	Subject string
	Score   int
}

type profileDirectionView struct {
	Priority    int
	Name        string
	Status      string
	StatusLabel string
	StatusClass string
	Score       int
	Position    string
	PlaceType   string
	Reason      string
	DirectionID int64
}

func (s *Server) handleApplicantPage(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUserPage(w, r)
	if !ok {
		return
	}
	studentID, ok := pathID(w, r, "studentID")
	if !ok {
		return
	}
	if user.Role == domain.RoleStudent && user.ID != studentID {
		writeError(w, http.StatusForbidden, "Access denied")
		return
	}
	applicant, err := s.repo.GetApplicantWithPriorities(r.Context(), studentID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Applicant not found")
		return
	}
	if err != nil {
		s.internalError(w, "load applicant", err)
		return
	}
	rankings, err := s.repo.ListRankings(r.Context())
	if err != nil {
		s.internalError(w, "load applicant rankings", err)
		return
	}
	byDirection := make(map[int64]repository.DirectionRanking, len(rankings))
	for _, ranking := range rankings {
		byDirection[ranking.Direction.ID] = ranking
	}

	directions := make([]profileDirectionView, 0, len(applicant.Priorities))
	for _, priority := range applicant.Priorities {
		ranking := byDirection[priority.DirectionID]
		sort.SliceStable(ranking.Applicants, func(i, j int) bool {
			a := competitiveScore(ranking.Applicants[i].Applicant, priority.Direction.Subjects)
			b := competitiveScore(ranking.Applicants[j].Applicant, priority.Direction.Subjects)
			if a == b {
				return ranking.Applicants[i].Applicant.ID < ranking.Applicants[j].Applicant.ID
			}
			return a > b
		})
		position := 1
		for i, item := range ranking.Applicants {
			if item.Applicant.ID == applicant.ID {
				position = i + 1
				break
			}
		}
		status := normalizedStatus(priority.Status)
		view := profileDirectionView{
			Priority: priority.Priority, Name: priority.Direction.DirectionName,
			Status: status, StatusLabel: statusLabels[status], StatusClass: profileStatusClasses[status],
			Score:     competitiveScore(applicant, priority.Direction.Subjects),
			Position:  fmt.Sprintf("%d из %d", position, len(ranking.Applicants)),
			PlaceType: "Платное", DirectionID: priority.DirectionID,
		}
		if position <= priority.Direction.BudgetPlaces {
			view.PlaceType = "Бюджет"
		}
		if status == domain.StatusRejected {
			view.Reason = "Высокий конкурс"
		}
		directions = append(directions, view)
	}

	scores := make([]scoreView, 0, len(applicant.EGEScores))
	for subject, score := range applicant.EGEScores {
		scores = append(scores, scoreView{Subject: subject, Score: score})
	}
	sort.Slice(scores, func(i, j int) bool { return scores[i].Subject < scores[j].Subject })
	bonus := 0
	for _, item := range applicant.Achievements {
		bonus += item.Points
	}
	s.render(w, http.StatusOK, "lk.html", profilePageData{
		CurrentUser: user, Applicant: applicant, AgeText: ageText(applicant.BirthDate),
		EGEScores: scores, TotalScore: applicant.TotalScore, Achievements: applicant.Achievements,
		BonusPoints: bonus, Directions: directions, CanEditStatus: user.Role == domain.RoleAdmissions,
	})
}

type updateStatusRequest struct {
	DirectionID int64  `json:"direction_id"`
	Status      string `json:"status"`
}

func (s *Server) handleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUserAPI(w, r)
	if !ok {
		return
	}
	if user.Role != domain.RoleAdmissions {
		writeError(w, http.StatusForbidden, "Admissions only")
		return
	}
	studentID, ok := pathID(w, r, "studentID")
	if !ok {
		return
	}
	if _, err := s.repo.GetApplicantByID(r.Context(), studentID); errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Applicant not found")
		return
	} else if err != nil {
		s.internalError(w, "load applicant for status update", err)
		return
	}
	var input updateStatusRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	if input.DirectionID <= 0 || (input.Status != domain.StatusAccepted && input.Status != domain.StatusPending && input.Status != domain.StatusRejected) {
		writeError(w, http.StatusBadRequest, "Invalid status update")
		return
	}
	if err := s.repo.SetPriorityStatus(r.Context(), studentID, input.DirectionID, input.Status); errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Priority not found")
	} else if err != nil {
		s.internalError(w, "update applicant status", err)
	} else {
		s.invalidateCache(r.Context(), cache.KeySummary)
		writeJSON(w, http.StatusOK, map[string]string{"message": "Status updated"})
	}
}

func competitiveScore(applicant domain.Applicant, subjects []string) int {
	total := 0
	for _, subject := range subjects {
		if score, ok := scoreForSubject(applicant.EGEScores, subject); ok {
			total += score
		}
	}
	return total
}

func scoreForSubject(scores map[string]int, subject string) (int, bool) {
	aliases := subjectAliases[subject]
	if len(aliases) == 0 {
		aliases = []string{subject}
	}
	for key, value := range scores {
		for _, alias := range aliases {
			if strings.EqualFold(key, alias) {
				return value, true
			}
		}
	}
	return 0, false
}

func normalizedStatus(status string) string {
	if _, ok := statusLabels[status]; ok {
		return status
	}
	return domain.StatusPending
}

func ageText(birthDate *time.Time) string {
	if birthDate == nil {
		return "-"
	}
	now := time.Now()
	age := now.Year() - birthDate.Year()
	anniversary := time.Date(now.Year(), birthDate.Month(), birthDate.Day(), 0, 0, 0, 0, now.Location())
	if now.Before(anniversary) {
		age--
	}
	return fmt.Sprintf("%d", age)
}
