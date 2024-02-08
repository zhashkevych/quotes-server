package server

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
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

	listener     net.Listener
	shutdownChan chan struct{}
	connections  map[net.Conn]struct{}
	connMutex    sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc

	// Metrics
	totalRequestsHandled int
	totalResponseTime    time.Duration
	metricsMutex         sync.Mutex
}

func NewTCPServer(port, powDifficulty int, quotesService Quoter, powManager ProofOfWorkManager) *TCPServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &TCPServer{
		port:          port,
		powDifficulty: powDifficulty,
		quotesService: quotesService,
		powManager:    powManager,
		shutdownChan:  make(chan struct{}),
		connections:   make(map[net.Conn]struct{}),
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (s *TCPServer) ListenAndServe() error {
	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}

	defer s.listener.Close()

	go func() {
		<-s.ctx.Done()
		s.listener.Close()
		log.Info("TCP server is shutting down")
	}()

	log.Infof("Starting TCP server at :%d", s.port)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return nil
			default:
				log.Error("Error accepting:", err.Error())
			}
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *TCPServer) handleConnection(conn net.Conn) {
	startTime := time.Now()

	// store current connection for graceful shutdown logic
	s.connMutex.Lock()
	s.connections[conn] = struct{}{}
	s.connMutex.Unlock()

	defer func() {
		s.connMutex.Lock()
		delete(s.connections, conn)
		s.connMutex.Unlock()
		conn.Close()
	}()

	// process request
	log.Infof("received request from %s", conn.RemoteAddr().String())

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

	log.Infof("received solution:  %d", nonce)

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

	endTime := time.Now()

	// collect metrics
	s.metricsMutex.Lock()
	s.totalRequestsHandled++
	s.totalResponseTime += endTime.Sub(startTime)
	s.metricsMutex.Unlock()
}

func (s *TCPServer) Shutdown() {
	s.cancel()

	s.connMutex.Lock()
	for conn := range s.connections {
		conn.Close()
		delete(s.connections, conn)
	}
	s.connMutex.Unlock()

	s.logMetrics()
}

func (s *TCPServer) logMetrics() {
	s.metricsMutex.Lock()

	averageResponseTime := time.Duration(0)
	if s.totalRequestsHandled > 0 {
		averageResponseTime = s.totalResponseTime / time.Duration(s.totalRequestsHandled)
	}

	log.Infof("Total requests handled: %d", s.totalRequestsHandled)
	log.Infof("Average response time: %s", averageResponseTime)

	s.metricsMutex.Unlock()
}
