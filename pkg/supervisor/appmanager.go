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

	// List of arguments to be passed when execute a command
	Arguments []string

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

// QuitApp send interrupt signal to exit process and waits for app to quit
// App should quit within a time of waitDurationForQuiting.
// If detected that app was quited, return nil.
// If app is still running with expiration of waitDurationForQuiting, return error.
func (am *AppManager) QuitApp() error {
	if am.process == nil {
		return fmt.Errorf("unknown process to quit")
	}
	err := am.process.Signal(syscall.SIGINT)
	if err != nil {
		return fmt.Errorf("error sending SIGINT for app to quit: %w", err)
	}

	// Wait for app to quit within a time of waitDurationForQuiting
	done := make(chan error, 1)
	go func() {
		_, err := am.process.Wait()
		am.process = nil
		done <- err
	}()

	select {
	case <-am.ctx.Done():
		return fmt.Errorf("context was cancelled while waiting for app to quit")
	case err := <-done:
		if err != nil {
			return fmt.Errorf("error waiting for app to quit: %w", err)
		}
	case <-time.After(am.waitDurationForQuiting):
		// if timeout elapsed, kill the process again and return timeout error
		err := am.process.Signal(syscall.SIGTERM)
		if err != nil {
			return fmt.Errorf("sending sigterm signal failed: %w", err)
		}

		// Wait for the process to exit after sending SIGTERM
		if err = <-done; err != nil {
			return fmt.Errorf("error waiting for app to quit after sending SIGTERM: %w", err)
		}
		return fmt.Errorf("app quited within specified timeout")
	}
	return nil
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
