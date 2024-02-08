package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/zhashkevych/quotes-server/pkg/hashcash"
)

const (
	defaultServerURL = "localhost:9000"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	SetLogLevel()
}

func main() {
	url := os.Getenv("SERVER_URL")
	if url == "" {
		url = defaultServerURL
	}
	powManager := hashcash.New()

	conn, err := net.Dial("tcp", url)
	if err != nil {
		log.Error("error connecting to the server:", err)
		return
	}
	defer conn.Close()

	log.Infof("Connected to server at %s", url)

	reader := bufio.NewReader(conn)
	challenge, err := reader.ReadString('\n')
	if err != nil {
		log.Error("error reading challenge from server:", err)
		return
	}
	challenge = strings.TrimSpace(challenge)

	log.Infof("Challenge received: %s", challenge)

	nonce, err := powManager.SolveChallenge(challenge)
	if err != nil {
		log.Error("error solving challenge from server:", err)
		return
	}
	fmt.Fprintf(conn, "%d\n", nonce)

	quote, err := reader.ReadString('\n')
	if err != nil {
		log.Error("error reading quote from server:", err)
		return
	}
	log.Infof("Quote received: %s", quote)
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
