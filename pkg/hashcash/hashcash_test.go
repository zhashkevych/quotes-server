package hashcash

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateChallenge(t *testing.T) {
	h := &Hashcash{}

	difficulty := 1
	challenge, err := h.GenerateChallenge(difficulty)
	assert.NoError(t, err)

	assert.Contains(t, challenge, ":")
	assert.Contains(t, challenge, strconv.Itoa(difficulty))
}

func TestSolveChallenge(t *testing.T) {
	h := &Hashcash{}

	difficulty := 1
	challenge, _ := h.GenerateChallenge(difficulty)

	nonce, err := h.SolveChallenge(challenge)
	assert.NoError(t, err)
	assert.NotZero(t, nonce)
}

func TestVerifySolution(t *testing.T) {
	h := &Hashcash{}

	difficulty := 1
	challenge, err := h.GenerateChallenge(difficulty)
	assert.NoError(t, err)

	nonce, err := h.SolveChallenge(challenge)
	assert.NoError(t, err)

	isValid, err := h.VerifySolution(challenge, nonce)
	assert.NoError(t, err)
	assert.NotEqual(t, isValid, false)
}

func TestVerifySolutionInvalidFormat(t *testing.T) {
	h := &Hashcash{}

	challenge := "invalid-format"
	nonce := 0

	_, err := h.VerifySolution(challenge, nonce)
	assert.Error(t, err)
}
