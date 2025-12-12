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
	"fmt"
	"math"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const N = 10

func TestMain(m *testing.M) {
	InitDefaultManager(DefaultManagerOptions())
	m.Run()
}

func TestResetDefaultManager(t *testing.T) {
	old := defaultManagerObj

	t.Run("arg", func(t *testing.T) {
		defaultManagerOnce = sync.Once{}
		exp := ManagerOptions{
			ShardNumber:                   1,
			EnableImplicitlyTransmitAsync: true,
			GCInterval:                    time.Second * 2,
		}
		InitDefaultManager(exp)
		act := defaultManagerObj.Options()
		require.Equal(t, exp, act)
	})

	t.Run("arg", func(t *testing.T) {
		defaultManagerOnce = sync.Once{}
		env := `true,10,10s`
		os.Setenv(SESSION_CONFIG_KEY, env)
		exp := DefaultManagerOptions()
		InitDefaultManager(exp)
		act := defaultManagerObj.Options()
		exp.ShardNumber = 10
		exp.EnableImplicitlyTransmitAsync = true
		exp.GCInterval = time.Second * 10
		require.Equal(t, exp, act)

		defaultManagerOnce = sync.Once{}
		env = `,1000`
		os.Setenv(SESSION_CONFIG_KEY, env)
		exp = DefaultManagerOptions()
		InitDefaultManager(exp)
		act = defaultManagerObj.Options()
		exp.ShardNumber = 1000
		require.Equal(t, exp, act)

		defaultManagerOnce = sync.Once{}
		env = `,1,2s`
		os.Setenv(SESSION_CONFIG_KEY, env)
		exp = DefaultManagerOptions()
		InitDefaultManager(exp)
		act = defaultManagerObj.Options()
		exp.ShardNumber = 1
		exp.GCInterval = time.Second * 2
		require.Equal(t, exp, act)

		defaultManagerOnce = sync.Once{}
		env = `true,,2s`
		os.Setenv(SESSION_CONFIG_KEY, env)
		exp = DefaultManagerOptions()
		InitDefaultManager(exp)
		act = defaultManagerObj.Options()
		exp.EnableImplicitlyTransmitAsync = true
		exp.GCInterval = time.Second * 2
		require.Equal(t, exp, act)
	})

	defaultManagerObj = old
	defaultManagerOnce = sync.Once{}
}

func TestSessionTimeout(t *testing.T) {
	s := NewSessionCtxWithTimeout(context.Background(), time.Second)
	ss := s.WithValue(1, 2)
	m := NewSessionMapWithTimeout(map[interface{}]interface{}{}, time.Second)
	mm := m.WithValue(1, 2)
	time.Sleep(time.Second * 2)
	require.False(t, ss.IsValid())
	require.False(t, mm.IsValid())
}

func TestSessionCtx(t *testing.T) {
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
		val := mustCurSession().Get(key) // val exists
		require.Equal(t, v, val)
		// doSomething....

		// set specific key-value under current session
		// NOTICE: current session won't change here
		next := mustCurSession().WithValue(key2, v2)
		val2 := mustCurSession().Get(key2) // val2 == nil
		require.Nil(t, val2)

		// pass both parent session and new session to sub goroutine
		GoSession(next, func() {
			// read specific key under current session
			val := mustCurSession().Get(key) // val exists
			require.Equal(t, v, val)

			val2 := mustCurSession().Get(key2) // val2 exists
			require.Equal(t, v2, val2)
			// doSomething....

			sig2 <- struct{}{}

			<-sig
			require.False(t, mustCurSession().IsValid()) // current session is invalid

			println("g2 done")
			sig2 <- struct{}{}
		})

		Go(func() {
			// read specific key under current session
			val := mustCurSession().Get(key) // val exists
			require.Equal(t, v, val)

			val2 := mustCurSession().Get(key2) // val2 == nil
			require.Nil(t, val2)
			// doSomething....

			sig2 <- struct{}{}

			<-sig
			require.False(t, mustCurSession().IsValid()) // current session is invalid

			println("g3 done")
			sig2 <- struct{}{}
		})

		BindSession(next)
		val2 = mustCurSession().Get(key2) // val2 exists
		require.Equal(t, v2, val2)

		sig2 <- struct{}{}

		<-sig
		require.False(t, next.IsValid()) // next is invalid

		println("g1 done")
		sig2 <- struct{}{}
	})

	<-sig2
	<-sig2
	<-sig2

	val2 := mustCurSession().Get(key2) // val2 == nil
	require.Nil(t, val2)

	// initiatively ends the session，
	// then all the inherited session (including next) will be disabled
	session.Disable()
	close(sig)

	require.False(t, start.IsValid()) // start is invalid

	<-sig2
	<-sig2
	<-sig2
	println("g0 done")

	UnbindSession()
}

func mustCurSession() Session {
	s, _ := CurSession()
	return s
}

