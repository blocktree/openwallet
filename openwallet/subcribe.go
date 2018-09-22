/*
 * Copyright 2018 The OpenSubcribe Authors
 * This file is part of the OpenSubcribe library.
 *
 * The OpenSubcribe library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The OpenSubcribe library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package openwallet


type Subcribe struct {
	AppID        string `json:"appID"`
	Type        int     `json:"type"`
	Symbol      string  `json:"symbol"`
}
