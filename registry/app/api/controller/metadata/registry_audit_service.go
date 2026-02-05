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

package metadata

import (
	"context"

	"github.com/harness/gitness/audit"
	"github.com/harness/gitness/registry/types"
	types2 "github.com/harness/gitness/types"
)

type RegistryAuditService interface {
	Log(
		ctx context.Context,
		action audit.Action,
		oldRegistry *types.Registry,
		newRegistry *types.Registry,
		oldUpstreamProxy *types.UpstreamProxy,
		newUpstreamProxy *types.UpstreamProxy,
		principal types2.Principal,
		parentRef string,
		resourceType audit.ResourceType,
	)
}

type NoOpRegistryAuditService struct{}

func NewNoOpRegistryAuditService() RegistryAuditService {
	return &NoOpRegistryAuditService{}
}

func (s *NoOpRegistryAuditService) Log(
	_ context.Context,
	_ audit.Action,
	_ *types.Registry,
	_ *types.Registry,
	_ *types.UpstreamProxy,
	_ *types.UpstreamProxy,
	_ types2.Principal,
	_ string,
	_ audit.ResourceType,
) {
	// NoOp implementation - does nothing
}
