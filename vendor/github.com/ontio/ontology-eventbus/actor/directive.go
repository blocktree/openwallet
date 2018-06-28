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

//Directive is an enum for supervision actions
type Directive int

// Directive determines how a supervisor should handle a faulting actor
const (
	// ResumeDirective instructs the supervisor to resume the actor and continue processing messages
	ResumeDirective Directive = iota

	// RestartDirective instructs the supervisor to discard the actor, replacing it with a new instance,
	// before processing additional messages
	RestartDirective

	// StopDirective instructs the supervisor to stop the actor
	StopDirective

	// EscalateDirective instructs the supervisor to escalate handling of the failure to the actor's parent supervisor
	EscalateDirective
)
