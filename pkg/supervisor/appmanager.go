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

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
)

const AppManagerLoggerTag = "APPMANAGER"

// AppManager is a service which has functionalities of
// - run app with environment variables, arguments
// - quit app via sending interrupt signal
// - wait for app to run for a while
// - wait for app to quit within a time
type AppManager struct {
	ctx    context.Context
	waitCh chan error
	log    log.Logger

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

	Logger log.Logger
}

func NewAppManager(cfg AppManagerConfig) (*AppManager, error) {
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}
	return &AppManager{
		waitCh:                 make(chan error, 1),
		envs:                   cfg.Envs,
		wd:                     cfg.WorkDir,
		bin:                    cfg.Bin,
		arguments:              cfg.Arguments,
		waitDurationForQuiting: cfg.WaitDurationForQuiting,
		log:                    cfg.Logger.WithField("tag", AppManagerLoggerTag),
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

	am.log.
		WithField("processName", processName).
		Info("Starting process")

	// Find the process running of which process name is given name
	process, err := isProcessRunning(processName)
	if err != nil {
		return err
	}
	// If process running, force quit process
	if process != nil {
		if err = process.Signal(syscall.SIGINT); err != nil {
			am.log.
				WithField("processName", processName).
				WithField("processId", process.Pid).
				Error("Failed interrupting process")
			return err
		}
	}

	// Make sure we start from running
	return am.RunApp()
}

func (am *AppManager) IsAppRunning() bool {
	return am.process != nil
}

// RunApp executes command to run application by setting environment variables and passing arguments
func (am *AppManager) RunApp() error {
	if am.process != nil { // already running
		go am.waitProcess()
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
		am.log.WithFields(log.Fields{
			"env":     am.envs,
			"workdir": am.wd,
			"bin":     am.bin,
			"args":    am.arguments,
		}).Error("Failed to run app")
		return err
	}

	am.process = cmd.Process
	am.log.WithFields(log.Fields{
		"env":     am.envs,
		"workdir": am.wd,
		"bin":     am.bin,
		"args":    am.arguments,
	}).Debug("Process started: ", am.process.Pid)
	go am.waitProcess()

	return nil
}

// QuitApp send interrupt signal to exit process and waits for app to quit
// App should quit within a time of waitDurationForQuiting.
// If detected that app was quited, return nil.
// If app is still running with expiration of waitDurationForQuiting, return error.
func (am *AppManager) QuitApp() error {
	if am.process == nil {
		return nil
	}

	err := am.process.Signal(syscall.SIGINT)
	if err != nil {
		return fmt.Errorf("error sending SIGINT for app to quit: %w", err)
	}

	// Wait for app to quit within a time of waitDurationForQuiting
	select {
	// case <-am.ctx.Done():
	//	return fmt.Errorf("context was cancelled while waiting for app to quit")
	case err := <-am.waitCh:
		if err != nil {
			return fmt.Errorf("error waiting for app to quit: %w", err)
		}
	case <-time.After(am.waitDurationForQuiting):
		am.log.Error("Waiting timeout to quit", am.process.Pid)

		// if timeout elapsed, kill the process again and return timeout error
		err := am.process.Signal(syscall.SIGTERM)
		if err != nil {
			return fmt.Errorf("sending sigterm signal failed: %w", err)
		}

		// Wait for the process to exit after sending SIGTERM
		err = <-am.waitCh
		if err != nil {
			return fmt.Errorf("error waiting for app to quit after sending SIGTERM: %w", err)
		}
		return fmt.Errorf("app quited within specified timeout")
	}
	am.log.Info("Process exited: ", am.wd, am.bin)
	return nil
}

func (am *AppManager) waitProcess() {
	if am.process == nil {
		return
	}

	_, err := am.process.Wait()
	am.process = nil
	am.waitCh <- err
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
