package gofere2e

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/chronicleprotocol/infestor/smocker"
	"github.com/stretchr/testify/suite"
)

const smockerPort = 8081

type SmockerAPISuite struct {
	suite.Suite
	api        smocker.API
	url        string
	ConfigPath string
}

func (s *SmockerAPISuite) Setup() {
	var err error

	smockerHost, exist := os.LookupEnv("SMOCKER_HOST")
	if !exist {
		smockerHost = "http://localhost"
	}

	s.api = smocker.API{
		URL: fmt.Sprintf("%s:%d", smockerHost, smockerPort),
	}

	s.url = fmt.Sprintf("%s:8080", smockerHost)

	goferConfigPath, exist := os.LookupEnv("GOFER_CONFIG")
	if !exist {
		goferConfigPath, err = filepath.Abs("./gofer_config.json")
		s.Require().NoError(err)
	}
	s.ConfigPath = goferConfigPath
}

func (s *SmockerAPISuite) Reset() {
	err := s.api.Reset(context.Background())
	s.Require().NoError(err)
}

func (s *SmockerAPISuite) SetupSuite() {
	s.Setup()
}

func (s *SmockerAPISuite) SetupTest() {
	s.Reset()
}

type goferPrice struct {
	Type       string
	Base       string
	Quote      string
	Price      float64
	Bid        float64
	Ask        float64
	Volume24h  float64           `json:"vol24h"`
	Timestamp  time.Time         `json:"ts"`
	Parameters map[string]string `json:"params,omitempty"`
	Prices     []goferPrice      `json:"prices,omitempty"`
	Error      string            `json:"error,omitempty"`
}

func parseGoferPrice(price string) (goferPrice, error) {
	var p goferPrice
	err := json.Unmarshal([]byte(price), &p)
	if err != nil {
		return p, err
	}
	return p, nil
}

func callGofer(params ...string) (string, int, error) {
	goferBin := os.Getenv("GOFER_BINARY_PATH")
	if goferBin == "" {
		goferBin = "gofer"
	}

	cmd := exec.Command(goferBin, params...)
	cmd.Env = os.Environ()

	out, err := cmd.Output()

	if err != nil {
		fmt.Println("Graph error:")
		fmt.Println("Output: ", string(out))
		fmt.Println("Error: ", err.Error())
	}

	if werr, ok := err.(*exec.ExitError); ok {
		if s := werr.Error(); s != "0" {
			if status, ok := werr.Sys().(syscall.WaitStatus); ok {
				return "", status.ExitStatus(), fmt.Errorf("gofer exited with exit code: %d", status.ExitStatus())
			}
			return "", 1, fmt.Errorf("gofer exited with exit code: %s", s)
		}
	}

	return strings.TrimSpace(string(out)), 0, nil
}
