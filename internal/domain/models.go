package domain

import "time"

const (
	RoleStudent    = "student"
	RoleAdmissions = "admissions"
	RoleAnalyst    = "analyst"
	RoleGuest      = "guest"

	StatusAccepted = "accepted"
	StatusPending  = "pending"
	StatusRejected = "rejected"
)

type Applicant struct {
	ID           int64
	LastName     string
	FirstName    string
	MiddleName   string
	Email        string
	Login        string
	PasswordHash string
	Phone        string
	Telegram     string
	BirthDate    *time.Time
	TotalScore   int
	Role         string
	Sex          bool
	Achievements []Achievement
	School       string
	Region       string
	EGEScores    map[string]int
	Priorities   []ApplicantPriority
}

func (a Applicant) FullName() string {
	name := a.LastName + " " + a.FirstName
	if a.MiddleName != "" {
		name += " " + a.MiddleName
	}
	return name
}

type Achievement struct {
	Text   string `json:"text"`
	Points int    `json:"points"`
}

type Direction struct {
	ID            int64
	FacultyName   string
	DirectionCode string
	DirectionName string
	BudgetPlaces  int
	PaidPlaces    int
	IsFull        bool
	Subjects      []string
}

type ApplicantPriority struct {
	ID          int64
	ApplicantID int64
	DirectionID int64
	Priority    int
	Status      string
	Direction   Direction
}

type News struct {
	ID        int64
	Title     string
	Subtitle  string
	Text      string
	ImageURL  string
	AuthorID  *int64
	CreatedAt time.Time
}

type PopularDirection struct {
	Name         string
	Applications int
	BudgetPlaces int
	PaidPlaces   int
	Competition  float64
	FillPercent  int
}

type Summary struct {
	TotalApplicants   int
	TotalApplications int
	BudgetPlaces      int
	PaidPlaces        int
	AverageScore      float64
	PriorityCounts    map[int]int
	StatusCounts      map[string]int
	PopularDirections []PopularDirection
}
