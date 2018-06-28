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

import "github.com/ontio/ontology-eventbus/eventstream"

type optionFn func()

// WithDeadLetterSubscriber option replaces the default DeadLetterEvent event subscriber with fn.
//
// fn will only receive *DeadLetterEvent messages
//
// Specifying nil will clear the existing.
func WithDeadLetterSubscriber(fn func(evt interface{})) optionFn {
	return func() {
		if deadLetterSubscriber != nil {
			eventstream.Unsubscribe(deadLetterSubscriber)
		}
		if fn != nil {
			deadLetterSubscriber = eventstream.Subscribe(fn).
				WithPredicate(func(m interface{}) bool {
					_, ok := m.(*DeadLetterEvent)
					return ok
				})
		}
	}
}

// WithSupervisorSubscriber option replaces the default SupervisorEvent event subscriber with fn.
//
// fn will only receive *SupervisorEvent messages
//
// Specifying nil will clear the existing.
func WithSupervisorSubscriber(fn func(evt interface{})) optionFn {
	return func() {
		if supervisionSubscriber != nil {
			eventstream.Unsubscribe(supervisionSubscriber)
		}
		if fn != nil {
			supervisionSubscriber = eventstream.Subscribe(fn).
				WithPredicate(func(m interface{}) bool {
					_, ok := m.(*SupervisorEvent)
					return ok
				})
		}
	}
}

// SetOptions is used to configure the actor system
func SetOptions(opts ...optionFn) {
	for _, opt := range opts {
		opt()
	}
}
