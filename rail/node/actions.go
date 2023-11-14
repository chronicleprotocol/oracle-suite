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

package node

import (
	"github.com/libp2p/go-libp2p/core/event"
)

type Action func(*Node) error

func closeSub(sub event.Subscription) {
	log.Debugw("closing", "subscription", sub.Name())
	go func() {
		log.Debugf("draining %T", sub.Out())
		for e := range sub.Out() {
			log.Debugf("got %T for %s", e, sub.Name())
		}
	}()
	if err := sub.Close(); err != nil {
		log.Errorw("error closing", "error", err, "subscription", sub.Name())
		return
	}
	log.Debugw("closed", "subscription", sub.Name())
}
