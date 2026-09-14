package validator

import (
	"fmt"
	"unicode"
)

type PasswordRequirements struct {
	MinLength      int
	MaxLength      int // 0 = no limit
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
	DisallowSpaces bool
}

type ValidationResult struct {
	Valid  bool
	Errors []string
}

func (r *ValidationResult) addError(msg string) {
	r.Valid = false
	r.Errors = append(r.Errors, msg)
}

func ValidatePassword(password string, req PasswordRequirements) ValidationResult {
	result := ValidationResult{Valid: true}

	if req.MinLength > 0 && len(password) < req.MinLength {
		result.addError(fmt.Sprintf(ErrMsgPasswordMinLength, req.MinLength))
	}

	if req.MaxLength > 0 && len(password) > req.MaxLength {
		result.addError(fmt.Sprintf(ErrMsgPasswordMaxLength, req.MaxLength))
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
		result.addError(ErrMsgPasswordRequireUpper)
	}
	if req.RequireLower && !hasLower {
		result.addError(ErrMsgPasswordRequireLower)
	}
	if req.RequireDigit && !hasDigit {
		result.addError(ErrMsgPasswordRequireDigit)
	}
	if req.RequireSpecial && !hasSpecial {
		result.addError(ErrMsgPasswordRequireSpecial)
	}
	if req.DisallowSpaces && hasSpace {
		result.addError(ErrMsgPasswordDisallowSpaces)
	}

	return result
}
