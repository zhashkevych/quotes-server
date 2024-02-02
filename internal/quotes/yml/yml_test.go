package quotes

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewYMLService(t *testing.T) {
	expectedQuotes := []string{"Test quote 1", "Test quote 2"}
	filepath, err := createTempQuotesFile(expectedQuotes)
	assert.NoError(t, err)

	defer os.Remove(filepath)

	service, err := NewYMLService(filepath)
	assert.NoError(t, err)

	assert.Equal(t, len(expectedQuotes), len(service.quotes))
}

func TestGetRandomQuote(t *testing.T) {
	rand.New(rand.NewSource(time.Now().Unix()))

	expectedQuotes := []string{"Test quote 1", "Test quote 2", "Test quote 3"}
	filepath, err := createTempQuotesFile(expectedQuotes)
	assert.NoError(t, err)

	defer os.Remove(filepath)

	service, err := NewYMLService(filepath)
	assert.NoError(t, err)

	for i := 0; i < 10; i++ {
		quote := service.GetRandomQuote()
		if quote == "" {
			t.Error("Expected a non-empty quote")
		}

		assert.NotEmpty(t, quote)
		assert.Equal(t, true, contains(expectedQuotes, quote))
	}
}

func createTempQuotesFile(quotes []string) (string, error) {
	content := "quotes:\n"
	for _, q := range quotes {
		content += "  - \"" + q + "\"\n"
	}
	tmpfile, err := os.CreateTemp("", "quotes*.yml")
	if err != nil {
		return "", err
	}
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		os.Remove(tmpfile.Name())
		return "", err
	}
	return tmpfile.Name(), nil
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
