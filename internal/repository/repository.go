package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yarik/unik/internal/domain"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrDuplicate = errors.New("duplicate")
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r *Repository) GetApplicantByID(ctx context.Context, id int64) (domain.Applicant, error) {
	return scanApplicant(r.pool.QueryRow(ctx, applicantSelect+" WHERE id = $1", id))
}

func (r *Repository) GetApplicantByLogin(ctx context.Context, login string) (domain.Applicant, error) {
	return scanApplicant(r.pool.QueryRow(ctx, applicantSelect+" WHERE login = $1 OR email = $1 LIMIT 1", strings.ToLower(login)))
}

func (r *Repository) GetFirstApplicantByRole(ctx context.Context, role string) (domain.Applicant, error) {
	return scanApplicant(r.pool.QueryRow(ctx, applicantSelect+" WHERE role = $1 ORDER BY id LIMIT 1", role))
}

func (r *Repository) ListDirections(ctx context.Context) ([]domain.Direction, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, faculty_name, direction_code, direction_name,
		       budget_places, paid_places, is_full, subjects
		FROM faculty_directions ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list directions: %w", err)
	}
	defer rows.Close()

	directions := make([]domain.Direction, 0)
	for rows.Next() {
		direction, scanErr := scanDirection(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		directions = append(directions, direction)
	}
	return directions, rows.Err()
}

type RegisterApplicantParams struct {
	LastName     string
	FirstName    string
	MiddleName   string
	Email        string
	PasswordHash string
	Phone        string
	Telegram     string
	BirthDate    string
	TotalScore   int
	Achievements []domain.Achievement
	School       string
	Region       string
	Sex          bool
	EGEScores    map[string]int
	Directions   []domain.Direction
}

