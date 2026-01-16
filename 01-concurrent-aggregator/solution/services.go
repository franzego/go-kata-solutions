package main

import (
	"context"
	"time"
)

type Service func(ctx context.Context) (string, error)

func ProfileService(ctx context.Context) (string, error) {
	var err error
	// time.Sleep(2 * time.Second)
	// return "", fmt.Errorf("profile service error")
	return "", err
}

func OrderService(ctx context.Context) (string, error) {
	select {
	case <-time.After(10 * time.Second):
		return "Order: 5", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
