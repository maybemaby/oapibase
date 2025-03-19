package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/maybemaby/oapibase/api"
)

type Args struct {
	Port   string
	DbPath string
}

func argParse() Args {
	var args Args
	flag.StringVar(&args.Port, "port", "8000", "port to listen on")
	flag.StringVar(&args.DbPath, "db", "app.db", "path to sqlite db")
	flag.Parse()

	return args
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	location, err := time.LoadLocation("UTC")
	if err != nil {
		log.Println("Error loading location")
	}

	time.Local = location
}

func main() {
	args := argParse()

	ctx, cancel := context.WithCancel(context.Background())

	loadEnv()

	// Server
	appEnv := os.Getenv("APP_ENV")

	isDebug := appEnv == "development"

	server, err := api.NewServer(!isDebug)

	if err != nil {
		log.Fatalf("Error creating server: %v", err)
		os.Exit(1)
	}

	server.WithPort(args.Port)

	// OS Signals
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	go func() {
		err := server.Start(ctx)

		if err != nil {
			log.Println(fmt.Printf("Error starting server: %v", err))
		}
	}()

	<-osSignals

	cancel()

}
