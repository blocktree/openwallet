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

type optionFn func()

// WithEventSubscriber option replaces the default Event subscriber with fn.
//
// Specifying nil will disable logging of events.
func WithEventSubscriber(fn func(evt Event)) optionFn {
	return func() {
		if sub != nil {
			Unsubscribe(sub)
		}
		if fn != nil {
			sub = Subscribe(fn)
		}
	}
}

// SetOptions is used to configure the log system
func SetOptions(opts ...optionFn) {
	for _, opt := range opts {
		opt()
	}
}
