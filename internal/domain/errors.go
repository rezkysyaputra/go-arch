package domain

import "errors"

var (
	ErrInvalidInput   = errors.New("invalid input")
	ErrUserNotFound   = errors.New("user not found")
	ErrEmailTaken     = errors.New("email already taken")
	ErrPublishFailed  = errors.New("event publish failed")
	ErrDuplicateEmail = errors.New("duplicate email")
)
