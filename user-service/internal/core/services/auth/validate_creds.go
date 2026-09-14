package auth

import "market/shared/validator"

func ValidateCreds(email, password string, passwordRequirements validator.PasswordRequirements) bool {
	return validator.IsValidEmail(email) && validator.ValidatePassword(password, passwordRequirements).Valid
}
