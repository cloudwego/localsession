//go:build !go1.24

// Copyright 2023 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package localsession

import (
	"reflect"
	"strconv"
	"unsafe"
	_ "unsafe"
)

//go:linkname getPproLabel runtime/pprof.runtime_getProfLabel
func getPproLabel() unsafe.Pointer

//go:linkname setPproLabel runtime/pprof.runtime_setProfLabel
func setPproLabel(m unsafe.Pointer)

// WARNING: labelMap must be aligned with runtime.labelMap
type labelMap map[string]string

func getSessionID() (SessionID, bool) {
	m := getPproLabel()
	if m == nil {
		return 0, false
	}

	mv := reflect.NewAt(reflect.TypeOf(labelMap{}), m).Elem().Interface()
	v, ok := mv.(labelMap)[Pprof_Label_Session_ID]
	if !ok {
		return 0, false
	}

	id, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, false
	}
	return SessionID(id), true
}

func clearSessionID() {
	m := getPproLabel()
	if m == nil {
		return
	}

	mv := *(reflect.NewAt(reflect.TypeOf(labelMap{}), m).Interface().(*labelMap))
	if _, ok := mv[Pprof_Label_Session_ID]; !ok {
		return
	}
	var n = make(labelMap, len(mv)-1)
	for k, v := range mv {
		if k == Pprof_Label_Session_ID {
			continue
		}
		n[k] = v
	}

	setPproLabel(unsafe.Pointer(&n))
}
