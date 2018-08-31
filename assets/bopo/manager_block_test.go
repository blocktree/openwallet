/*
 * Copyright 2018 The OpenWallet Authors
 * This file is part of the OpenWallet library.
 *
 * The OpenWallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The OpenWallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */
package bopo

import "testing"

func TestGetBlockChainInfo(t *testing.T) {
	if b, err := tw.GetBlockChainInfo(); err != nil {
		t.Errorf("GetBlockChainInfo failed unexpected error: %v\n", err)
	} else {
		t.Logf("TestGetBlockChainInfo: \n\t%+v\n", b)
	}
}

func TestGetBlockContent(t *testing.T) {
	if content, err := tw.GetBlockContent(332431); err != nil {
		t.Errorf("GetBlockContent failed unexpected error: %v\n", err)
	} else {
		t.Logf("GetBlockContent: \n\t%+v \n", content)
	}
}

func TestGetBlockHash(t *testing.T) {
	var currentHash string

	if block, err := tw.GetBlockContent(332431); err != nil {
		t.Errorf("TestGetBlockHash Failed: %v\n", err)
	} else {
		currentHash = block.Previousblockhash
		t.Logf("TestGetBlockHash: RealHash[%+v] \n", currentHash)
	}

	if hash, err := tw.GetBlockHash(332430); err != nil {
		t.Errorf("TestGetBlockHash Failed: %v\n", err)
	} else {
		if hash != currentHash {
			t.Errorf("TestGetBlockHash Failed: '!=' \n\tGetHash[%s]\n\tReaHash[%s] \n",
				hash, currentHash)
		} else {
			t.Logf("TestGetBlockHash Success: \n\t%+v \n", hash)
		}
	}
}

func TestGetBlockPayload(t *testing.T) {
	payload := "CpwDCAESERIPZ2FtZXBhaWNvcmVfdjAxGuYCCglVU0VSX0ZVTkQKvAFDaUkxV21GUVdHWktZVXhPY2tkdVdIVjVXSFZ1UmtVMGVFdDRZV3RGZW1kVVNWcFJFaHhHY21rZ1FYVm5JREUzSURFeE9qQTVPakkxSUVOVFZDQXlNREU0SWtnSUFSSkVDaURpaTdoUTZGSGZFc2NXV0p4a1YzZ2NLWnBoaVhudDMrMnZiSjNoVXFrTDVoSWc3dDh3RDNTamhNb0ZPaUliNTRVVVAwYUxwTndyRVhpZmxMRGd4ZFdqUzlzPQo4Q2lZSUFSSWlOVnBHVmxaUU5EZFNaalZxTFdzM1RHOXBVbU5PYjNwc1l6aGtlVzVpVUZsdVp3PT0KYENrUUtJR253RTFNcHFyYUNBdmcybXhueHNuWDVUK2lOYkMwVEhlallZMVFYQi9lWEVpQUQ0RlpRMjVHM09qeFJDUFBKRERkY0xRWEEvMTgweWtVSFc0VEIwRnMrVmc9PUIMUGFpQWRtaW5Sb2xlQg5QYWlBZG1pblJlZ2lvbg=="
	if payloadSpec, err := tw.GetBlockPayload(payload); err != nil {
		t.Errorf("GetBlockPayload failed unexpected error: %v\n", err)
	} else {
		t.Logf("GetBlockPayload: \n\t%+v \n", payloadSpec)
	}
	payload = "CtsCCAESERIPZ2FtZXBhaWNvcmVfdjAxGqUCChFVU0VSX1JFR1BVQkxJQ0tFWQpIQ2lJMVV6aFNka1ptV2xWaldVMUNXbXcwVGs5R1h6WmtUblExWmpSQ01VWkljbE4zRWhBYWNWOFY2OHo0UHRwZ0lZMUp4SDF5CmRDa2dJQVJKRUNpRG5MdEd0TTlzbml2UlpMRitoR3djMkhXR0VSRkVPYXJOcHYrRlNlZklvd3hJZ1dkQ2ZHSm93OFNCcVNUeUJxSm9aaDhqc1NrY0pzNktrb0pEYkpubUowVUk9CmBDa1FLSU9Zako0K3lhZFAwT2lKM1VWSmNOa0M0MmRnS1NZTTNuUzZBZGJEY253QnRFaUE0dTZlMEQ0TFlWblRWRm8wVENnSnlwZytCVTBKaDlyTDZ5TjFwaGh6RnVRPT1CDFBhaUFkbWluUm9sZUIOUGFpQWRtaW5SZWdpb24="
	if _, err := tw.GetBlockPayload(payload); err == nil {
		t.Errorf("GetBlockPayload failed unexpected error: %v\n", err)
	}
}
