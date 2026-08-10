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

import "net/mail"

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
