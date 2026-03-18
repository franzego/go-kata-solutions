package main

import (
	"context"
	"errors"
	"fmt"
)

type MockMetadataService struct {
	Timeout bool
}

type MetadataError struct {
	Username string
	Url      string
	Err      error
}

func (m *MetadataError) Error() string {
	return fmt.Sprintf("metadata failed on url: %s, with error: %v", m.Url, m.Err)
}

func (m *MetadataError) Unwrap() error {
	return m.Err
}

func (m *MetadataError) Timeout() bool {
	return errors.Is(m.Err, context.DeadlineExceeded)
}

// Authenticate: The service itself
func (m *MockMetadataService) Metadata() error {
	if m.Timeout {
		return fmt.Errorf("metadata: %w", context.DeadlineExceeded)
	}
	return &MetadataError{
		Username: "Meee",
		Url:      "www.yuit.com",
		Err:      context.DeadlineExceeded,
	}
}
