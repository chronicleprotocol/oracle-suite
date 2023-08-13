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

package dataprovider

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/hcl/v2"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/graph"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/origin"

	"github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/sliceutil"
)

type Dependencies struct {
	HTTPClient *http.Client
	Clients    ethereum.ClientRegistry
	Logger     log.Logger
}

type Config struct {
	Origins    []configOrigin    `hcl:"origin,block"`
	DataModels []configDataModel `hcl:"data_model,block"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`
}

func (c *Config) ConfigureDataProvider(d Dependencies) (datapoint.Provider, error) {
	var err error

	// Configure origins:
	origins, err := c.configureOrigins(d)
	if err != nil {
		return nil, err
	}

	// Configure data models:
	models, err := c.configureDataModels(origins)
	if err != nil {
		return nil, err
	}

	// Configure data provider:
	return graph.NewProvider(models, graph.NewUpdater(origins, d.Logger)), nil
}

func (c *Config) configureOrigins(d Dependencies) (map[string]origin.Origin, error) {
	var err error
	origins := map[string]origin.Origin{}
	for _, o := range c.Origins {
		origins[o.Name], err = o.configureOrigin(d)
		if err != nil {
			return nil, err
		}
	}
	return origins, nil
}

func (c *Config) configureDataModels(origins map[string]origin.Origin) (map[string]graph.Node, error) {
	// First generate root nodes for each data model. It is necessary to do this
	// because the data models may reference each other.
	models := map[string]graph.Node{}
	for _, pm := range c.DataModels {
		models[pm.Name] = graph.NewReferenceNode()
	}

	// Configure each data model.
	for _, pm := range c.DataModels {
		dataModel, err := pm.configureDataModel(origins, models)
		if err != nil {
			return nil, err
		}
		if err := models[pm.Name].AddNodes(dataModel); err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail:   fmt.Sprintf("Failed to add node to data model %s: %s", pm.Name, err),
				Subject:  pm.Range.Ptr(),
			}
		}
		if nodes := graph.DetectCycle(models[pm.Name]); len(nodes) > 0 {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail: fmt.Sprintf(
					"Cycle detected in the data model %s: %s",
					pm.Name,
					strings.Join(sliceutil.Map(nodes, func(n graph.Node) string {
						return n.Meta()["type"].(string)
					}), " -> "),
				),
				Subject: pm.Range.Ptr(),
			}
		}
	}

	return models, nil
}
