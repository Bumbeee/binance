package validator

import (
	"fmt"
	"unicode"
)

// PasswordRequirements описывает набор правил для проверки пароля
type PasswordRequirements struct {
	MinLength      int
	MaxLength      int // 0 = без ограничения
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
	DisallowSpaces bool
}

// ValidationResult — результат проверки
type ValidationResult struct {
	Valid  bool
	Errors []string
}

func (r *ValidationResult) addError(msg string) {
	r.Valid = false
	r.Errors = append(r.Errors, msg)
}

// ValidatePassword проверяет пароль по заданным требованиям
func ValidatePassword(password string, req PasswordRequirements) ValidationResult {
	result := ValidationResult{Valid: true}

	if req.MinLength > 0 && len(password) < req.MinLength {
		result.addError(fmt.Sprintf("пароль должен быть не короче %d символов", req.MinLength))
	}

	if req.MaxLength > 0 && len(password) > req.MaxLength {
		result.addError(fmt.Sprintf("пароль должен быть не длиннее %d символов", req.MaxLength))
	}

	var hasUpper, hasLower, hasDigit, hasSpecial, hasSpace bool

	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		case unicode.IsSpace(ch):
			hasSpace = true
		}
	}

	if req.RequireUpper && !hasUpper {
		result.addError("пароль должен содержать хотя бы одну заглавную букву")
	}
	if req.RequireLower && !hasLower {
		result.addError("пароль должен содержать хотя бы одну строчную букву")
	}
	if req.RequireDigit && !hasDigit {
		result.addError("пароль должен содержать хотя бы одну цифру")
	}
	if req.RequireSpecial && !hasSpecial {
		result.addError("пароль должен содержать хотя бы один спецсимвол")
	}
	if req.DisallowSpaces && hasSpace {
		result.addError("пароль не должен содержать пробелы")
	}

	return result
}
