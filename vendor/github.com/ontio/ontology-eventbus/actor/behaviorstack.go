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

type behaviorStack []ActorFunc

func (b *behaviorStack) Clear() {
	if len(*b) == 0 {
		return
	}

	for i := range *b {
		(*b)[i] = nil
	}
	*b = (*b)[:0]
}

func (b *behaviorStack) Peek() (v ActorFunc, ok bool) {
	l := b.Len()
	if l > 0 {
		ok = true
		v = (*b)[l-1]
	}
	return
}

func (b *behaviorStack) Push(v ActorFunc) {
	*b = append(*b, v)
}

func (b *behaviorStack) Pop() (v ActorFunc, ok bool) {
	l := b.Len()
	if l > 0 {
		l--
		ok = true
		v = (*b)[l]
		(*b)[l] = nil
		*b = (*b)[:l]
	}
	return
}

func (b *behaviorStack) Len() int {
	return len(*b)
}
