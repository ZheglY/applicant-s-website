package service

import (
	"encoding/json"
	"testing"
)

func TestValidateRegistration(t *testing.T) {
	valid := RegisterInput{
		FullName: "Иванов Иван", Password: "secret", PasswordConfirm: "secret",
		BirthDate: "2007-01-01", Email: "test@example.com", School: "Школа",
		Priorities: []string{"Программная инженерия"}, Agreement: true,
		EGEScores: ScoreMap{"Математика": 90},
	}
	if err := validateRegistration(valid); err != nil {
		t.Fatalf("valid registration rejected: %v", err)
	}

	invalid := valid
	invalid.PasswordConfirm = "another"
	if err := validateRegistration(invalid); err == nil {
		t.Fatal("expected mismatched passwords to be rejected")
	}
}

func TestParseAchievements(t *testing.T) {
	items := parseAchievements("Медаль; ГТО, Олимпиада\nВолонтёр")
	if len(items) != 4 {
		t.Fatalf("expected 4 achievements, got %d", len(items))
	}
}

func TestRegistrationAcceptsFrontendScoreFormat(t *testing.T) {
	var input RegisterInput
	data := []byte(`{"ege_scores":{"Математика":"90","Информатика":88}}`)
	if err := json.Unmarshal(data, &input); err != nil {
		t.Fatal(err)
	}
	if input.EGEScores["Математика"] != 90 || input.EGEScores["Информатика"] != 88 {
		t.Fatalf("unexpected scores: %#v", input.EGEScores)
	}
}
