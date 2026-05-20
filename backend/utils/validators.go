package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	validCities       = map[string]bool{"almaty": true, "astana": true, "shymkent": true, "global": true}
	usernameRegex     = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)
	emailRegex        = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
)

func ValidateCity(city string) error {
	city = strings.ToLower(strings.TrimSpace(city))
	if !validCities[city] {
		return fmt.Errorf("utils.ValidateCity: city must be one of: almaty, astana, shymkent")
	}
	return nil
}

func ValidateUsername(username string) error {
	username = strings.TrimSpace(username)
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("utils.ValidateUsername: username must be 3-32 characters, alphanumeric and underscore only")
	}
	return nil
}

func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("utils.ValidateEmail: email format is invalid")
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 || len(password) > 128 {
		return fmt.Errorf("utils.ValidatePassword: password must be 8-128 characters")
	}
	return nil
}
