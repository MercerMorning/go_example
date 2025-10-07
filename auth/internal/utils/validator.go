package utils

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidPassword = errors.New("password must be at least 8 characters long")
	ErrEmptyName       = errors.New("name cannot be empty")
	ErrEmptyEmail      = errors.New("email cannot be empty")
	ErrEmptyPassword   = errors.New("password cannot be empty")
	ErrPasswordMismatch = errors.New("passwords do not match")
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateUserInfo валидирует данные пользователя при создании
func ValidateUserInfo(name, email, password, passwordConfirm string) error {
	if strings.TrimSpace(name) == "" {
		return ErrEmptyName
	}

	if strings.TrimSpace(email) == "" {
		return ErrEmptyEmail
	}

	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	if strings.TrimSpace(password) == "" {
		return ErrEmptyPassword
	}

	if len(password) < 8 {
		return ErrInvalidPassword
	}

	if password != passwordConfirm {
		return ErrPasswordMismatch
	}

	return nil
}

// ValidateUserUpdate валидирует данные при обновлении пользователя
func ValidateUserUpdate(name, email *string) error {
	if name != nil && strings.TrimSpace(*name) == "" {
		return ErrEmptyName
	}

	if email != nil {
		if strings.TrimSpace(*email) == "" {
			return ErrEmptyEmail
		}
		if !emailRegex.MatchString(*email) {
			return ErrInvalidEmail
		}
	}

	return nil
}
