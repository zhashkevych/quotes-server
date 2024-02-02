package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/zhashkevych/quotes-server/pkg/hashcash"
)

const (
	DefaultServerURL = "localhost:9000"
)

func main() {
	url := os.Getenv("SERVER_URL")
	if url == "" {
		url = DefaultServerURL
	}
	powManager := hashcash.New()

	conn, err := net.Dial("tcp", url)
	if err != nil {
		fmt.Println("Error connecting to the server:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to server at", url)

	reader := bufio.NewReader(conn)
	challenge, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading challenge from server:", err)
		return
	}
	challenge = strings.TrimSpace(challenge)
	fmt.Println("Challenge received:", challenge)

	nonce, err := powManager.SolveChallenge(challenge)
	if err != nil {
		fmt.Println("Error solving challenge from server:", err)
		return
	}
	fmt.Fprintf(conn, "%d\n", nonce)

	quote, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading quote from server:", err)
		return
	}
	fmt.Printf("Quote received: %s\n", quote)
}
