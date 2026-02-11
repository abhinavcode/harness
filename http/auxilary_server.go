// Copyright 2023 Harness, Inc.
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

package http

import "golang.org/x/sync/errgroup"

// ListenAndServeServer is an optional auxiliary server (e.g. metrics) that can be started.
type ListenAndServeServer interface {
	ListenAndServe() (*errgroup.Group, ShutdownFunction)
}

// NoOpListenAndServeServer is a no-op implementation of ListenAndServeServer.
type NoOpListenAndServeServer struct{}

// ListenAndServe returns (nil, nil); callers should skip starting/shutting down when both are nil.
func (NoOpListenAndServeServer) ListenAndServe() (*errgroup.Group, ShutdownFunction) {
	return nil, nil
}
