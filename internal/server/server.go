package server

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

//go:generate mockgen -source=server.go -destination=mocks/mock.go

const (
	IncorrectSolutionResonse    = "Incorrect solution. Try again."
	InternalServerErrorResponse = "Internal server error"
)

type Quoter interface {
	GetRandomQuote() string
}

type ProofOfWorkManager interface {
	GenerateChallenge(difficulty int) (string, error)
	SolveChallenge(challenge string) (int, error)
	VerifySolution(challenge string, nonce int) (bool, error)
}

type TCPServer struct {
	port          int
	powDifficulty int
	quotesService Quoter
	powManager    ProofOfWorkManager

	listener net.Listener
}

func NewTCPServer(port, powDifficulty int, quotesService Quoter, powManger ProofOfWorkManager) *TCPServer {
	return &TCPServer{port, powDifficulty, quotesService, powManger, nil}
}

func (s *TCPServer) ListenAndServe() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}

	s.listener = listener

	defer listener.Close()

	fmt.Printf("starting tcp server at :%d\n", s.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *TCPServer) handleConnection(conn net.Conn) {
	fmt.Printf("received request from %s\n", conn.RemoteAddr().String())

	defer conn.Close()
	reader := bufio.NewReader(conn)

	challenge, err := s.powManager.GenerateChallenge(s.powDifficulty)
	if err != nil {
		fmt.Fprintf(conn, "%s\n", InternalServerErrorResponse)
		return
	}
	fmt.Fprintf(conn, "%s\n", challenge)

	response, _ := reader.ReadString('\n')
	nonce, err := strconv.Atoi(strings.TrimSpace(response))
	if err != nil {
		fmt.Fprintf(conn, "%s\n", IncorrectSolutionResonse)
		return
	}

	fmt.Printf("received solution:  %d\n", nonce)

	isValid, err := s.powManager.VerifySolution(challenge, nonce)
	if err != nil {
		fmt.Fprintf(conn, "%s\n", IncorrectSolutionResonse)
		return
	}

	if !isValid {
		fmt.Fprintf(conn, "%s\n", IncorrectSolutionResonse)
		return
	}

	fmt.Fprintf(conn, "%s\n", s.quotesService.GetRandomQuote())
}

// used for testing
func (s *TCPServer) getAddr() string {
	return s.listener.Addr().String()
}
