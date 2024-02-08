package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	quotes "github.com/zhashkevych/quotes-server/internal/quotes/yml"
	"github.com/zhashkevych/quotes-server/internal/server"
	"github.com/zhashkevych/quotes-server/pkg/hashcash"

	log "github.com/sirupsen/logrus"
)

const (
	defaultListenPort    = 9000
	defaultPowDifficulty = 4
	defaultYMLFilePath   = "./quotes.yml"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	SetLogLevel()
}

func main() {
	listenPort, _ := strconv.Atoi(os.Getenv("LISTEN_PORT"))
	if listenPort == 0 {
		listenPort = defaultListenPort
	}

	powDifficulty, _ := strconv.Atoi(os.Getenv("POW_DIFFICULTY"))
	if powDifficulty == 0 {
		powDifficulty = defaultPowDifficulty
	}

	ymlFilePath := os.Getenv("QUOTES_FILEPATH")
	if ymlFilePath == "" {
		ymlFilePath = defaultYMLFilePath
	}

	quotesService, err := quotes.NewYMLService(ymlFilePath)
	if err != nil {
		log.Fatal(err)
	}

	powManager := hashcash.New()

	srv := server.NewTCPServer(listenPort, powDifficulty, quotesService, powManager)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != nil {
			log.Error("Server stopped with error:", err)
		}
	}()

	<-ctx.Done()
	log.Info("Shutting down server...")

	srv.Shutdown()

	wg.Wait()

	log.Info("Server stopped gracefully")
}

func SetLogLevel() {
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	default:
		log.SetLevel(log.DebugLevel)
	}
}
