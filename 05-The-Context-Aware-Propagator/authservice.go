package main

import (
	"context"
	"errors"
	"fmt"
)

type MockAuthService struct {
	ForceTimeout     bool
	ForceWrongApiKey string
}

type AuthError struct {
	ID     string
	Secret string // it must be redacted.
	Err    error
}

func (a *AuthError) Error() string {
	return fmt.Sprintf("auth failed for user %s (secret redacted): %v", a.ID, a.Err)
}

func (a *AuthError) Unwrap() error {
	return a.Err
}

// Timeout is tracking Timeout errors
func (a *AuthError) Timeout() bool {
	return errors.Is(a.Err, context.DeadlineExceeded)
}

// Authenticate: The service itself
func (m *MockAuthService) Authenticate() error {
	if m.ForceTimeout {
		return fmt.Errorf("auth: %w", context.DeadlineExceeded)
	}

	return &AuthError{
		ID:     "user",
		Secret: "today",
		Err:    context.DeadlineExceeded,
	}
}
