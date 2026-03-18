package main

import (
	"context"
	"errors"
	"fmt"
)

type MockStorageService struct {
	Timeout bool
}

type StorageQuotaError struct {
	UserID string
	Size   string
	Err    error
}

func (s *StorageQuotaError) Error() string {
	return fmt.Sprintf("storage failed for user: %s, err: %v", s.UserID, s.Err)
}

func (s *StorageQuotaError) Unwrap() error {
	return s.Err
}

func (a *StorageQuotaError) Timeout() bool {
	return errors.Is(a.Err, context.DeadlineExceeded)
}

func (s *MockStorageService) StorageService() error {
	if s.Timeout {
		return fmt.Errorf("storage: %w", context.DeadlineExceeded)
	}
	return &StorageQuotaError{
		UserID: "me",
		Size:   "24kb",
		Err:    context.DeadlineExceeded,
	}
}
