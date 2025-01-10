// Copyright 2025 CloudWeGo Authors
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
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	_ "unsafe"
)

func TestGoID(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			testGoID(t)
			wg.Done()
		}()
	}
	wg.Wait()
}

func testGoID(t *testing.T) {
	id := GoID()
	lines := strings.Split(stackTrace(), "\n")
	for i, line := range lines {
		if !strings.HasPrefix(line, fmt.Sprintf("goroutine %d ", id)) {
			continue
		}
		if i+1 == len(lines) {
			break
		}
		if !strings.Contains(lines[i+1], ".stackTrace") {
			t.Errorf("there are goroutine id %d but it is not me: %s", id, lines[i+1])
		}
		return
	}
	t.Errorf("there are no goroutine %d", id)
}

func stackTrace() string {
	var n int
	for n = 4096; n < 16777216; n *= 2 {
		buf := make([]byte, n)
		ret := runtime.Stack(buf, true)
		if ret != n {
			return string(buf[:ret])
		}
	}
	panic(n)
}
