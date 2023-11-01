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
	"fmt"
	"reflect"
	"time"

	"github.com/hashicorp/hcl/v2"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	morphService "github.com/chronicleprotocol/oracle-suite/pkg/morph"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

const defaultInterval = 60 * 60

type Config struct {
	MorphFile string `hcl:"cache_path"`
	Interval  uint32 `hcl:"interval"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`
}

type Dependencies struct {
	Base   reflect.Value
	Logger log.Logger
}

func (c *Config) ConfigureMorph(d Dependencies) (*morphService.Morph, error) {
	interval := c.Interval
	if interval == 0 {
		interval = defaultInterval
	}

	cfg := morphService.Config{
		MorphFile: c.MorphFile,
		Interval:  timeutil.NewTicker(time.Second * time.Duration(interval)),
		Base:      d.Base,
		Logger:    d.Logger,
	}
	morph, err := morphService.NewMorphService(cfg)
	if err != nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Runtime error",
			Detail:   fmt.Sprintf("Failed to create the Morph service: %v", err),
			Subject:  c.Range.Ptr(),
		}
	}
	return morph, nil
}
