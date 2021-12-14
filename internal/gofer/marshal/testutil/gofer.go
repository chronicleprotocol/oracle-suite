//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
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

package testutil

import (
	"errors"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/gofer"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer/graph"
	"github.com/chronicleprotocol/oracle-suite/pkg/gofer/graph/nodes"
)

func Gofer(ps ...gofer.Pair) gofer.Gofer {
	graphs := map[gofer.Pair]nodes.Aggregator{}
	for _, p := range ps {
		root := nodes.NewMedianAggregatorNode(p, 1)

		ttl := time.Second * time.Duration(time.Now().Unix()+10)
		on1 := nodes.NewOriginNode(nodes.OriginPair{Origin: "a", Pair: p}, 0, ttl)
		on2 := nodes.NewOriginNode(nodes.OriginPair{Origin: "b", Pair: p}, 0, ttl)
		in := nodes.NewIndirectAggregatorNode(p)
		mn := nodes.NewMedianAggregatorNode(p, 1)

		root.AddChild(on1)
		root.AddChild(in)
		root.AddChild(mn)

		in.AddChild(on1)
		mn.AddChild(on1)
		mn.AddChild(on2)

		_ = on1.Ingest(nodes.OriginPrice{
			PairPrice: nodes.PairPrice{
				Pair:      p,
				Price:     10,
				Bid:       10,
				Ask:       10,
				Volume24h: 10,
				Time:      time.Unix(10, 0),
			},
			Origin: "a",
			Error:  nil,
		})

		_ = on2.Ingest(nodes.OriginPrice{
			PairPrice: nodes.PairPrice{
				Pair:      p,
				Price:     20,
				Bid:       20,
				Ask:       20,
				Volume24h: 20,
				Time:      time.Unix(20, 0),
			},
			Origin: "b",
			Error:  errors.New("something"),
		})

		graphs[p] = root
	}

	return graph.NewGofer(graphs, nil)
}

func Models(ps ...gofer.Pair) map[gofer.Pair]*gofer.Model {
	g := Gofer(ps...)
	ns, err := g.Models()
	if err != nil {
		panic(err)
	}
	return ns
}

func Prices(ps ...gofer.Pair) map[gofer.Pair]*gofer.Price {
	g := Gofer(ps...)
	ts, err := g.Prices()
	if err != nil {
		panic(err)
	}
	return ts
}
