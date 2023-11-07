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

package morph

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/hashicorp/hcl/v2"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/reflectutil"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

type EnvVarsConfig struct {
	EnvVars map[string]string `hcl:"env_vars"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`
}

type Morph struct {
	ctx    context.Context
	waitCh chan error

	morphFile string
	interval  *timeutil.Ticker
	base      reflect.Value
	log       log.Logger
}

type Config struct {
	MorphFile string
	Interval  *timeutil.Ticker
	Base      reflect.Value
	Logger    log.Logger
}

const LoggerTag = "MORPH"

// NewMorphService creates Morph, which proceeds the following steps:
// - Periodically pull the config/env from on-chain.
// - Compares with previous one, if found difference, exit app.
func NewMorphService(cfg Config) (*Morph, error) {
	m := &Morph{
		waitCh:    make(chan error),
		log:       cfg.Logger.WithField("tag", LoggerTag),
		morphFile: cfg.MorphFile,
		interval:  cfg.Interval,
		base:      cfg.Base,
	}
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}
	return m, nil
}

func (m *Morph) Start(ctx context.Context) error {
	if m.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	m.ctx = ctx
	m.log.
		WithFields(log.Fields{
			"interval": m.interval.Duration(),
		}).
		Debug("Starting")
	m.interval.Start(m.ctx)
	go m.reloadRoutine()
	go m.contextCancelHandler()
	return nil
}

func (m *Morph) Wait() <-chan error {
	return m.waitCh
}

func (m *Morph) Monitor() error {
	// todo, pull down from IPFS
	var onChainConfig = m.base.Interface().(config.HasDefaults).DefaultEmbeds()

	// Load env variables from on-chain config
	var vars EnvVarsConfig
	err := config.LoadEmbeds(&vars, onChainConfig)
	if err != nil {
		m.log.WithError(err).Error("Failed loading on-chain config for env vars")
		return err
	}

	// Set env variables to OS ENV
	for key, value := range vars.EnvVars {
		os.Setenv(key, value)
	}

	// Create new instance with same type
	alternative := reflect.New(m.base.Type().Elem())
	alternativeVal := alternative.Interface()
	// Load again on-chain hcl config
	err = config.LoadEmbeds(&alternativeVal, onChainConfig)

	// Cleanup OS ENV
	for key := range vars.EnvVars {
		os.Setenv(key, "")
	}

	if err != nil {
		fields := log.Fields{}
		for key, value := range vars.EnvVars {
			fields[key] = value
		}
		m.log.WithError(err).WithFields(fields).Error("Failed loading on-chain config with env vars")
		return err
	}

	if reflectutil.DeepEqual(m.base.Interface(), alternative.Interface(), m.filterValue) {
		return nil
	}

	// todo, export to cache config file

	os.Exit(1)
	return nil
}

var (
	hclRangeTy       = reflect.TypeOf((*hcl.Range)(nil)).Elem()
	hclBodyTy        = reflect.TypeOf((*hcl.Body)(nil)).Elem()
	hclBodyContentTy = reflect.TypeOf((*hcl.BodyContent)(nil)).Elem()
)

func (m *Morph) filterValue(v1, v2 any) bool {
	refVal1, ok1 := v1.(reflect.Value)
	refVal2, ok2 := v2.(reflect.Value)
	if ok1 != ok2 {
		return false
	}
	if ok1 && ok2 {
		if refVal1.Type() == hclRangeTy || refVal2.Type() == hclRangeTy {
			return false
		}
		if refVal1.Type() == hclBodyTy || refVal2.Type() == hclBodyTy {
			return false
		}
		if refVal1.Type() == hclBodyContentTy || refVal2.Type() == hclBodyContentTy {
			return false
		}
	}
	refStruct1, ok1 := v1.(reflect.StructField)
	refStruct2, ok2 := v1.(reflect.StructField)
	if ok1 != ok2 {
		return false
	}
	if ok1 && ok2 {
		if _, tagged := refStruct1.Tag.Lookup("hcl"); !tagged {
			return false
		}
		if _, tagged := refStruct2.Tag.Lookup("hcl"); !tagged {
			return false
		}
	}
	return true
}

func (m *Morph) reloadRoutine() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-m.interval.TickCh():
			err := m.Monitor()
			if err != nil {
				fmt.Println(err)
			}
			return
		}
	}
}

func (m *Morph) contextCancelHandler() {
	defer func() { close(m.waitCh) }()
	defer m.log.Info("Stopped")
	<-m.ctx.Done()
}
