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
)

// ErrNameExists is the error used when an existing name is used for spawning an actor.
var ErrNameExists = errors.New("spawn: name exists")

type SpawnFunc func(id string, props *Props, parent *PID) (*PID, error)

// DefaultSpawner conforms to Spawner and is used to spawn a local actor
var DefaultSpawner SpawnFunc = spawn

// Spawn starts a new actor based on props and named with a unique id
func Spawn(props *Props) *PID {
	pid, _ := SpawnNamed(props, ProcessRegistry.NextId())
	return pid
}

// SpawnPrefix starts a new actor based on props and named using a prefix followed by a unique id
func SpawnPrefix(props *Props, prefix string) (*PID, error) {
	return SpawnNamed(props, prefix+ProcessRegistry.NextId())
}

// SpawnNamed starts a new actor based on props and named using the specified name
//
// If name exists, error will be ErrNameExists
func SpawnNamed(props *Props, name string) (*PID, error) {
	var parent *PID
	if props.guardianStrategy != nil {
		parent = guardians.getGuardianPid(props.guardianStrategy)
	}
	return props.spawn(name, parent)
}

func spawn(id string, props *Props, parent *PID) (*PID, error) {
	lp := &localProcess{}
	pid, absent := ProcessRegistry.Add(lp, id)
	if !absent {
		return pid, ErrNameExists
	}

	cell := newLocalContext(props.actorProducer, props.getSupervisor(), props.inboundMiddleware, props.outboundMiddleware, parent)
	mb := props.produceMailbox(cell, props.getDispatcher())
	lp.mailbox = mb
	var ref Process = lp
	pid.p = &ref
	cell.self = pid
	mb.Start()
	mb.PostSystemMessage(startedMessage)

	return pid, nil
}
