//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package supervisor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/mitchellh/go-ps"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

const WaitCheckInterval = 1000 // in second

// AppManager is a service which has functionalities of
// - run app with environment variables, arguments
// - quit app via sending interrupt signal
// - wait for app to run for a while
// - wait for app to quit within a time
type AppManager struct {
	ctx                    context.Context
	envs                   []string
	wd                     string
	bin                    string
	arguments              []string
	waitDurationForRunning time.Duration
	waitDurationForQuiting time.Duration

	// Process instance that indicates to the process which AppManager handles
	process *os.Process
}

type AppManagerConfig struct {
	// List of environment variables, with the format of "x=y"
	Envs []string

	// Working directory to execute command to run application
	WorkDir string

	// Executable binary to execute
	Bin string

	// List of arguments to be passed when execute commmand
	Arguments []string

	// Time duration to wait for app seamlessly running for a while
	// App should not be crashed itself before expiration of WaitDurationForRunning
	WaitDurationForRunning time.Duration

	// Time duration to wait for app seamlessly to quit within a time
	// App should quit itself before expiration of WaitDurationForQuiting
	WaitDurationForQuiting time.Duration
}

func NewAppManager(cfg AppManagerConfig) (*AppManager, error) {
	return &AppManager{
		envs:                   cfg.Envs,
		wd:                     cfg.WorkDir,
		bin:                    cfg.Bin,
		arguments:              cfg.Arguments,
		waitDurationForRunning: cfg.WaitDurationForRunning,
		waitDurationForQuiting: cfg.WaitDurationForQuiting,
	}, nil
}

// Start checks if application was already running and make sure that app is running
// If running, take its process, else execute command to run application
func (am *AppManager) Start(ctx context.Context) error {
	am.ctx = ctx

	// Extract process name from bin
	filename := filepath.Base(am.bin)
	ext := filepath.Ext(am.bin)
	processName := filename[0 : len(filename)-len(ext)]

	// Find the process running of which process name is given name
	process, err := isProcessRunning(processName)
	if err != nil {
		return err
	}
	am.process = process

	if am.process == nil {
		// Make sure we start from running
		if err := am.RunApp(); err != nil {
			return err
		}
		if _, err := am.WaitForAppRunning(); err != nil {
			return err
		}
	}

	return nil
}

// RunApp executes command to run application by setting environment variables and passing arguments
func (am *AppManager) RunApp() error {
	if am.process != nil { // already running
		return nil
	}

	// Running app with env variables
	env := os.Environ()
	env = append(env, am.envs...)
	cmd := exec.CommandContext(am.ctx, am.bin, am.arguments...) //nolint:gosec
	cmd.Dir = am.wd
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}

	am.process = cmd.Process
	return nil
}

// WaitForAppRunning waits for app to run, expecting app not crashed until the expiration of waitDurationForRunning
// If detected that app is not running until waitDurationForRunning, returns false and error.
// Returns true if app is running for a while, time of waitDurationForRunning.
func (am *AppManager) WaitForAppRunning() (bool, error) {
	startTime := time.Now()
	ticker := timeutil.NewTicker(WaitCheckInterval * time.Millisecond)
	ticker.Start(am.ctx)
	for {
		select {
		case <-am.ctx.Done():
			return false, fmt.Errorf("context was cancelled while waiting for app to run")
		case <-ticker.TickCh():
			if !am.isAppRunning() {
				return false, fmt.Errorf("app is not running for some reason")
			}
			if time.Since(startTime) >= am.waitDurationForRunning {
				return true, nil
			}
		}
	}
}

// QuitApp send interrupt signal to exit process
func (am *AppManager) QuitApp() error {
	if am.process != nil {
		err := am.process.Signal(syscall.SIGINT)
		return err
	}
	return nil
}

// WaitForAppQuiting waits for app to quit, app should quit within a time of waitDurationForQuiting.
// If detected that app was quited, return true.
// If app is still running with expiration of waitDurationForQuiting, return false and error.
func (am *AppManager) WaitForAppQuiting() (bool, error) {
	startTime := time.Now()
	ticker := timeutil.NewTicker(WaitCheckInterval * time.Millisecond)
	ticker.Start(am.ctx)
	for {
		select {
		case <-am.ctx.Done():
			return false, fmt.Errorf("context was cancelled while waiting for app to quit")
		case <-ticker.TickCh():
			running := am.isAppRunning()
			if !running {
				return true, nil
			}
			if time.Since(startTime) >= am.waitDurationForQuiting {
				return false, fmt.Errorf("app is still running that should quit")
			}
		}
	}
}

// isAppRunning checks if the application is still running.
func (am *AppManager) isAppRunning() bool {
	if am.process == nil {
		return false
	}

	// Send a signal of 0 to check if the process is still alive.
	err := am.process.Signal(syscall.Signal(0))
	return err == nil
}

// isProcessRunning finds the process which has a process name of given parameter
// Returns the os.Process instance and error
func isProcessRunning(processName string) (*os.Process, error) {
	processes, err := ps.Processes()
	if err != nil {
		return nil, err
	}

	for _, process := range processes {
		if process.Executable() == processName {
			return os.FindProcess(process.Pid())
		}
	}
	return nil, nil
}
