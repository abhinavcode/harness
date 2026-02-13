//  Copyright 2023 Harness, Inc.
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

package entitynode

import (
	"context"
)

var _ Service = (*noopService)(nil)

// noopService is a no-op implementation of Service.
// This is used in gitness standalone mode where entity-node linking is not available.
type noopService struct{}

// NewNoopService creates a new no-op Service.
func NewNoopService() Service {
	return &noopService{}
}

// LinkEntityToNodes does nothing in the no-op implementation.
func (n *noopService) LinkEntityToNodes(_ context.Context, _ EntityInput) error {
	// No-op: entity-node linking not available in gitness standalone
	return nil
}
