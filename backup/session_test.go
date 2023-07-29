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
	opts.ShouldUseSession = func(ctx context.Context) bool {
		return true
	}
	Enable(opts)
	os.Exit(m.Run())
}

func TestCurSession(t *testing.T) {
	BackupCtx(metainfo.WithPersistentValues(context.Background(), "a", "a", "b", "b"))
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want context.Context
	}{
		{
			name: "",
			args: args{
				ctx: metainfo.WithValue(metainfo.WithPersistentValue(context.Background(), "a", "aa"), "b", "bb"),
			},
			want: metainfo.WithPersistentValues(context.Background(), "a", "aa", "b", "b"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RecoverCtxOndemands(tt.args.ctx); got != nil {
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
