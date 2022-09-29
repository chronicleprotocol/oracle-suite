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

package replayer

import (
	"container/list"
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/event/publisher"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

// Config is the configuration for EventProvider.
type Config struct {
	// EventProvider is the event provider to replay events from.
	EventProvider publisher.EventProvider
	// Interval specifies the interval at which the event cache is checked for
	// events that need to be replayed.
	Interval time.Duration
	// ReplayAfter is a list of time durations after which events should be
	// replayed.
	ReplayAfter []time.Duration
}

// EventProvider replays events from the event provider at configurable time
// periods. It is used to guarantee that events are eventually delivered to
// subscribers even if they are not online at the time the event was published.
//
// It wraps another event provider and forwards all events from that provider,
// at the same time all received events are cached. The cache is checked at the
// configured interval, and events that are older than the configured playback
// periods are replayed. Events are removed from the cache when they are older
// than the oldest playback period.
type EventProvider struct {
	mu            sync.Mutex
	eventCh       chan *messages.Event
	eventCache    events
	eventProvider publisher.EventProvider
	expireAfter   time.Duration
	interval      time.Duration
	replayAfter   []time.Duration
}

// New returns a new instance of the EventProvider struct.
func New(cfg Config) (*EventProvider, error) {
	if cfg.EventProvider == nil {
		return nil, errors.New("eventProvider must not be nil")
	}
	if len(cfg.ReplayAfter) == 0 {
		return nil, errors.New("replayAfter must not be empty")
	}
	if cfg.Interval == 0 {
		return nil, errors.New("interval must not be zero")
	}
	// Find the oldest replayAfter time and use it as expireAfter.
	// The expireAfter field indicates how long an event can be kept in
	// the cache.
	sort.Slice(cfg.ReplayAfter, func(i, j int) bool {
		return cfg.ReplayAfter[i] < cfg.ReplayAfter[j]
	})
	expireAfter := cfg.ReplayAfter[len(cfg.ReplayAfter)-1]
	return &EventProvider{
		eventCh:       make(chan *messages.Event),
		eventCache:    events{list: list.New()},
		eventProvider: cfg.EventProvider,
		interval:      cfg.Interval,
		expireAfter:   expireAfter + cfg.Interval,
		replayAfter:   cfg.ReplayAfter,
	}, nil
}

// Start implements the publisher.EventPublisher interface.
func (r *EventProvider) Start(ctx context.Context) error {
	go r.collectEventsRoutine(ctx)
	go r.replayEventsRoutine(ctx)
	return r.eventProvider.Start(ctx)
}

// Events implements the publisher.EventPublisher interface.
func (r *EventProvider) Events() chan *messages.Event {
	return r.eventCh
}

// collectEventsRoutine collects events from the event provider and adds them to
// the cache.
func (r *EventProvider) collectEventsRoutine(ctx context.Context) {
	evtCh := r.eventProvider.Events()
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-evtCh:
			func() {
				r.mu.Lock()
				defer r.mu.Unlock()
				r.eventCache.add(evt)
				r.eventCh <- evt
			}()
		}
	}
}

// replayEventsRoutine replays events from the cache at the configured time
// periods.
func (r *EventProvider) replayEventsRoutine(ctx context.Context) {
	last := time.Now() // Used to calculate the time since the last replay.
	t := time.NewTicker(r.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			func() {
				r.mu.Lock()
				defer r.mu.Unlock()
				now := time.Now()
				r.eventCache.iterate(func(evt *messages.Event) {
					age := now.Sub(evt.EventDate)
					if age > r.expireAfter {
						r.eventCache.remove()
						return
					}
					for _, from := range r.replayAfter {
						to := from + now.Sub(last)
						if age >= from && age < to {
							r.eventCh <- evt
						}
						if to >= age {
							// The replayAfter times are sorted, so we can
							// break here.
							break
						}
					}
				})
				last = now
			}()
		}
	}
}

// events is a list of events. It is optimized for frequent additions and
// removals.
type events struct {
	list *list.List
	last *list.Element
}

// add adds an event to the list.
func (m *events) add(event *messages.Event) {
	m.last = m.list.PushBack(event)
}

// iterate iterates over the list and calls the given function for each event.
func (m *events) iterate(fn func(*messages.Event)) {
	var next *list.Element
	for e := m.list.Front(); e != nil; e = next {
		m.last = e
		next = e.Next()
		fn(e.Value.(*messages.Event))
	}
}

// remove removes the last added event from the list or the last iterated event.
func (m *events) remove() {
	if m.last == nil {
		return
	}
	m.list.Remove(m.last)
}
