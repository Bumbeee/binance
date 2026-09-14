package validator

const (
	ErrMsgPasswordMinLength      = "password must be at least %d characters long"
	ErrMsgPasswordMaxLength      = "password must be at most %d characters long"
	ErrMsgPasswordRequireUpper   = "password must contain at least one uppercase letter"
	ErrMsgPasswordRequireLower   = "password must contain at least one lowercase letter"
	ErrMsgPasswordRequireDigit   = "password must contain at least one digit"
	ErrMsgPasswordRequireSpecial = "password must contain at least one special character"
	ErrMsgPasswordDisallowSpaces = "password must not contain spaces"
)
