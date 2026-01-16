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

package pkg

import (
	"context"
	"fmt"

	apiauth "github.com/harness/gitness/app/api/auth"
	"github.com/harness/gitness/app/api/request"
	"github.com/harness/gitness/app/auth/authz"
	"github.com/harness/gitness/app/services/refcache"
	registrytypes "github.com/harness/gitness/registry/types"
	"github.com/harness/gitness/types"
	"github.com/harness/gitness/types/enum"

	"github.com/rs/zerolog/log"
)

// GetRegistryCheckAccess checks if the current user has permission to access the registry.
// Uses cached space from ArtifactInfo to avoid redundant DB lookups.
// The space must be cached in art.ParentSpace.
func GetRegistryCheckAccess(
	ctx context.Context,
	authorizer authz.Authorizer,
	art ArtifactInfo,
	reqPermissions ...enum.Permission,
) error {
	if art.ParentSpace == nil {
		return fmt.Errorf("parent space not cached in ArtifactInfo")
	}

	return checkRegistryAccess(ctx, authorizer, art.Registry, art.ParentSpace, reqPermissions...)
}

// GetRegistryCheckAccessWithFinder checks if the current user has permission to access the registry.
// Legacy version that fetches space from DB. Use GetRegistryCheckAccess when space is cached.
func GetRegistryCheckAccessWithFinder(
	ctx context.Context,
	authorizer authz.Authorizer,
	spaceFinder refcache.SpaceFinder,
	parentID int64,
	art ArtifactInfo,
	reqPermissions ...enum.Permission,
) error {
	space, err := spaceFinder.FindByID(ctx, parentID)
	if err != nil {
		return fmt.Errorf("failed to find parent by ref: %w", err)
	}

	return checkRegistryAccess(ctx, authorizer, art.Registry, space, reqPermissions...)
}

// checkRegistryAccess is the shared implementation for registry access checks.
func checkRegistryAccess(
	ctx context.Context,
	authorizer authz.Authorizer,
	registry registrytypes.Registry,
	space *types.SpaceCore,
	reqPermissions ...enum.Permission,
) error {
	session, _ := request.AuthSessionFrom(ctx)
	var permissionChecks []types.PermissionCheck

	for i := range reqPermissions {
		permissionCheck := types.PermissionCheck{
			Permission: reqPermissions[i],
			Scope:      types.Scope{SpacePath: space.Path},
			Resource: types.Resource{
				Type:       enum.ResourceTypeRegistry,
				Identifier: registry.Name,
			},
		}
		permissionChecks = append(permissionChecks, permissionCheck)
	}

	if err := apiauth.CheckRegistry(ctx, authorizer, session, permissionChecks...); err != nil {
		err = fmt.Errorf("registry access check failed: %w", err)
		log.Ctx(ctx).Error().Msgf("Error: %v", err)
		return err
	}

	return nil
}
