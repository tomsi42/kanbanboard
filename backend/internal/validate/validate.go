// Package validate provides input validation rules.
package validate

import "unicode"

// Password checks the password against the password policy:
//   - Minimum 8 characters
//   - At least one uppercase letter
//   - At least one lowercase letter
//   - At least one number
//   - At least one special character
//
// Returns an error message if invalid, or empty string if valid.
func Password(password string) string {
	if len(password) < 8 {
		return "Password must be at least 8 characters"
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasNumber = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return "Password must contain at least one uppercase letter"
	}
	if !hasLower {
		return "Password must contain at least one lowercase letter"
	}
	if !hasNumber {
		return "Password must contain at least one number"
	}
	if !hasSpecial {
		return "Password must contain at least one special character"
	}

	return ""
}

// Priority checks the task priority value.
// Valid values: none, low, medium, high, critical.
//
// Returns an error message if invalid, or empty string if valid.
func Priority(priority string) string {
	switch priority {
	case "none", "low", "medium", "high", "critical":
		return ""
	default:
		return "Priority must be none, low, medium, high, or critical"
	}
}

// ProjectTag checks the project tag:
//   - 2-4 characters
//   - Uppercase letters only (A-Z)
//
// Returns an error message if invalid, or empty string if valid.
func ProjectTag(tag string) string {
	if len(tag) < 2 || len(tag) > 4 {
		return "Tag must be 2-4 characters"
	}
	for _, r := range tag {
		if r < 'A' || r > 'Z' {
			return "Tag must contain only uppercase letters (A-Z)"
		}
	}
	return ""
}
