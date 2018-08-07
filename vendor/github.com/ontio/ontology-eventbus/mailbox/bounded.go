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
package mailbox

import (
	rbqueue "github.com/Workiva/go-datastructures/queue"
	"github.com/ontio/ontology-eventbus/internal/queue/mpsc"
)

type boundedMailboxQueue struct {
	userMailbox *rbqueue.RingBuffer
	dropping    bool
}

func (q *boundedMailboxQueue) Push(m interface{}) {
	if q.dropping {
		if q.userMailbox.Len() > 0 && q.userMailbox.Cap()-1 == q.userMailbox.Len() {
			q.userMailbox.Get()
		}
	}
	q.userMailbox.Put(m)
}

func (q *boundedMailboxQueue) Pop() interface{} {
	if q.userMailbox.Len() > 0 {
		m, _ := q.userMailbox.Get()
		return m
	}
	return nil
}

// Bounded returns a producer which creates an bounded mailbox of the specified size
func Bounded(size int, mailboxStats ...Statistics) Producer {
	return bounded(size, false, mailboxStats...)
}

// Bounded dropping returns a producer which creates an bounded mailbox of the specified size that drops front element on push
func BoundedDropping(size int, mailboxStats ...Statistics) Producer {
	return bounded(size, true, mailboxStats...)
}

func bounded(size int, dropping bool, mailboxStats ...Statistics) Producer {
	return func(invoker MessageInvoker, dispatcher Dispatcher) Inbound {
		q := &boundedMailboxQueue{
			userMailbox: rbqueue.NewRingBuffer(uint64(size)),
			dropping:    dropping,
		}
		return &defaultMailbox{
			systemMailbox: mpsc.New(),
			userMailbox:   q,
			invoker:       invoker,
			mailboxStats:  mailboxStats,
			dispatcher:    dispatcher,
		}
	}
}