func TestSessionMap(t *testing.T) {
	var key, v = "a", "b"
	var key2, v2 = "c", "d"
	var sig = make(chan struct{})
	var sig2 = make(chan struct{})

	// initialize new session with context
	var session = NewSessionMap(map[interface{}]interface{}{}) // implementation...

	// set specific key-value and update session
	start := session.WithValue(key, v)

	// set current session
	BindSession(start)

	// pass to new goroutine...
	Go(func() {
		// read specific key under current session
		val := mustCurSession().Get(key) // val exists
		require.Equal(t, v, val)
		// doSomething....

		// set specific key-value under current session
		// NOTICE: current session won't change here
		next := mustCurSession().WithValue(key2, v2)
		val2 := mustCurSession().Get(key2) // val2 exist
		require.Equal(t, v2, val2)

		// pass both parent session and new session to sub goroutine
		GoSession(next, func() {
			// read specific key under current session
			val := mustCurSession().Get(key) // val exists
			require.Equal(t, v, val)

			val2 := mustCurSession().Get(key2) // val2 exists
			require.Equal(t, v2, val2)
			// doSomething....

			sig2 <- struct{}{}

			<-sig
			require.False(t, mustCurSession().IsValid()) // current session is invalid

			println("g2 done")
			sig2 <- struct{}{}
		})

		Go(func() {
			// read specific key under current session
			val := mustCurSession().Get(key) // val exists
			require.Equal(t, v, val)

			val2 := mustCurSession().Get(key2) // val2 exist
			require.Equal(t, v2, val2)
			// doSomething....

			sig2 <- struct{}{}

			<-sig
			require.False(t, mustCurSession().IsValid()) // current session is invalid

			println("g3 done")
			sig2 <- struct{}{}
		})

		BindSession(next)
		val2 = mustCurSession().Get(key2) // val2 exists
		require.Equal(t, v2, val2)

		sig2 <- struct{}{}

		<-sig
		require.False(t, next.IsValid()) // next is invalid

		println("g1 done")
		sig2 <- struct{}{}
	})

	<-sig2
	<-sig2
	<-sig2

	val2 := mustCurSession().Get(key2) // val2 exists
	require.Equal(t, v2, val2)

	// initiatively ends the session，
	// then all the inherited session (including next) will be disabled
	session.Disable()
	close(sig)

	require.False(t, start.IsValid()) // start is invalid

	<-sig2
	<-sig2
	<-sig2
	println("g0 done")

	UnbindSession()
}

func TestSessionManager_GC(t *testing.T) {
	inter := time.Second * 2
	sd := 10
	manager := NewSessionManager(ManagerOptions{
		ShardNumber: sd,
		GCInterval:  inter,
	})

	var N = 1000
	for i := 0; i < N; i++ {
		m := map[interface{}]interface{}{}
		s := NewSessionMap(m)
		manager.BindSession(SessionID(i), s)
		if i%2 == 1 {
			s.Disable()
		}
	}
	for _, shard := range manager.shards {
		shard.lock.Lock()
		l := len(shard.m)
		shard.lock.Unlock()
		require.Equal(t, N/sd, l)
	}
	time.Sleep(inter + inter>>1)
	sum := 0
	for _, shard := range manager.shards {
		shard.lock.Lock()
		l := len(shard.m)
		shard.lock.Unlock()
		sum += l
	}
	require.Equal(t, N/2, sum)
}

func TestRace(t *testing.T) {
	manager := NewSessionManager(ManagerOptions{
		ShardNumber: 1,
		GCInterval:  time.Second,
	})
	var N = 1000
	var start sync.RWMutex
	start.Lock()
	wg := sync.WaitGroup{}
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s := NewSessionMap(map[interface{}]interface{}{}).WithValue("a", "b")
			start.RLock()
			manager.BindSession(SessionID(i), s)
			ss, ok := manager.GetSession(SessionID(i))
			if !ok || ss.Get("a") != "b" {
				t.Fatal("not equal")
			}
		}(i)
	}
	start.Unlock()
	wg.Wait()
}

func BenchmarkSessionManager_CurSession(b *testing.B) {
	s := NewSessionCtx(context.Background())

	b.Run("sync", func(b *testing.B) {
		BindSession(s)
		for i := 0; i < b.N; i++ {
			_ = mustCurSession()
		}
		UnbindSession()
	})

	b.Run("parallel", func(b *testing.B) {
		b.RunParallel(func(p *testing.PB) {
			BindSession(s)
			for p.Next() {
				_ = mustCurSession()
			}
			UnbindSession()
		})
	})
}

func BenchmarkSessionManager_BindSession(b *testing.B) {
	s := NewSessionCtx(context.Background())

	b.Run("sync", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			BindSession(s)
		}
	})

	b.Run("parallel", func(b *testing.B) {
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				BindSession(s)
			}
		})
	})
}

func BenchmarkSessionCtx_WithValue(b *testing.B) {
	s := NewSessionCtx(context.Background())
	var ss Session = s
	for i := 0; i < N; i++ {
		ss = ss.WithValue(i, i)
	}

	b.Run("sync", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ss.WithValue(N/2, -1)
		}
	})

	b.Run("parallel", func(b *testing.B) {
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				_ = ss.WithValue(N/2, -1)
			}
		})
	})
}

