/****************************************************
Copyright 2018 The ont-eventbus Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*****************************************************/

/***************************************************
Copyright 2016 https://github.com/AsynkronIT/protoactor-go

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*****************************************************/

package log

import "sync"

var es = &eventStream{}

func Subscribe(fn func(evt Event)) *Subscription {
	return es.Subscribe(fn)
}

func Unsubscribe(sub *Subscription) {
	es.Unsubscribe(sub)
}

type eventStream struct {
	sync.RWMutex
	subscriptions []*Subscription
}

func (es *eventStream) Subscribe(fn func(evt Event)) *Subscription {
	es.Lock()
	sub := &Subscription{
		es: es,
		i:  len(es.subscriptions),
		fn: fn,
	}
	es.subscriptions = append(es.subscriptions, sub)
	es.Unlock()
	return sub
}

func (ps *eventStream) Unsubscribe(sub *Subscription) {
	if sub.i == -1 {
		return
	}

	ps.Lock()
	i := sub.i
	l := len(ps.subscriptions) - 1

	ps.subscriptions[i] = ps.subscriptions[l]
	ps.subscriptions[i].i = i
	ps.subscriptions[l] = nil
	ps.subscriptions = ps.subscriptions[:l]
	sub.i = -1

	// TODO(SGC): implement resizing
	if len(ps.subscriptions) == 0 {
		ps.subscriptions = nil
	}

	ps.Unlock()
}

func (ps *eventStream) Publish(evt Event) {
	ps.RLock()
	defer ps.RUnlock()

	for _, s := range ps.subscriptions {
		if evt.Level >= s.l {
			s.fn(evt)
		}
	}
}

// Subscription is returned from the Subscribe function.
//
// This value and can be passed to Unsubscribe when the observer is no longer interested in receiving messages
type Subscription struct {
	es *eventStream
	i  int
	fn func(event Event)
	l  Level
}

// WithMinLevel filter messages below the provided level
//
// For example, setting ErrorLevel will only pass error messages. Setting MinLevel will
// allow all messages, and is the default.
func (s *Subscription) WithMinLevel(level Level) *Subscription {
	s.es.Lock()
	s.l = level
	s.es.Unlock()
	return s
}
