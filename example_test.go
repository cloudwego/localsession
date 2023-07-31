// Copyright 2023 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package localsession

import (
	"context"
	"runtime/pprof"
	"sync"
	"time"
)

func ASSERT(v bool) {
	if !v {
		panic("not true!")
	}
}

func GetCurSession() Session {
	s, ok := CurSession()
	if !ok {
		panic("can't get current session!")
	}
	return s
}

func ExampleSessionCtx_EnableImplicitlyTransmitAsync() {
	// EnableImplicitlyTransmitAsync must be true
	InitDefaultManager(ManagerOptions{
		ShardNumber:                   10,
		EnableImplicitlyTransmitAsync: true,
		GCInterval:                    time.Hour,
	})

	// WARNING: pprof.Do() must be called before BindSession(),
	// otherwise transparently transmitting session will be dysfunctional
	labels := pprof.Labels("c", "d")
	pprof.Do(context.Background(), labels, func(ctx context.Context) {})

	s := NewSessionMap(map[interface{}]interface{}{
		"a": "b",
	})
	BindSession(s)

	// WARNING: pprof.Do() must be called before BindSession(),
	// otherwise transparently transmitting session will be dysfunctional
	// labels := pprof.Labels("c", "d")
	// pprof.Do(context.Background(), labels, func(ctx context.Context){})

	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		ASSERT("b" == mustCurSession().Get("a"))

		go func() {
			defer wg.Done()
			ASSERT("b" == mustCurSession().Get("a"))
		}()

		ASSERT("b" == mustCurSession().Get("a"))
		UnbindSession()
		ASSERT(nil == mustCurSession())

		go func() {
			defer wg.Done()
			ASSERT(nil == mustCurSession())
		}()

	}()
	wg.Wait()
}

func ExampleSessionCtx() {
	// initialize default manager first
	InitDefaultManager(DefaultManagerOptions())

	var ctx = context.Background()
	var key, v = "a", "b"
	var key2, v2 = "c", "d"
	var sig = make(chan struct{})
	var sig2 = make(chan struct{})

	// initialize new session with context
	var session = NewSessionCtx(ctx) // implementation...

	// set specific key-value and update session
	start := session.WithValue(key, v)

	// set current session
	BindSession(start)

	// pass to new goroutine...
	Go(func() {
		// read specific key under current session
		val := GetCurSession().Get(key) // val exists
		ASSERT(val == v)
		// doSomething....

		// set specific key-value under current session
		// NOTICE: current session won't change here
		next := GetCurSession().WithValue(key2, v2)
		val2 := GetCurSession().Get(key2) // val2 == nil
		ASSERT(val2 == nil)

		// pass both parent session and new session to sub goroutine
		GoSession(next, func() {
			// read specific key under current session
			val := GetCurSession().Get(key) // val exists
			ASSERT(val == v)

			val2 := GetCurSession().Get(key2) // val2 exists
			ASSERT(val2 == v2)
			// doSomething....

			sig2 <- struct{}{}

			<-sig
			ASSERT(GetCurSession().IsValid() == false) // current session is invalid

			println("g2 done")
			sig2 <- struct{}{}
		})

		Go(func() {
			// read specific key under current session
			val := GetCurSession().Get(key) // val exists
			ASSERT(v == val)

			val2 := GetCurSession().Get(key2) // val2 == nil
			ASSERT(val2 == nil)
			// doSomething....

			sig2 <- struct{}{}

			<-sig
			ASSERT(GetCurSession().IsValid() == false) // current session is invalid

			println("g3 done")
			sig2 <- struct{}{}
		})

		BindSession(next)
		val2 = GetCurSession().Get(key2) // val2 exists
		ASSERT(v2 == val2)

		sig2 <- struct{}{}

		<-sig
		ASSERT(next.IsValid() == false) // next is invalid

		println("g1 done")
		sig2 <- struct{}{}
	})

	<-sig2
	<-sig2
	<-sig2

	val2 := GetCurSession().Get(key2) // val2 == nil
	ASSERT(val2 == nil)

	// initiatively ends the session，
	// then all the inherited session (including next) will be disabled
	session.Disable()
	close(sig)

	ASSERT(start.IsValid() == false) // start is invalid

	<-sig2
	<-sig2
	<-sig2
	println("g0 done")

	UnbindSession()
}
