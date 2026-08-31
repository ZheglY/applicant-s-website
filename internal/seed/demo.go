package seed

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/yarik/unik/internal/domain"
	"github.com/yarik/unik/internal/repository"
	"github.com/yarik/unik/internal/service"
)

const (
	StudentCount    = 300
	StudentPassword = "student123"
)

func Run(ctx context.Context, repo *repository.Repository, authService *service.AuthService, logger *zap.Logger) error {
	if err := repo.ResetApplicationData(ctx); err != nil {
		return err
	}
	if err := authService.Bootstrap(ctx); err != nil {
		return err
	}
	directions, err := repo.ListDirections(ctx)
	if err != nil {
		return err
	}
	byCode := make(map[string]domain.Direction, len(directions))
	for _, direction := range directions {
		byCode[direction.DirectionCode] = direction
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(StudentPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash demo password: %w", err)
	}
	rng := rand.New(rand.NewPCG(20260530, 12))
	for index := 1; index <= StudentCount; index++ {
		name := firstNames[rng.IntN(len(firstNames))]
		first, middle := name[0], name[1]
		last := lastNames[rng.IntN(len(lastNames))]
		if index == 1 {
			first, middle, last = "Ярослав", "Олегович", "Смирнов"
		}
		codes := directionGroups[rng.IntN(len(directionGroups))]
		count := 1 + rng.IntN(min(3, len(codes)))
		codes = append([]string(nil), codes...)
		rng.Shuffle(len(codes), func(i, j int) { codes[i], codes[j] = codes[j], codes[i] })
		selected := make([]domain.Direction, 0, count)
		for _, code := range codes[:count] {
			selected = append(selected, byCode[code])
		}

		subjectSet := make(map[string]struct{})
		for _, direction := range selected {
			for _, subject := range direction.Subjects {
				subjectSet[subject] = struct{}{}
			}
		}
		scores := make(map[string]int, len(subjectSet))
		for subject := range subjectSet {
			scores[subject] = 55 + rng.IntN(46)
		}
		achievements := []domain.Achievement(nil)
		if rng.IntN(100) < 42 {
			achievements = []domain.Achievement{achievementCatalog[rng.IntN(len(achievementCatalog))]}
		}
		total := 0
		for _, direction := range selected {
			score := 0
			for _, subject := range direction.Subjects {
				score += scores[subject]
			}
			if score > total {
				total = score
			}
		}
		for _, achievement := range achievements {
			total += achievement.Points
		}
		birthDate := time.Date(2006+rng.IntN(3), time.Month(1+rng.IntN(12)), 1+rng.IntN(28), 0, 0, 0, 0, time.UTC)
		email := fmt.Sprintf("student%03d@demo.unik", index)
		applicant, createErr := repo.RegisterApplicant(ctx, repository.RegisterApplicantParams{
			LastName: last, FirstName: first, MiddleName: middle, Email: email,
			PasswordHash: string(hash), Phone: fmt.Sprintf("+7 (900) %03d-%02d-%02d", index, index%100, (index*7)%100),
			Telegram: fmt.Sprintf("@unik_student_%03d", index), BirthDate: birthDate.Format("2006-01-02"),
			TotalScore: total, Achievements: achievements, School: fmt.Sprintf("Школа № %d", 100+index%80),
			Region: regions[rng.IntN(len(regions))], Sex: rng.IntN(2) == 0,
			EGEScores: scores, Directions: selected,
		})
		if createErr != nil {
			return fmt.Errorf("seed student %d: %w", index, createErr)
		}
		accepted := false
		for _, direction := range selected {
			status := domain.StatusPending
			roll := rng.IntN(100)
			if roll < 18 && !accepted {
				status, accepted = domain.StatusAccepted, true
			} else if roll >= 86 {
				status = domain.StatusRejected
			}
			if status != domain.StatusPending {
				if err = repo.SetPriorityStatus(ctx, applicant.ID, direction.ID, status); err != nil {
					return err
				}
			}
		}
		if index%50 == 0 {
			logger.Info("demo applicants created", zap.Int("count", index))
		}
	}

	admin, err := repo.GetFirstApplicantByRole(ctx, domain.RoleAdmissions)
	if err != nil {
		return err
	}
	news := []domain.News{
		{Title: "Старт приёмной кампании 2026", Subtitle: "Открыт приём заявлений на программы бакалавриата", Text: "Подать заявление можно онлайн. Проверьте контактные данные и расставьте направления в порядке приоритета."},
		{Title: "Опубликован график консультаций", Subtitle: "Приёмная комиссия отвечает на вопросы абитуриентов", Text: "Консультации проходят по будням. Подробности можно уточнить у сотрудников приёмной комиссии."},
		{Title: "Проверьте выбранные направления", Subtitle: "До завершения кампании можно уточнить приоритеты", Text: "Убедитесь, что выбранные направления расположены в желаемом порядке и баллы ЕГЭ указаны верно."},
	}
	for _, item := range news {
		item.AuthorID = &admin.ID
		if _, err = repo.CreateNews(ctx, item); err != nil {
			return err
		}
	}
	logger.Info("demo data created", zap.Int("students", StudentCount), zap.Int("news", len(news)))
	return nil
}

var firstNames = [][2]string{
	{"Ярослав", "Олегович"}, {"Александр", "Сергеевич"}, {"Дмитрий", "Андреевич"},
	{"Максим", "Ильич"}, {"Михаил", "Алексеевич"}, {"Иван", "Романович"},
	{"Анна", "Сергеевна"}, {"Мария", "Алексеевна"}, {"София", "Андреевна"},
	{"Екатерина", "Ильинична"}, {"Полина", "Дмитриевна"}, {"Дарья", "Павловна"},
}

var lastNames = []string{
	"Смирнов", "Иванов", "Кузнецов", "Соколов", "Попов", "Лебедев", "Козлов", "Новиков",
	"Морозов", "Волков", "Соловьёв", "Васильев", "Зайцев", "Павлов", "Орлов", "Макаров",
}

var regions = []string{
	"Москва", "Московская область", "Санкт-Петербург", "Республика Татарстан",
	"Нижегородская область", "Свердловская область", "Новосибирская область", "Краснодарский край",
}

var directionGroups = [][]string{
	{"ПИ", "ИИ", "ПМ", "ИБ", "РОБ"}, {"ПИ", "ВЕБ", "СА"}, {"БА", "ЭК"}, {"СА", "ИБ", "ПИ"}, {"ДИЗ"},
}

var achievementCatalog = []domain.Achievement{
	{Text: "Золотая медаль", Points: 5}, {Text: "Серебряная медаль", Points: 3},
	{Text: "Призёр региональной олимпиады", Points: 4}, {Text: "Значок ГТО", Points: 2},
	{Text: "Волонтёрская деятельность", Points: 1},
}

func SummaryLine() string {
	values := []string{fmt.Sprintf("students=%d", StudentCount), "staff=2", "news=3"}
	sort.Strings(values)
	return strings.Join(values, " ")
}
