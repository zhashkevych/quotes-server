package hashcash

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Hashcash implements hashcash logic for proof-of-work challenge-response mechanism
type Hashcash struct{}

func New() *Hashcash {
	return &Hashcash{}
}

func (h *Hashcash) GenerateChallenge(difficulty int) (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", errors.Wrap(err, "error generating random seed")
	}

	seed := hex.EncodeToString(b)
	return fmt.Sprintf("%s:%d", seed, difficulty), nil
}

func (h *Hashcash) SolveChallenge(challenge string) (int, error) {
	parts := strings.Split(challenge, ":")
	seed := parts[0]
	difficulty, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	var nonce int
	for {
		data := fmt.Sprintf("%s:%d", seed, nonce)
		hash := sha256.Sum256([]byte(data))
		hashStr := hex.EncodeToString(hash[:])

		if strings.HasPrefix(hashStr, strings.Repeat("0", difficulty)) {
			return nonce, nil
		}
		nonce++
	}
}

func (h *Hashcash) VerifySolution(challenge string, nonce int) (bool, error) {
	parts := strings.Split(challenge, ":")
	if len(parts) < 2 {
		return false, errors.New("invalid challenge format")
	}

	seed := parts[0]
	difficulty, err := strconv.Atoi(parts[1])
	if err != nil {
		return false, errors.Wrap(err, "invalid challenge format")
	}

	data := fmt.Sprintf("%s:%d", seed, nonce)
	hash := sha256.Sum256([]byte(data))
	hashStr := hex.EncodeToString(hash[:])

	return strings.HasPrefix(hashStr, strings.Repeat("0", difficulty)), nil
}
