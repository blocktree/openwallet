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
	"errors"
	"sync"

	"github.com/ontio/ontology-eventbus/log"
)

type guardiansValue struct {
	sync.RWMutex
	guardians map[SupervisorStrategy]*guardianProcess
}

var guardians = &guardiansValue{guardians: make(map[SupervisorStrategy]*guardianProcess)}

func (gs *guardiansValue) getGuardianPid(s SupervisorStrategy) *PID {
	gs.Lock()
	defer gs.Unlock()
	if g, ok := gs.guardians[s]; ok {
		return g.pid
	}
	g := gs.newGuardian(s)
	gs.guardians[s] = g
	//gs.guardians.Store(s, g)
	return g.pid
}

// newGuardian creates and returns a new actor.guardianProcess with a timeout of duration d
func (gs *guardiansValue) newGuardian(s SupervisorStrategy) *guardianProcess {
	ref := &guardianProcess{strategy: s}
	id := ProcessRegistry.NextId()

	pid, ok := ProcessRegistry.Add(ref, "guardian"+id)
	if !ok {
		plog.Error("failed to register guardian process", log.Stringer("pid", pid))
	}

	ref.pid = pid
	return ref
}

type guardianProcess struct {
	pid      *PID
	strategy SupervisorStrategy
}

func (g *guardianProcess) SendUserMessage(pid *PID, message interface{}) {
	panic(errors.New("Guardian actor cannot receive any user messages"))
}

func (g *guardianProcess) SendSystemMessage(pid *PID, message interface{}) {
	if msg, ok := message.(*Failure); ok {
		g.strategy.HandleFailure(g, msg.Who, msg.RestartStats, msg.Reason, msg.Message)
	}
}

func (g *guardianProcess) Stop(pid *PID) {
	//Ignore
}

func (g *guardianProcess) Children() []*PID {
	panic(errors.New("Guardian does not hold its children PIDs"))
}

func (*guardianProcess) EscalateFailure(reason interface{}, message interface{}) {
	panic(errors.New("Guardian cannot escalate failure"))
}

func (*guardianProcess) RestartChildren(pids ...*PID) {
	for _, pid := range pids {
		pid.sendSystemMessage(restartMessage)
	}
}

func (*guardianProcess) StopChildren(pids ...*PID) {
	for _, pid := range pids {
		pid.sendSystemMessage(stopMessage)
	}
}

func (*guardianProcess) ResumeChildren(pids ...*PID) {
	for _, pid := range pids {
		pid.sendSystemMessage(resumeMailboxMessage)
	}
}
