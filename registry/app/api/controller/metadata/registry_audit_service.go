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
	"github.com/harness/gitness/udp"

	"github.com/rs/zerolog/log"
)

type RegistryAuditService struct {
	auditService audit.Service
	udpService   udp.Service
}

func NewRegistryAuditService(auditService audit.Service, udpService udp.Service) *RegistryAuditService {
	return &RegistryAuditService{
		auditService: auditService,
		udpService:   udpService,
	}
}

func (s *RegistryAuditService) Log(
	ctx context.Context,
	action audit.Action,
	oldRegistry *types.Registry,
	newRegistry *types.Registry,
	principal types2.Principal,
	parentRef string,
	resourceType audit.ResourceType,
) {
	var oldObj, newObj interface{}
	var auditOptions []audit.Option
	registryName := ""

	if newRegistry != nil {
		registryName = newRegistry.Name
		newObj = createRegistryAuditObject(newRegistry)
		auditOptions = append(auditOptions, audit.WithNewObject(newObj))
	}

	if oldRegistry != nil {
		if registryName == "" {
			registryName = oldRegistry.Name
		}
		oldObj = createRegistryAuditObject(oldRegistry)
		auditOptions = append(auditOptions, audit.WithOldObject(oldObj))
	}

	if action == audit.ActionDeleted {
		auditOptions = append(auditOptions, audit.WithData("registry name", registryName))
	}

	auditErr := s.auditService.Log(
		ctx,
		principal,
		audit.NewResource(resourceType, registryName),
		action,
		parentRef,
		auditOptions...,
	)
	if auditErr != nil {
		log.Ctx(ctx).Warn().Msgf("failed to insert audit log for %s registry operation: %s", action, auditErr)
	}

	udpAction := convertAuditActionToUDP(action)
	udpResourceType := convertAuditResourceTypeToUDP(resourceType)

	s.udpService.InsertEvent(
		ctx,
		udpAction,
		udpResourceType,
		registryName,
		parentRef,
		principal,
		newObj,
		oldObj,
	)
}

func (s *RegistryAuditService) LogWithUpstreamProxy(
	ctx context.Context,
	action audit.Action,
	oldRegistry *types.Registry,
	newRegistry *types.Registry,
	oldUpstreamProxy *types.UpstreamProxy,
	newUpstreamProxy *types.UpstreamProxy,
	principal types2.Principal,
	parentRef string,
) {
	var oldObj, newObj interface{}
	var auditOptions []audit.Option
	registryName := ""

	if newRegistry != nil && newUpstreamProxy != nil {
		registryName = newRegistry.Name
		newObj = audit.RegistryUpstreamProxyConfigObjectEnhanced{
			UUID:            newRegistry.UUID,
			Name:            newRegistry.Name,
			ParentID:        newRegistry.ParentID,
			RootParentID:    newRegistry.RootParentID,
			Description:     newRegistry.Description,
			Type:            string(newRegistry.Type),
			PackageType:     string(newRegistry.PackageType),
			UpstreamProxies: newRegistry.UpstreamProxies,
			AllowedPattern:  newRegistry.AllowedPattern,
			BlockedPattern:  newRegistry.BlockedPattern,
			Labels:          newRegistry.Labels,
			Source:          newUpstreamProxy.Source,
			URL:             newUpstreamProxy.RepoURL,
			AuthType:        newUpstreamProxy.RepoAuthType,
			CreatedAt:       newUpstreamProxy.CreatedAt,
			UpdatedAt:       newUpstreamProxy.UpdatedAt,
			CreatedBy:       newUpstreamProxy.CreatedBy,
			UpdatedBy:       newUpstreamProxy.UpdatedBy,
			IsPublic:        newRegistry.IsPublic,
		}
		auditOptions = append(auditOptions, audit.WithNewObject(newObj))
	}

	if oldRegistry != nil && oldUpstreamProxy != nil {
		if registryName == "" {
			registryName = oldRegistry.Name
		}
		oldObj = audit.RegistryUpstreamProxyConfigObjectEnhanced{
			UUID:            oldRegistry.UUID,
			Name:            oldRegistry.Name,
			ParentID:        oldRegistry.ParentID,
			RootParentID:    oldRegistry.RootParentID,
			Description:     oldRegistry.Description,
			Type:            string(oldRegistry.Type),
			PackageType:     string(oldRegistry.PackageType),
			UpstreamProxies: oldRegistry.UpstreamProxies,
			AllowedPattern:  oldRegistry.AllowedPattern,
			BlockedPattern:  oldRegistry.BlockedPattern,
			Labels:          oldRegistry.Labels,
			Source:          oldUpstreamProxy.Source,
			URL:             oldUpstreamProxy.RepoURL,
			AuthType:        oldUpstreamProxy.RepoAuthType,
			CreatedAt:       oldUpstreamProxy.CreatedAt,
			UpdatedAt:       oldUpstreamProxy.UpdatedAt,
			CreatedBy:       oldUpstreamProxy.CreatedBy,
			UpdatedBy:       oldUpstreamProxy.UpdatedBy,
			IsPublic:        oldRegistry.IsPublic,
		}
		auditOptions = append(auditOptions, audit.WithOldObject(oldObj))
	}

	auditErr := s.auditService.Log(
		ctx,
		principal,
		audit.NewResource(audit.ResourceTypeRegistryUpstreamProxy, registryName),
		action,
		parentRef,
		auditOptions...,
	)
	if auditErr != nil {
		log.Ctx(ctx).Warn().Msgf("failed to insert audit log for %s upstream proxy: %s", action, auditErr)
	}

	udpAction := convertAuditActionToUDP(action)
	udpResourceType := convertAuditResourceTypeToUDP(audit.ResourceTypeRegistryUpstreamProxy)

	s.udpService.InsertEvent(
		ctx,
		udpAction,
		udpResourceType,
		registryName,
		parentRef,
		principal,
		newObj,
		oldObj,
	)
}

// convertAuditActionToUDP converts audit.Action to UDP action format.
func convertAuditActionToUDP(action audit.Action) string {
	//nolint:exhaustive // Only handle registry-specific actions
	switch action {
	case audit.ActionCreated:
		return udp.ActionRegistryCreated
	case audit.ActionUpdated:
		return udp.ActionRegistryUpdated
	case audit.ActionDeleted:
		return udp.ActionRegistryDeleted
	default:
		return string(action)
	}
}

// convertAuditResourceTypeToUDP converts audit.ResourceType to UDP resource type format.
func convertAuditResourceTypeToUDP(resourceType audit.ResourceType) string {
	//nolint:exhaustive // Only handle registry-specific resource types
	switch resourceType {
	case audit.ResourceTypeRegistry:
		return udp.ResourceTypeRegistryVirtual
	case audit.ResourceTypeRegistryUpstreamProxy:
		return udp.ResourceTypeRegistryUpstreamProxy
	default:
		return string(resourceType)
	}
}

func createRegistryAuditObject(registry *types.Registry) audit.RegistryObject {
	return audit.RegistryObject{Registry: *registry}
}
