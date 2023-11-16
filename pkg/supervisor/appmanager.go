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

const WaitTimeout = 3000 // in second

type AppManager struct {
	ctx                    context.Context
	envs                   []string
	wd                     string
	bin                    string
	arguments              []string
	waitDurationForRunning time.Duration
	waitDurationForQuiting time.Duration

	process *os.Process
}

type AppManagerConfig struct {
	Envs                   []string
	WorkDir                string
	Bin                    string
	Arguments              []string
	WaitDurationForRunning time.Duration
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
		if err := am.WaitForAppRunning(); err != nil {
			return err
		}
	}

	return nil
}

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

func (am *AppManager) WaitForAppRunning() error {
	ctxChild, ctxCancel := context.WithTimeout(am.ctx, am.waitDurationForRunning)
	defer ctxCancel()

	ticker := timeutil.NewTicker(WaitTimeout * time.Millisecond)
	ticker.Start(am.ctx)
	for {
		select {
		case <-ctxChild.Done():
			return fmt.Errorf("timeout waiting for app to start")
		case <-ticker.TickCh():
			if am.isAppRunning() {
				return nil
			}
		}
	}
}

func (am *AppManager) QuitApp() error {
	if am.process != nil {
		err := am.process.Signal(syscall.SIGINT)
		return err
	}
	return nil
}

func (am *AppManager) WaitForAppQuitting() error {
	ctxChild, ctxCancel := context.WithTimeout(am.ctx, am.waitDurationForQuiting)
	defer ctxCancel()

	ticker := timeutil.NewTicker(WaitTimeout * time.Millisecond)
	for {
		select {
		case <-ctxChild.Done():
			return fmt.Errorf("timeout waiting for app to quit")
		case <-ticker.TickCh():
			if !am.isAppRunning() {
				return nil
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
