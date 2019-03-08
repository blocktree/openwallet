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

// The Producer type is a function that creates a new actor
type Producer func() Actor

// Actor is the interface that defines the Receive method.
//
// Receive is sent messages to be processed from the mailbox associated with the instance of the actor
type Actor interface {
	Receive(c Context)
}

// The ActorFunc type is an adapter to allow the use of ordinary functions as actors to process messages
type ActorFunc func(c Context)

// Receive calls f(c)
func (f ActorFunc) Receive(c Context) {
	f(c)
}

type SenderFunc func(c Context, target *PID, envelope *MessageEnvelope)

//FromProducer creates a props with the given actor producer assigned
func FromProducer(actorProducer Producer) *Props {
	return &Props{actorProducer: actorProducer}
}

//FromFunc creates a props with the given receive func assigned as the actor producer
func FromFunc(f ActorFunc) *Props {
	return FromProducer(func() Actor { return f })
}

func FromSpawnFunc(spawn SpawnFunc) *Props {
	return &Props{spawner: spawn}
}

//Deprecated: FromInstance is deprecated
//Please use FromProducer(func() actor.Actor {...}) instead
func FromInstance(template Actor) *Props {
	return &Props{actorProducer: makeProducerFromInstance(template)}
}

//Deprecated: makeProducerFromInstance is deprecated.
func makeProducerFromInstance(a Actor) Producer {
	return func() Actor {
		return a
	}
}
