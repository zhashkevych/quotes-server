package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zhashkevych/quotes-server/pkg/hashcash"
)

const (
	defaultServerURL = "localhost:9000"
)

var (
	totalRequestsSent int
	errorCount        int
	totalResponseTime time.Duration
	metricsMutex      sync.Mutex
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var wg sync.WaitGroup

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				wg.Add(1)
				go func() {
					defer wg.Done()
					sendRequest(ctx, url, powManager)
				}()
			}
		}
	}()

	<-shutdownChan
	log.Info("Shutdown signal received; stopping new requests...")

	cancel()
	ticker.Stop()
	wg.Wait()

	logMetrics()
}

func sendRequest(ctx context.Context, url string, powManager *hashcash.Hashcash) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	startTime := time.Now()

	conn, err := net.Dial("tcp", url)
	if err != nil {
		incrementErrorCount()
		log.Error("Failed to connect to the server:", err)
		return
	}
	defer conn.Close()

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	log.Debugf("Connected to server at %s", url)
	incrementRequestsCount()

	reader := bufio.NewReader(conn)
	challenge, err := reader.ReadString('\n')
	if err != nil {
		incrementErrorCount()
		log.Error("Failed to read challenge from server:", err)
		return
	}
	challenge = strings.TrimSpace(challenge)

	log.Debugf("Challenge received: %s", challenge)

	nonce, err := powManager.SolveChallenge(challenge)
	if err != nil {
		incrementErrorCount()
		log.Error("Failed to solve challenge from server:", err)
		return
	}
	fmt.Fprintf(conn, "%d\n", nonce)

	quote, err := reader.ReadString('\n')
	if err != nil {
		incrementErrorCount()
		log.Error("Failed to read quote from server:", err)
		return
	}
	log.Infof("Quote received: %s", quote)

	endTime := time.Now()

	collectResponseTimeMetric(startTime, endTime)
}

func incrementErrorCount() {
	metricsMutex.Lock()
	errorCount++
	metricsMutex.Unlock()
}

func incrementRequestsCount() {
	metricsMutex.Lock()
	totalRequestsSent++
	metricsMutex.Unlock()
}

func collectResponseTimeMetric(startTime, endTime time.Time) {
	metricsMutex.Lock()
	totalResponseTime += endTime.Sub(startTime)
	metricsMutex.Unlock()
}

func logMetrics() {
	metricsMutex.Lock()
	defer metricsMutex.Unlock()

	averageResponseTime := time.Duration(0)
	if totalRequestsSent > 0 {
		averageResponseTime = totalResponseTime / time.Duration(totalRequestsSent)
	}

	log.Infof("Total requests sent: %d", totalRequestsSent)
	log.Infof("Error count: %d", errorCount)
	log.Infof("Average response time: %s", averageResponseTime)
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
