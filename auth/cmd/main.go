package main

import (
	"context"
	"flag"
	"log"

	"github.com/MercerMorning/go_example/auth/internal/app"
	"github.com/MercerMorning/go_example/auth/internal/logger"
)

var logLevel = flag.String("l", "info", "log level")

func main() {
	flag.Parse()

	ctx := context.Background()
	a, err := app.NewApp(ctx)
	if err != nil {
		log.Fatalf("failed to init app: %s", err.Error())
	}

	logger.Info("gnom")

	err = a.Run()
	if err != nil {
		log.Fatalf("failed to run app: %s", err.Error())
	}
}
