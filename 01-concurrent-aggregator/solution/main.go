package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"
)

func main() {
	log.Print("starting aggregator exercise")
	ctx := context.Background()
	agg, _ := NewUserAggregator(WithTimeout(1*time.Second), WithLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))))
	ctx, cancel := context.WithTimeout(ctx, agg.Timeout)
	defer cancel()
	result, err := agg.Aggregate(ctx)
	if err != nil {
		agg.Logger.Error("an error has occured", "error", err)
		os.Exit(1)
	}
	fmt.Println(result)
}
