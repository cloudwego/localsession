/**
 * Copyright 2025 ByteDance Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
//
// Copyright 2016 Huan Du. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

#ifdef GOARCH_arm
#define LR R14
#endif

#ifdef GOARCH_amd64
#define get_tls(r) MOVQ TLS, r
#define g(r) 0(r)(TLS * 1)
#endif

#ifdef GOARCH_amd64p32
#define get_tls(r) MOVL TLS, r
#define g(r) 0(r)(TLS * 1)
#endif

#ifdef GOARCH_386
#define get_tls(r) MOVL TLS, r
#define g(r) 0(r)(TLS * 1)
#endif
