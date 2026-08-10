package domain

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleUser, RoleAdmin:
		return true
	}
	return false
}
