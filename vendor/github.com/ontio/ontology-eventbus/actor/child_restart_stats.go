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
	"time"
)

//RestartStatistics keeps track of how many times an actor have restarted and when
type RestartStatistics struct {
	failureTimes []time.Time
}

//NewRestartStatistics construct a RestartStatistics
func NewRestartStatistics() *RestartStatistics {
	return &RestartStatistics{[]time.Time{}}
}

//FailureCount returns failure count
func (rs *RestartStatistics) FailureCount() int {
	return len(rs.failureTimes)
}

//Fail increases the associated actors failure count
func (rs *RestartStatistics) Fail() {
	rs.failureTimes = append(rs.failureTimes, time.Now())
}

//Reset the associated actors failure count
func (rs *RestartStatistics) Reset() {
	rs.failureTimes = []time.Time{}
}

//NumberOfFailures returns number of failures within a given duration
func (rs *RestartStatistics) NumberOfFailures(withinDuration time.Duration) int {
	if withinDuration == 0 {
		return len(rs.failureTimes)
	}

	num := 0
	currTime := time.Now()
	for _, t := range rs.failureTimes {
		if currTime.Sub(t) < withinDuration {
			num++
		}
	}
	return num
}
