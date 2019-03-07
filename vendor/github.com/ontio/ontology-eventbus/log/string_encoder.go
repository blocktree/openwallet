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
	"fmt"
	"io"
	"os"
	"reflect"
	"time"
)

type ioLogger struct {
	c   chan Event
	out io.Writer
	buf []byte
}

var (
	sub *Subscription
)

func (l *ioLogger) listenEvent() {
	for true {
		e := <-l.c
		l.writeEvent(e)
	}
}

func fileOpen(path string) (*os.File, error) {
	if fi, err := os.Stat(path); err == nil {
		if !fi.IsDir() {
			return nil, fmt.Errorf("open %s: not a directory", path)
		}
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0766); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	var currenttime string = time.Now().Format("2006-01-02_15.04.05")

	logfile, err := os.OpenFile(path+currenttime+"_LOG.log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return logfile, nil
}

// Cheap integer to fixed-width decimal ASCII.  Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

func (l *ioLogger) formatHeader(buf *[]byte, prefix string, t time.Time) {
	t = t.UTC()
	// Y/M/D
	year, month, day := t.Date()
	itoa(buf, year, 4)
	*buf = append(*buf, '/')
	itoa(buf, int(month), 2)
	*buf = append(*buf, '/')
	itoa(buf, day, 2)
	*buf = append(*buf, ' ')

	// H/M/S
	hour, min, sec := t.Clock()
	itoa(buf, hour, 2)
	*buf = append(*buf, ':')
	itoa(buf, min, 2)
	*buf = append(*buf, ':')
	itoa(buf, sec, 2)

	// no microseconds
	//*buf = append(*buf, '.')
	//itoa(buf, t.Nanosecond()/1e3, 6)

	*buf = append(*buf, ' ')
	if len(prefix) > 0 {
		*buf = append(*buf, prefix...)
		*buf = append(*buf, ' ')
	}
}

func (l *ioLogger) writeEvent(e Event) {
	l.buf = l.buf[:0]
	l.formatHeader(&l.buf, e.Prefix, e.Time)
	l.out.Write(l.buf)
	if len(e.Message) > 0 {
		l.out.Write([]byte(e.Message))
		l.out.Write([]byte{' '})
	}

	wr := ioEncoder{l.out}
	for _, f := range e.Context {
		f.Encode(wr)
		l.out.Write([]byte{' '})
	}
	for _, f := range e.Fields {
		f.Encode(wr)
		l.out.Write([]byte{' '})
	}
	wr.Write([]byte{'\n'})
}

type ioEncoder struct {
	io.Writer
}

func (e ioEncoder) EncodeBool(key string, val bool) {
	fmt.Fprintf(e, "%s=%t", key, val)
}

func (e ioEncoder) EncodeFloat64(key string, val float64) {
	fmt.Fprintf(e, "%s=%f", key, val)
}

func (e ioEncoder) EncodeInt(key string, val int) {
	fmt.Fprintf(e, "%s=%d", key, val)
}

func (e ioEncoder) EncodeInt64(key string, val int64) {
	fmt.Fprintf(e, "%s=%d", key, val)
}

func (e ioEncoder) EncodeDuration(key string, val time.Duration) {
	fmt.Fprintf(e, "%s=%s", key, val)
}

func (e ioEncoder) EncodeUint(key string, val uint) {
	fmt.Fprintf(e, "%s=%d", key, val)
}

func (e ioEncoder) EncodeUint64(key string, val uint64) {
	fmt.Fprintf(e, "%s=%d", key, val)
}

func (e ioEncoder) EncodeString(key string, val string) {
	fmt.Fprintf(e, "%s=%q", key, val)
}

func (e ioEncoder) EncodeObject(key string, val interface{}) {
	fmt.Fprintf(e, "%s=%q", key, val)
}

func (e ioEncoder) EncodeType(key string, val reflect.Type) {
	fmt.Fprintf(e, "%s=%v", key, val)
}
