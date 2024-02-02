package main

import (
	"log"
	"os"
	"strconv"

	quotes "github.com/zhashkevych/quotes-server/internal/quotes/yml"
	"github.com/zhashkevych/quotes-server/internal/server"
	"github.com/zhashkevych/quotes-server/pkg/hashcash"
)

const (
	defaultListenPort    = 9000
	defaultPowDifficulty = 4
	defaultYMLFilePath   = "./quotes.yml"
)

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

	srv := server.NewTCPServer(defaultListenPort, defaultPowDifficulty, quotesService, powManager)

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
