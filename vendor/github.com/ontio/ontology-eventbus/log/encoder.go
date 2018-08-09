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

import (
	"reflect"
	"time"
)

type Encoder interface {
	EncodeBool(key string, val bool)
	EncodeFloat64(key string, val float64)
	EncodeInt(key string, val int)
	EncodeInt64(key string, val int64)
	EncodeDuration(key string, val time.Duration)
	EncodeUint(key string, val uint)
	EncodeUint64(key string, val uint64)
	EncodeString(key string, val string)
	EncodeObject(key string, val interface{})
	EncodeType(key string, val reflect.Type)
}
