package validator

func ValidateCreds(email, password string, passwordRequirements PasswordRequirements) bool {
	return IsValidEmail(email) || ValidatePassword(password, passwordRequirements).Valid
}
