package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/zhashkevych/quotes-server/internal/server/mocks"
)

type powMockBehavior func(m *mocks.MockProofOfWorkManager)
type quoterMockBehavior func(m *mocks.MockQuoter, response string)

type testCase struct {
	description      string
	powDifficulty    int
	challenge        string
	nonce            string
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
			nonce:            "42",
			expectedResponse: "The only true wisdom is in knowing you know nothing. - Socrates",
			powMockBehavior: func(m *mocks.MockProofOfWorkManager) {
				m.EXPECT().GenerateChallenge(4).Return("123456789", nil)
				m.EXPECT().VerifySolution("123456789", 42).Return(true, nil)

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
			nonce:            "42",
			expectedResponse: "The only true wisdom is in knowing you know nothing. - Socrates",
			powMockBehavior: func(m *mocks.MockProofOfWorkManager) {
				m.EXPECT().GenerateChallenge(4).Return("123456789", nil)
				m.EXPECT().VerifySolution("123456789", 42).Return(false, nil)

			},
			quoterMockBehavior:     func(m *mocks.MockQuoter, response string) {},
			verificationShouldFail: true,
		},
		{
			description:      "Request exceeds limit size",
			powDifficulty:    4,
			challenge:        "123456789",
			nonce:            strings.Repeat("a", maxRequestSize+1),
			expectedResponse: "Request data too large. Limit is 1024 bytes.\n",
			powMockBehavior: func(m *mocks.MockProofOfWorkManager) {
				m.EXPECT().GenerateChallenge(4).Return("123456789", nil)
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

			tc.powMockBehavior(powManager)
			tc.quoterMockBehavior(quoter, tc.expectedResponse)

			// init server
			server := NewTCPServer(0, tc.powDifficulty, quoter, powManager)

			go func() {
				err := server.ListenAndServe()
				assert.NoError(t, err)
			}()
			time.Sleep(time.Second) //g wait for the server initialization inside gorotuine

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
