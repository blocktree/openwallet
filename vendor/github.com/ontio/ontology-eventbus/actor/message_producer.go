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
package actor

import "time"

type MessageProducer interface {
	// Tell sends a messages asynchronously to the PID
	Tell(pid *PID, message interface{})

	// Request sends a messages asynchronously to the PID. The actor may send a response back via respondTo, which is
	// available to the receiving actor via Context.Sender
	Request(pid *PID, message interface{}, respondTo *PID)

	// RequestFuture sends a message to a given PID and returns a Future
	RequestFuture(pid *PID, message interface{}, timeout time.Duration) *Future
}

type rootMessageProducer struct {
}

var (
	EmptyContext MessageProducer = &rootMessageProducer{}
)

// Tell sends a messages asynchronously to the PID
func (*rootMessageProducer) Tell(pid *PID, message interface{}) {
	pid.Tell(message)
}

// Request sends a messages asynchronously to the PID. The actor may send a response back via respondTo, which is
// available to the receiving actor via Context.Sender
func (*rootMessageProducer) Request(pid *PID, message interface{}, respondTo *PID) {
	pid.Request(message, respondTo)
}

// RequestFuture sends a message to a given PID and returns a Future
func (*rootMessageProducer) RequestFuture(pid *PID, message interface{}, timeout time.Duration) *Future {
	return pid.RequestFuture(message, timeout)
}