func (r *Repository) RegisterApplicant(ctx context.Context, params RegisterApplicantParams) (domain.Applicant, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return domain.Applicant{}, fmt.Errorf("begin registration: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	achievements, _ := json.Marshal(params.Achievements)
	scores, _ := json.Marshal(params.EGEScores)
	var id int64
	err = tx.QueryRow(ctx, `
		INSERT INTO applicants (
			last_name, first_name, middle_name, email, login, password_hash,
			phone, telegram, birth_date, total_score, role, sex,
			achievements, school, region, ege_scores
		) VALUES ($1, $2, NULLIF($3, ''), $4, $4, $5, NULLIF($6, ''), NULLIF($7, ''),
		          $8::date, $9, 'student', $10, $11, NULLIF($12, ''), NULLIF($13, ''), $14)
		RETURNING id`,
		params.LastName, params.FirstName, params.MiddleName, params.Email,
		params.PasswordHash, params.Phone, params.Telegram, params.BirthDate,
		params.TotalScore, params.Sex, nullableJSON(params.Achievements, achievements), params.School, params.Region,
		nullableJSON(params.EGEScores, scores),
	).Scan(&id)
	if err != nil {
		return domain.Applicant{}, classifyError("insert applicant", err)
	}

	for index, direction := range params.Directions {
		if _, err = tx.Exec(ctx, `
			INSERT INTO applicant_priorities(applicant_id, direction_id, priority, status)
			VALUES ($1, $2, $3, 'pending')`, id, direction.ID, index+1); err != nil {
			return domain.Applicant{}, classifyError("insert applicant priority", err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Applicant{}, classifyError("commit registration", err)
	}
	return r.GetApplicantByID(ctx, id)
}

func (r *Repository) EnsureStaffUser(
	ctx context.Context,
	login, passwordHash, role, lastName, firstName, middleName string,
) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO applicants (
			last_name, first_name, middle_name, email, login, password_hash,
			role, total_score, sex
		) VALUES ($1, $2, $3, $4, $4, $5, $6, 0, TRUE)
		ON CONFLICT (login) DO UPDATE SET
			last_name = EXCLUDED.last_name,
			first_name = EXCLUDED.first_name,
			middle_name = EXCLUDED.middle_name,
			password_hash = EXCLUDED.password_hash,
			role = EXCLUDED.role,
			updated_at = NOW()`,
		lastName, firstName, middleName, strings.ToLower(login), passwordHash, role,
	)
	if err != nil {
		return classifyError("ensure staff user", err)
	}
	return nil
}

func (r *Repository) SeedDirections(ctx context.Context, directions []domain.Direction) error {
	batch := &pgx.Batch{}
	for _, d := range directions {
		subjects, _ := json.Marshal(d.Subjects)
		batch.Queue(`
			INSERT INTO faculty_directions (
				faculty_name, direction_code, direction_name, budget_places, paid_places, subjects
			) VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (direction_code) DO UPDATE SET
				faculty_name = EXCLUDED.faculty_name,
				direction_name = EXCLUDED.direction_name,
				budget_places = EXCLUDED.budget_places,
				paid_places = EXCLUDED.paid_places,
				subjects = EXCLUDED.subjects,
				updated_at = NOW()`,
			d.FacultyName, d.DirectionCode, d.DirectionName, d.BudgetPlaces, d.PaidPlaces, subjects,
		)
	}
	results := r.pool.SendBatch(ctx, batch)
	defer results.Close()
	for range directions {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("seed directions: %w", err)
		}
	}
	return results.Close()
}

func (r *Repository) ListNews(ctx context.Context) ([]domain.News, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, title, subtitle, text, COALESCE(image_url, ''), author_id, created_at
		FROM news ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list news: %w", err)
	}
	defer rows.Close()
	items := make([]domain.News, 0)
	for rows.Next() {
		var item domain.News
		if err = rows.Scan(&item.ID, &item.Title, &item.Subtitle, &item.Text, &item.ImageURL, &item.AuthorID, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan news: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateNews(ctx context.Context, item domain.News) (domain.News, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO news(title, subtitle, text, image_url, author_id)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5)
		RETURNING id, created_at`,
		item.Title, item.Subtitle, item.Text, item.ImageURL, item.AuthorID,
	).Scan(&item.ID, &item.CreatedAt)
	if err != nil {
		return domain.News{}, classifyError("create news", err)
	}
	return item, nil
}

func (r *Repository) DeleteNews(ctx context.Context, id int64) error {
	result, err := r.pool.Exec(ctx, "DELETE FROM news WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete news: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type RankedApplicant struct {
	Applicant domain.Applicant
	Priority  int
	Status    string
}

type DirectionRanking struct {
	Direction  domain.Direction
	Applicants []RankedApplicant
}

func (r *Repository) ListRankings(ctx context.Context) ([]DirectionRanking, error) {
	directions, err := r.ListDirections(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]DirectionRanking, len(directions))
	byID := make(map[int64]int, len(directions))
	for i, direction := range directions {
		result[i] = DirectionRanking{Direction: direction, Applicants: make([]RankedApplicant, 0)}
		byID[direction.ID] = i
	}

	rows, err := r.pool.Query(ctx, `
		SELECT p.direction_id, p.priority, p.status,
		       a.id, a.last_name, a.first_name, COALESCE(a.middle_name, ''),
		       a.email, a.login, a.password_hash, COALESCE(a.phone, ''),
		       COALESCE(a.telegram, ''), a.birth_date, a.total_score, a.role, a.sex,
		       a.achievements, COALESCE(a.school, ''), COALESCE(a.region, ''), a.ege_scores
		FROM applicant_priorities p
		JOIN applicants a ON a.id = p.applicant_id
		ORDER BY p.direction_id, a.id`)
	if err != nil {
		return nil, fmt.Errorf("list rankings: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var directionID int64
		var priority int
		var status string
		applicant, scanErr := scanApplicantWithPrefix(rows, &directionID, &priority, &status)
		if scanErr != nil {
			return nil, scanErr
		}
		index, ok := byID[directionID]
		if ok {
			result[index].Applicants = append(result[index].Applicants, RankedApplicant{
				Applicant: applicant,
				Priority:  priority,
				Status:    status,
			})
		}
	}
	return result, rows.Err()
}

func (r *Repository) GetApplicantWithPriorities(ctx context.Context, id int64) (domain.Applicant, error) {
	applicant, err := r.GetApplicantByID(ctx, id)
	if err != nil {
		return domain.Applicant{}, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.applicant_id, p.direction_id, p.priority, p.status,
		       d.id, d.faculty_name, d.direction_code, d.direction_name,
		       d.budget_places, d.paid_places, d.is_full, d.subjects
		FROM applicant_priorities p
		JOIN faculty_directions d ON d.id = p.direction_id
		WHERE p.applicant_id = $1
		ORDER BY p.priority`, id)
	if err != nil {
		return domain.Applicant{}, fmt.Errorf("list applicant priorities: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var priority domain.ApplicantPriority
		var subjects []byte
		if err = rows.Scan(
			&priority.ID, &priority.ApplicantID, &priority.DirectionID, &priority.Priority, &priority.Status,
			&priority.Direction.ID, &priority.Direction.FacultyName, &priority.Direction.DirectionCode,
			&priority.Direction.DirectionName, &priority.Direction.BudgetPlaces, &priority.Direction.PaidPlaces,
			&priority.Direction.IsFull, &subjects,
		); err != nil {
			return domain.Applicant{}, fmt.Errorf("scan applicant priority: %w", err)
		}
		if err = json.Unmarshal(subjects, &priority.Direction.Subjects); err != nil {
			return domain.Applicant{}, fmt.Errorf("decode direction subjects: %w", err)
		}
		applicant.Priorities = append(applicant.Priorities, priority)
	}
	return applicant, rows.Err()
}

func (r *Repository) SetPriorityStatus(ctx context.Context, applicantID, directionID int64, status string) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE applicant_priorities SET status = $1, updated_at = NOW()
		WHERE applicant_id = $2 AND direction_id = $3`, status, applicantID, directionID)
	if err != nil {
		return fmt.Errorf("update priority status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) ListStudentApplicants(ctx context.Context) ([]domain.Applicant, error) {
	rows, err := r.pool.Query(ctx, applicantSelect+" WHERE role = 'student' ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("list applicants: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Applicant, 0)
	for rows.Next() {
		applicant, scanErr := scanApplicant(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, applicant)
	}
	return items, rows.Err()
}

func (r *Repository) Summary(ctx context.Context) (domain.Summary, error) {
	result := domain.Summary{
		PriorityCounts: make(map[int]int),
		StatusCounts:   make(map[string]int),
	}
	if err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM applicants WHERE role = 'student'),
			(SELECT COUNT(*) FROM applicant_priorities),
			COALESCE((SELECT SUM(budget_places) FROM faculty_directions), 0),
			COALESCE((SELECT SUM(paid_places) FROM faculty_directions), 0),
			COALESCE((SELECT AVG(total_score) FROM applicants WHERE role = 'student'), 0)
	`).Scan(&result.TotalApplicants, &result.TotalApplications, &result.BudgetPlaces, &result.PaidPlaces, &result.AverageScore); err != nil {
		return domain.Summary{}, fmt.Errorf("summary totals: %w", err)
	}

	priorityRows, err := r.pool.Query(ctx, `SELECT priority, COUNT(*) FROM applicant_priorities GROUP BY priority`)
	if err != nil {
		return domain.Summary{}, fmt.Errorf("summary priorities: %w", err)
	}
	for priorityRows.Next() {
		var key, value int
		if err = priorityRows.Scan(&key, &value); err != nil {
			priorityRows.Close()
			return domain.Summary{}, err
		}
		result.PriorityCounts[key] = value
	}
	priorityRows.Close()

	statusRows, err := r.pool.Query(ctx, `SELECT status, COUNT(*) FROM applicant_priorities GROUP BY status`)
	if err != nil {
		return domain.Summary{}, fmt.Errorf("summary statuses: %w", err)
	}
	for statusRows.Next() {
		var key string
		var value int
		if err = statusRows.Scan(&key, &value); err != nil {
			statusRows.Close()
			return domain.Summary{}, err
		}
		result.StatusCounts[key] = value
	}
	statusRows.Close()

	rows, err := r.pool.Query(ctx, `
		SELECT d.direction_name, COUNT(p.id), d.budget_places, d.paid_places
		FROM faculty_directions d
		LEFT JOIN applicant_priorities p ON p.direction_id = d.id
		GROUP BY d.id
		ORDER BY COUNT(p.id) DESC, d.id`)
	if err != nil {
		return domain.Summary{}, fmt.Errorf("summary directions: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.PopularDirection
		if err = rows.Scan(&item.Name, &item.Applications, &item.BudgetPlaces, &item.PaidPlaces); err != nil {
			return domain.Summary{}, err
		}
		if item.BudgetPlaces > 0 {
			item.Competition = math.Round(float64(item.Applications)/float64(item.BudgetPlaces)*10) / 10
		}
		places := item.BudgetPlaces + item.PaidPlaces
		if places > 0 {
			item.FillPercent = min(100, int(math.Round(float64(item.Applications)/float64(places)*100)))
		}
		result.PopularDirections = append(result.PopularDirections, item)
	}
	return result, rows.Err()
}

func (r *Repository) ResetApplicationData(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `
		TRUNCATE TABLE news, applicant_priorities, applicants, faculty_directions
		RESTART IDENTITY CASCADE`)
	if err != nil {
		return fmt.Errorf("reset application data: %w", err)
	}
	return nil
}

const applicantSelect = `
	SELECT id, last_name, first_name, COALESCE(middle_name, ''),
	       email, login, password_hash, COALESCE(phone, ''), COALESCE(telegram, ''),
	       birth_date, total_score, role, sex, achievements,
	       COALESCE(school, ''), COALESCE(region, ''), ege_scores
	FROM applicants`

type scanner interface {
	Scan(dest ...any) error
}

func scanApplicant(row scanner) (domain.Applicant, error) {
	var applicant domain.Applicant
	var achievements, scores []byte
	err := row.Scan(
		&applicant.ID, &applicant.LastName, &applicant.FirstName, &applicant.MiddleName,
		&applicant.Email, &applicant.Login, &applicant.PasswordHash, &applicant.Phone,
		&applicant.Telegram, &applicant.BirthDate, &applicant.TotalScore, &applicant.Role,
		&applicant.Sex, &achievements, &applicant.School, &applicant.Region, &scores,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Applicant{}, ErrNotFound
	}
	if err != nil {
		return domain.Applicant{}, fmt.Errorf("scan applicant: %w", err)
	}
	if len(achievements) > 0 {
		if err = json.Unmarshal(achievements, &applicant.Achievements); err != nil {
			return domain.Applicant{}, fmt.Errorf("decode achievements: %w", err)
		}
	}
	if len(scores) > 0 {
		if err = json.Unmarshal(scores, &applicant.EGEScores); err != nil {
			return domain.Applicant{}, fmt.Errorf("decode EGE scores: %w", err)
		}
	}
	return applicant, nil
}

func scanApplicantWithPrefix(row scanner, prefix ...any) (domain.Applicant, error) {
	var applicant domain.Applicant
	var achievements, scores []byte
	dest := append(prefix,
		&applicant.ID, &applicant.LastName, &applicant.FirstName, &applicant.MiddleName,
		&applicant.Email, &applicant.Login, &applicant.PasswordHash, &applicant.Phone,
		&applicant.Telegram, &applicant.BirthDate, &applicant.TotalScore, &applicant.Role,
		&applicant.Sex, &achievements, &applicant.School, &applicant.Region, &scores,
	)
	if err := row.Scan(dest...); err != nil {
		return domain.Applicant{}, fmt.Errorf("scan ranked applicant: %w", err)
	}
	if len(achievements) > 0 {
		_ = json.Unmarshal(achievements, &applicant.Achievements)
	}
	if len(scores) > 0 {
		_ = json.Unmarshal(scores, &applicant.EGEScores)
	}
	return applicant, nil
}

func scanDirection(row scanner) (domain.Direction, error) {
	var direction domain.Direction
	var subjects []byte
	if err := row.Scan(
		&direction.ID, &direction.FacultyName, &direction.DirectionCode,
		&direction.DirectionName, &direction.BudgetPlaces, &direction.PaidPlaces,
		&direction.IsFull, &subjects,
	); err != nil {
		return domain.Direction{}, fmt.Errorf("scan direction: %w", err)
	}
	if err := json.Unmarshal(subjects, &direction.Subjects); err != nil {
		return domain.Direction{}, fmt.Errorf("decode direction subjects: %w", err)
	}
	return direction, nil
}

func nullableJSON(value any, encoded []byte) any {
	switch typed := value.(type) {
	case []domain.Achievement:
		if len(typed) == 0 {
			return nil
		}
	case map[string]int:
		if len(typed) == 0 {
			return nil
		}
	}
	return encoded
}

func classifyError(action string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return fmt.Errorf("%s: %w", action, ErrDuplicate)
	}
	return fmt.Errorf("%s: %w", action, err)
}