func BenchmarkSessionCtx_Get(b *testing.B) {
	s := NewSessionCtx(context.Background())
	var ss Session = s
	for i := 0; i < N; i++ {
		ss = ss.WithValue(i, i)
	}

	b.Run("sync", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ss.Get(N / 2)
		}
	})

	b.Run("parallel", func(b *testing.B) {
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				_ = ss.Get(N / 2)
			}
		})
	})
}

func BenchmarkSessionMap_WithValue(b *testing.B) {
	s := NewSessionMap(map[interface{}]interface{}{})
	var ss Session = s
	for i := 0; i < N; i++ {
		ss = ss.WithValue(i, i)
	}

	b.Run("sync", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ss.WithValue(N/2, -1)
		}
	})

	b.Run("parallel", func(b *testing.B) {
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				_ = ss.WithValue(N/2, -1)
			}
		})
	})
}

func BenchmarkSessionMap_Get(b *testing.B) {
	s := NewSessionMap(map[interface{}]interface{}{})
	var ss Session = s
	for i := 0; i < N; i++ {
		ss = ss.WithValue(i, i)
	}

	b.Run("sync", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ss.Get(N / 2)
		}
	})

	b.Run("parallel", func(b *testing.B) {
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				_ = ss.Get(N / 2)
			}
		})
	})
}

func BenchmarkGLS_Get(b *testing.B) {
	s := NewSessionCtx(context.Background())
	var ss Session = s
	for i := 0; i < N; i++ {
		ss = ss.WithValue(i, i)
	}

	b.Run("sync", func(b *testing.B) {
		BindSession(ss)
		for i := 0; i < b.N; i++ {
			_ = mustCurSession().Get(N / 2)
		}
		UnbindSession()
	})

	b.Run("parallel", func(b *testing.B) {
		b.RunParallel(func(p *testing.PB) {
			BindSession(ss)
			for p.Next() {
				_ = mustCurSession().Get(N / 2)
			}
			UnbindSession()
		})
	})
}

func BenchmarkGLS_Set(b *testing.B) {
	s := NewSessionCtx(context.Background())
	var ss Session = s

	for i := 0; i < N; i++ {
		ss = ss.WithValue(i, i)
	}

	b.Run("sync", func(b *testing.B) {
		BindSession(ss)
		for i := 0; i < b.N; i++ {
			BindSession(mustCurSession().WithValue(N/2, -1))
		}
		UnbindSession()
	})

	b.Run("parallel", func(b *testing.B) {
		b.RunParallel(func(p *testing.PB) {
			BindSession(ss)
			for p.Next() {
				BindSession(mustCurSession().WithValue(N/2, -1))
			}
			UnbindSession()
		})
	})
}

func emitLoops(m *SessionManager, ctx context.Context, N int, s *stat) {
	for i := 0; i < N; i++ {
		go func() {
			for {
				if ctx.Err() != nil {
					return
				}
				start := time.Now()
				session := NewSessionCtx(ctx)
				ss := session.WithValue("a", "b")
				m.BindSession(SessionID(goID()), ss)
				sss, _ := m.GetSession(SessionID(goID()))
				if val := sss.Get("a"); val != "b" {
					panic(fmt.Sprintf("unexpected val: %#v", val))
				}
				m.UnbindSession(SessionID(goID()))
				cost := time.Now().Sub(start)
				s.Update(cost)
				for a := 0; a < 10; a++ {
					time.Sleep(time.Microsecond * 50)
					for b := 0; b < 100000; b++ {
						_ = b
					}
				}
			}
		}()
	}
}

func BenchmarkLoops(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for b := 0; b < 100000; b++ {
			_ = b
		}
	}
}

type stat struct {
	max   time.Duration
	min   time.Duration
	sum   time.Duration
	count int

	mux sync.RWMutex
}

func (st *stat) Update(cost time.Duration) {
	st.mux.Lock()
	defer st.mux.Unlock()
	if cost > st.max {
		st.max = cost
	} else if cost < st.min {
		st.min = cost
	}
	st.count++
	st.sum += cost
	return
}

func (st *stat) String() string {
	st.mux.RLock()
	defer st.mux.RUnlock()
	return fmt.Sprintf("min:%dns, max:%dns, avg:%dns", st.min, st.max, st.sum/time.Duration(st.count))
}

func TestRealBizGLS(t *testing.T) {
	var runner = func(N int) {
		m := NewSessionManager(ManagerOptions{
			ShardNumber: 100,
			GCInterval:  time.Second,
		})
		s := &stat{
			min: time.Duration(math.MaxInt64),
		}
		ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
		emitLoops(&m, ctx, N, s)
		go func(ctx context.Context) {
			tt := time.NewTicker(time.Second)
			for {
				select {
				case <-tt.C:
					{
						println(s.String())
					}
				case <-ctx.Done():
					return
				}

			}
		}(ctx)
		<-ctx.Done()
	}

	t.Run("10", func(t *testing.T) {
		runner(10)
	})
	t.Run("100", func(t *testing.T) {
		runner(100)
	})
}
