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
// Self-healing: uses cached space if available in art.ParentSpace, otherwise fetches from DB.
// This ensures correctness regardless of whether the handler cached spaces or not.
func GetRegistryCheckAccess(
	ctx context.Context,
	authorizer authz.Authorizer,
	spaceFinder refcache.SpaceFinder,
	art ArtifactInfo,
	reqPermissions ...enum.Permission,
) error {
	// Use cached space if available (optimization), otherwise fetch
	parentSpace := art.ParentSpace
	if parentSpace == nil {
		var err error
		parentSpace, err = spaceFinder.FindByID(ctx, art.ParentID)
		if err != nil {
			return fmt.Errorf("failed to find parent space: %w", err)
		}
	}

	return checkRegistryAccess(ctx, authorizer, art.Registry, parentSpace, reqPermissions...)
}

// GetRegistryCheckAccessWithFinder is deprecated. Use GetRegistryCheckAccess instead.
// Kept for backward compatibility during migration.
func GetRegistryCheckAccessWithFinder(
	ctx context.Context,
	authorizer authz.Authorizer,
	spaceFinder refcache.SpaceFinder,
	parentID int64,
	art ArtifactInfo,
	reqPermissions ...enum.Permission,
) error {
	return GetRegistryCheckAccess(ctx, authorizer, spaceFinder, art, reqPermissions...)
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
