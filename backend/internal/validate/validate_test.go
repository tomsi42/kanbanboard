package validate

import "testing"

func TestPassword_valid(t *testing.T) {
	valid := []string{
		"Ab1!@#$%",
		"P4ssw0rd!",
		"Hello1World!",
		"Str0ng&Pass",
		"MyP@ssw0rd",
	}
	for _, p := range valid {
		if msg := Password(p); msg != "" {
			t.Errorf("Password(%q) = %q, want empty (valid)", p, msg)
		}
	}
}

func TestPassword_tooShort(t *testing.T) {
	if msg := Password("Ab1!"); msg == "" {
		t.Error("Password(\"Ab1!\") should be invalid (too short)")
	}
}

func TestPassword_noUppercase(t *testing.T) {
	if msg := Password("abc123!!"); msg == "" {
		t.Error("Password(\"abc123!!\") should be invalid (no uppercase)")
	}
}

func TestPassword_noLowercase(t *testing.T) {
	if msg := Password("ABC123!!"); msg == "" {
		t.Error("Password(\"ABC123!!\") should be invalid (no lowercase)")
	}
}

func TestPassword_noNumber(t *testing.T) {
	if msg := Password("Abcdefg!"); msg == "" {
		t.Error("Password(\"Abcdefg!\") should be invalid (no number)")
	}
}

func TestPassword_noSpecial(t *testing.T) {
	if msg := Password("Abcdef12"); msg == "" {
		t.Error("Password(\"Abcdef12\") should be invalid (no special character)")
	}
}

func TestPassword_empty(t *testing.T) {
	if msg := Password(""); msg == "" {
		t.Error("Password(\"\") should be invalid")
	}
}

// --- Priority ---

func TestPriority_valid(t *testing.T) {
	valid := []string{"none", "low", "medium", "high", "critical"}
	for _, p := range valid {
		if msg := Priority(p); msg != "" {
			t.Errorf("Priority(%q) = %q, want empty (valid)", p, msg)
		}
	}
}

func TestPriority_invalid(t *testing.T) {
	invalid := []string{"", "urgent", "HIGH", "CRITICAL", "1"}
	for _, p := range invalid {
		if msg := Priority(p); msg == "" {
			t.Errorf("Priority(%q) should be invalid", p)
		}
	}
}

// --- ProjectTag ---

func TestProjectTag_valid(t *testing.T) {
	valid := []string{"KB", "MKB", "PROJ"}
	for _, tag := range valid {
		if msg := ProjectTag(tag); msg != "" {
			t.Errorf("ProjectTag(%q) = %q, want empty (valid)", tag, msg)
		}
	}
}

func TestProjectTag_tooShort(t *testing.T) {
	if msg := ProjectTag("K"); msg == "" {
		t.Error("single character tag should be invalid")
	}
}

func TestProjectTag_tooLong(t *testing.T) {
	if msg := ProjectTag("ABCDE"); msg == "" {
		t.Error("5-character tag should be invalid")
	}
}

func TestProjectTag_lowercase(t *testing.T) {
	if msg := ProjectTag("kb"); msg == "" {
		t.Error("lowercase tag should be invalid")
	}
}

func TestProjectTag_numbers(t *testing.T) {
	if msg := ProjectTag("K1"); msg == "" {
		t.Error("tag with numbers should be invalid")
	}
}

func TestProjectTag_empty(t *testing.T) {
	if msg := ProjectTag(""); msg == "" {
		t.Error("empty tag should be invalid")
	}
}
