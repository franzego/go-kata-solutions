package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"golang.org/x/sync/errgroup"
)

type UserAggregator struct {
	Timeout        time.Duration
	Logger         *slog.Logger
	ProfileService Service
	OrderService   Service
}

type AggregatorFunc func(*UserAggregator) error

func WithTimeout(t time.Duration) AggregatorFunc {
	return func(ua *UserAggregator) error {
		ua.Timeout = t
		return nil
	}
}

func WithLogger(l *slog.Logger) AggregatorFunc {
	return func(ua *UserAggregator) error {
		ua.Logger = l
		return nil
	}
}

func NewUserAggregator(opts ...AggregatorFunc) (*UserAggregator, error) {
	userAgg := &UserAggregator{
		Timeout:        10 * time.Second,
		Logger:         slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
		ProfileService: ProfileService,
		OrderService:   OrderService,
	}
	for _, opt := range opts {
		if err := opt(userAgg); err != nil {
			return nil, err
		}
	}
	return userAgg, nil
}

func (ua *UserAggregator) Aggregate(ctx context.Context) (string, error) {
	var prof string
	var ord string
	ctx, cancel := context.WithTimeout(ctx, ua.Timeout)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		prof, err = ua.ProfileService(ctx)
		return err
	})
	g.Go(func() error {
		var err error
		ord, err = ua.OrderService(ctx)
		return err
	})
	if err := g.Wait(); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s | %s", prof, ord), nil
}
