package server

import (
	"bufio"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/zhashkevych/quotes-server/internal/server/mocks"
)

type powMockBehavior func(m *mocks.MockProofOfWorkManager, powDifficulty, nonce int, challenge string)
type quoterMockBehavior func(m *mocks.MockQuoter, response string)

type testCase struct {
	description      string
	powDifficulty    int
	challenge        string
	nonce            int
	expectedResponse string

	powMockBehavior    powMockBehavior
	quoterMockBehavior quoterMockBehavior

	verificationShouldFail bool
}

func TestTCPServer_HandleConnection(t *testing.T) {
	tests := []testCase{
		{
			description:      "Success",
			powDifficulty:    4,
			challenge:        "123456789",
			nonce:            42,
			expectedResponse: "The only true wisdom is in knowing you know nothing. - Socrates",
			powMockBehavior: func(m *mocks.MockProofOfWorkManager, powDifficulty, nonce int, challenge string) {
				m.EXPECT().GenerateChallenge(powDifficulty).Return(challenge, nil)
				m.EXPECT().VerifySolution(challenge, nonce).Return(true, nil)

			},
			quoterMockBehavior: func(m *mocks.MockQuoter, response string) {
				m.EXPECT().GetRandomQuote().Return(response)
			},
			verificationShouldFail: false,
		},
		{
			description:      "Verification failed",
			powDifficulty:    4,
			challenge:        "123456789",
			nonce:            42,
			expectedResponse: "The only true wisdom is in knowing you know nothing. - Socrates",
			powMockBehavior: func(m *mocks.MockProofOfWorkManager, powDifficulty, nonce int, challenge string) {
				m.EXPECT().GenerateChallenge(powDifficulty).Return(challenge, nil)
				m.EXPECT().VerifySolution(challenge, nonce).Return(false, nil)

			},
			quoterMockBehavior:     func(m *mocks.MockQuoter, response string) {},
			verificationShouldFail: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			// setup testing dependencies
			c := gomock.NewController(t)
			defer c.Finish()

			quoter := mocks.NewMockQuoter(c)
			powManager := mocks.NewMockProofOfWorkManager(c)

			tc.powMockBehavior(powManager, tc.powDifficulty, tc.nonce, tc.challenge)
			tc.quoterMockBehavior(quoter, tc.expectedResponse)

			// init server
			server := NewTCPServer(0, tc.powDifficulty, quoter, powManager)

			go server.ListenAndServe()
			time.Sleep(time.Second) // wait for the server initialization inside gorotuine

			conn, err := net.Dial("tcp", server.getAddr())
			assert.NoError(t, err)

			defer conn.Close()

			// read & assert challange
			reader := bufio.NewReader(conn)
			challenge, err := reader.ReadString('\n')
			assert.NoError(t, err)

			assert.Equal(t, tc.challenge+"\n", challenge)

			// send nonce
			fmt.Fprintln(conn, tc.nonce)

			// read & assert the server's response
			response, err := reader.ReadString('\n')
			assert.NoError(t, err)
			if tc.verificationShouldFail {
				assert.Equal(t, IncorrectSolutionResonse+"\n", response)
			} else {
				assert.Equal(t, tc.expectedResponse+"\n", response)
			}

		})
	}
}

// used for testing
func (s *TCPServer) getAddr() string {
	return s.listener.Addr().String()
}
