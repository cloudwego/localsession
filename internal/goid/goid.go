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
//
// Copyright 2016 Huan Du. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package goid

import (
	"unsafe"

	"github.com/modern-go/reflect2"
)

// offset for go1.4
var goidOffset uintptr = 128

func init() {
	gType := reflect2.TypeByName("runtime.g").(reflect2.StructType)
	if gType == nil {
		panic("failed to get runtime.g type")
	}
	goidField := gType.FieldByName("goid")
	goidOffset = goidField.Offset()
}

// GoID returns the goroutine id of current goroutine
func GoID() int64 {
	p_goid := (*int64)(unsafe.Add(unsafe.Pointer(getg()), goidOffset))
	return *p_goid
}

func getg() uintptr
