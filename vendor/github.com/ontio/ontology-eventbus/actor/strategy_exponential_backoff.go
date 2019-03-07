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

import (
	"math/rand"
	"time"
)

// NewExponentialBackoffStrategy creates a new Supervisor strategy that restarts a faulting child using an exponential
// back off algorithm:
//
//	delay =
func NewExponentialBackoffStrategy(backoffWindow time.Duration, initialBackoff time.Duration) SupervisorStrategy {
	return &exponentialBackoffStrategy{
		backoffWindow:  backoffWindow,
		initialBackoff: initialBackoff,
	}
}

type exponentialBackoffStrategy struct {
	backoffWindow  time.Duration
	initialBackoff time.Duration
}

func (strategy *exponentialBackoffStrategy) HandleFailure(supervisor Supervisor, child *PID, rs *RestartStatistics, reason interface{}, message interface{}) {
	strategy.setFailureCount(rs)

	backoff := rs.FailureCount() * int(strategy.initialBackoff.Nanoseconds())
	noise := rand.Intn(500)
	dur := time.Duration(backoff + noise)
	time.AfterFunc(dur, func() {
		supervisor.RestartChildren(child)
	})
}

func (strategy *exponentialBackoffStrategy) setFailureCount(rs *RestartStatistics) {
	if rs.NumberOfFailures(strategy.backoffWindow) == 0 {
		rs.Reset()
	}

	rs.Fail()
}
