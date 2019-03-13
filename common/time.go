/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package common

import (
	"time"
	"fmt"
)

func ToISO8601(t ...time.Time) string {
	var tt time.Time
	if len(t) > 0 {
		tt = t[0]
	} else {
		tt = time.Now().UTC()
	}

	var tz string
	name, offset := tt.Zone()
	if name == "UTC" {
		tz = "Z"
	} else {
		tz = fmt.Sprintf("%03d00", offset/3600)
	}
	return fmt.Sprintf("%04d-%02d-%02dT%02d-%02d-%02d.%09d%s", tt.Year(), tt.Month(), tt.Day(), tt.Hour(), tt.Minute(), tt.Second(), tt.Nanosecond(), tz)
}

//timeFormat 格式化时间
//@param format 在go中，为2006-01-02 15:04:05
func TimeFormat(format string, t ...time.Time) string {
	var tt time.Time
	if len(t) > 0 {
		tt = t[0]
	} else {
		tt = time.Now()
	}
	s := tt.Format(format)
	return s
}