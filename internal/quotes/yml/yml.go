package quotes

import (
	"math/rand"
	"os"

	"gopkg.in/yaml.v2"
)

type Quotes struct {
	Quotes []string `yaml:"quotes"`
}

type YMLService struct {
	quotes []string
}

func NewYMLService(filepath string) (*YMLService, error) {
	ymlData, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var quotes Quotes
	err = yaml.Unmarshal(ymlData, &quotes)
	if err != nil {
		return nil, err
	}

	return &YMLService{quotes.Quotes}, nil
}

func (s *YMLService) GetRandomQuote() string {
	if len(s.quotes) == 0 {
		return ""
	}

	return s.quotes[rand.Intn(len(s.quotes))]
}
