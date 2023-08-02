/*
 * Copyright 2023 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package backup

import (
	"context"
	"os"
	"testing"

	"github.com/bytedance/gopkg/cloud/metainfo"
)

func TestMain(m *testing.M) {
	opts := DefaultOptions()
	opts.EnableImplicitlyTransmitAsync = true
	opts.Enable = true
	Init(opts)
	os.Exit(m.Run())
}

type CtxKeyTestType struct{}

var CtxKeyTest1 CtxKeyTestType

func TestRecoverCtxOndemands(t *testing.T) {
	ctx := context.WithValue(context.Background(), CtxKeyTest1, "c")
	BackupCtx(metainfo.WithPersistentValues(ctx, "a", "a", "b", "b"))

	handler := BackupHandler(func(prev, cur context.Context) (ctx context.Context, backup bool) {
		if v := cur.Value(CtxKeyTest1); v == nil {
			v = prev.Value(CtxKeyTest1)
			if v != nil {
				ctx = context.WithValue(cur, CtxKeyTest1, v)
			} else {
				ctx = cur
			}
			return ctx, true
		}
		return cur, false
	})
	type args struct {
		ctx     context.Context
		handler BackupHandler
	}
	tests := []struct {
		name string
		args args
		want context.Context
	}{
		{
			name: "triggered",
			args: args{
				ctx:     metainfo.WithValue(metainfo.WithPersistentValue(context.Background(), "a", "aa"), "b", "bb"),
				handler: handler,
			},
			want: metainfo.WithPersistentValues(ctx, "a", "aa", "b", "b"),
		},
		{
			name: "not triggered",
			args: args{
				ctx:     metainfo.WithValue(metainfo.WithPersistentValue(ctx, "a", "aa"), "b", "bb"),
				handler: handler,
			},
			want: metainfo.WithValue(metainfo.WithPersistentValue(ctx, "a", "aa"), "b", "bb"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RecoverCtxOnDemands(tt.args.ctx, tt.args.handler); got != nil {
				if v := got.Value(CtxKeyTest1); v == nil {
					t.Errorf("not got CtxKeyTest1")
				}
				a, _ := metainfo.GetPersistentValue(got, "a")
				ae, _ := metainfo.GetPersistentValue(tt.want, "a")
				if a != ae {
					t.Errorf("CurSession() = %v, want %v", a, ae)
				}
				b, _ := metainfo.GetPersistentValue(got, "b")
				be, _ := metainfo.GetPersistentValue(tt.want, "b")
				if b != be {
					t.Errorf("CurSession() = %v, want %v", b, be)
				}
			} else {
				t.Fatal("no got")
			}
		})
	}
}
