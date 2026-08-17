// package validator

// import (
// 	"net"
// 	"net/mail"
// 	"strings"
// )

// func IsValidEmail(email string) bool {
// 	addr, err := mail.ParseAddress(email)
// 	if err != nil {
// 		return false
// 	}

// 	parts := strings.Split(addr.Address, "@")
// 	if len(parts) != 2 {
// 		return false
// 	}
// 	domain := parts[1]

// 	mxRecords, err := net.LookupMX(domain)
// 	return err == nil && len(mxRecords) > 0
// }

package validator

import (
	"net/mail"
	"regexp"
	"strings"
)

// emailRegex — быстрый предварительный фильтр по структуре email.
// Не пропускает "display name", лишние пробелы, управляющие символы и т.п.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+$`)

// NormalizeEmail убирает пробелы по краям и управляющие символы.
func NormalizeEmail(email string) string {
	email = strings.TrimSpace(email)
	email = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, email)
	return email
}

// IsValidEmail проверяет email в два этапа:
// 1) regex — отсекает мусор, "display name", лишние пробелы/символы;
// 2) mail.ParseAddress — финальная проверка на соответствие RFC 5322.
func IsValidEmail(email string) bool {
	email = NormalizeEmail(email)

	if email == "" || len(email) > 254 {
		return false
	}

	if !emailRegex.MatchString(email) {
		return false
	}

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	// защита от случаев вида "John <john@example.com>",
	// которые regex уже отсёк, но проверяем ещё раз на всякий случай:
	// адрес после парсинга должен совпадать с исходной строкой.
	return addr.Address == email
}
