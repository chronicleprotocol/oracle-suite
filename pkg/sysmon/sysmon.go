package sysmon

import (
	"context"
	"errors"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"golang.org/x/sys/unix"

	suite "github.com/chronicleprotocol/oracle-suite"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

const LoggerTag = "SYSMON"

var startTime = time.Now()

// Sysmon is a system monitor. It periodically logs system status.
// It works only if the logger uses the debug level.
type Sysmon struct {
	ctx      context.Context
	waitCh   chan error
	interval time.Duration
	log      log.Logger
}

// New returns a new instance of Sysmon.
func New(interval time.Duration, logger log.Logger) *Sysmon {
	return &Sysmon{
		waitCh:   make(chan error),
		interval: interval,
		log:      logger.WithField("tag", LoggerTag),
	}
}

// Start implements the supervisor.Service interface.
func (s *Sysmon) Start(ctx context.Context) error {
	if !log.IsLevel(s.log, log.Debug) {
		// Sysmon shows logs only for the debug level, if current logger uses
		// a lower log level, there is no point of starting the service.
		close(s.waitCh)
		return nil
	}
	if s.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	s.log.Info("Starting")
	fields := log.Fields{
		"appVersion": suite.Version,
		"goVersion":  runtime.Version(),
		"goCompiler": runtime.Compiler,
		"goOS":       runtime.GOOS,
		"goArch":     runtime.GOARCH,
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "CGO_ENABLED":
				fields["cgoEnabled"] = setting.Value
			case "vcs.revision":
				fields["gitCommit"] = setting.Value
			case "vcs.modified":
				fields["gitModified"] = setting.Value
			}
		}
	}
	s.log.WithFields(fields).Debug("Build info")
	s.ctx = ctx
	go s.monitorRoutine()
	go s.contextCancelHandler()
	return nil
}

// Wait implements the supervisor.Service interface.
func (s *Sysmon) Wait() <-chan error {
	return s.waitCh
}

func (s *Sysmon) monitorRoutine() {
	var m runtime.MemStats
	var stat unix.Statfs_t
	var spaceAvail uint64
	wd, err := os.Getwd()
	if err != nil {
		s.log.WithError(err).Warn("Failed to get current working directory, disk space monitoring is disabled")
	}
	t := time.NewTicker(s.interval)
	defer t.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-t.C:
			if len(wd) > 0 && unix.Statfs(wd, &stat) == nil {
				spaceAvail = stat.Bavail * uint64(stat.Bsize)
			} else {
				spaceAvail = 0
			}
			runtime.ReadMemStats(&m)
			s.log.
				WithFields(log.Fields{
					"uptime":              time.Since(startTime),
					"memAlloc":            m.Alloc,
					"memTotalAlloc":       m.TotalAlloc,
					"memSys":              m.Sys,
					"memNumGC":            m.NumGC,
					"runtimeNumCPU":       runtime.NumCPU(),
					"runtimeNumGoroutine": runtime.NumGoroutine(),
					"spaceAvail":          spaceAvail,
				}).
				Debug("Status")
		}
	}
}

// contextCancelHandler handles context cancellation.
func (s *Sysmon) contextCancelHandler() {
	defer func() { close(s.waitCh) }()
	defer s.log.Info("Stopped")
	<-s.ctx.Done()
}
