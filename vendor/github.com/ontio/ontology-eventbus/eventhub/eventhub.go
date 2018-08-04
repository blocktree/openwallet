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
package eventhub

import (
	"math/rand"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/orcaman/concurrent-map"
)

type PublishPolicy int

type RoundRobinState struct {
	state map[string]int
}

const (
	PublishPolicyAll = iota
	PublishPolicyRoundRobin
	PublishPolicyRandom
)

type EventHub struct {
	//sync.RWMutex
	Subscribers cmap.ConcurrentMap
	RoundRobinState
}

type Event struct {
	Publisher *actor.PID
	Topic     string
	Message   interface{}
	Policy    PublishPolicy
}

var GlobalEventHub = &EventHub{Subscribers: cmap.New(), RoundRobinState: RoundRobinState{make(map[string]int)}}

func (this *EventHub) Publish(event *Event) {
	//go func() {
	actors, ok := this.Subscribers.Get(event.Topic)
	if !ok {
		return
	}
	subscribers := actors.([]*actor.PID)
	this.sendEventByPolicy(subscribers, event, this.RoundRobinState)
	//}()
}

func (this *EventHub) Subscribe(topic string, subscriber *actor.PID) {
	subscribers, _ := this.Subscribers.Get(topic)

	//defer this.RWMutex.Unlock()
	//this.RWMutex.Lock()
	if subscribers == nil {
		this.Subscribers.Set(topic, []*actor.PID{subscriber})
	} else {
		this.Subscribers.Set(topic, append(subscribers.([]*actor.PID), subscriber))
	}

}

func (this *EventHub) Unsubscribe(topic string, subscriber *actor.PID) {

	tmpslice, ok := this.Subscribers.Get(topic)
	if !ok {
		return
	}
	//defer this.RWMutex.Unlock()
	//this.RWMutex.Lock()
	subscribers := tmpslice.([]*actor.PID)
	for i, s := range subscribers {
		if s == subscriber {
			this.Subscribers.Set(topic, append(subscribers[0:i], subscribers[i+1:]...))
			return
		}
	}

}

func (this *EventHub) sendEventByPolicy(subscribers []*actor.PID, event *Event, state RoundRobinState) {
	switch event.Policy {
	case PublishPolicyAll:
		for _, subscriber := range subscribers {
			subscriber.Request(event.Message, event.Publisher)
		}
	case PublishPolicyRandom:
		length := len(subscribers)
		if length == 0 {
			return
		}
		var i int
		i = rand.Intn(length)
		subscribers[i].Request(event.Message, event.Publisher)
	case PublishPolicyRoundRobin:
		latestIdx := state.state[event.Topic]
		i := latestIdx + 1
		if i < 0 {
			latestIdx = 0
			i = 0
		}
		state.state[event.Topic] = i
		mod := len(subscribers)
		subscribers[i%mod].Request(event.Message, event.Publisher)
	}
}

func (this *EventHub) RemovePID(pid actor.PID) {
	if this.Subscribers.Count() == 0 {
		return
	}
	keys := this.Subscribers.Keys()
	for index, _ := range keys {
		this.Unsubscribe(keys[index], &pid)
	}
}
