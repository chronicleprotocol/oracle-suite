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

package sets

import (
	"context"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

type Validator func(ctx context.Context, topic string, id peer.ID, msg *pubsub.Message) pubsub.ValidationResult

// ValidatorSet stores multiple instances of validators that implements
// the pubsub.ValidatorEx functions. Validators are groped by topic.
type ValidatorSet struct {
	validators []Validator
}

// NewValidatorSet creates new instance of the ValidatorSet.
func NewValidatorSet() *ValidatorSet {
	return &ValidatorSet{}
}

// Add adds new pubsub.ValidatorEx to the set.
func (n *ValidatorSet) Add(validator ...Validator) {
	n.validators = append(n.validators, validator...)
}

// Validator returns function that implements pubsub.ValidatorEx. That function
// will invoke all registered validators for given topic.
func (n *ValidatorSet) Validator(topic string) pubsub.ValidatorEx {
	return func(ctx context.Context, id peer.ID, psMsg *pubsub.Message) pubsub.ValidationResult {
		for _, validator := range n.validators {
			if result := validator(ctx, topic, id, psMsg); result != pubsub.ValidationAccept {
				return result
			}
		}
		return pubsub.ValidationAccept
	}
}
