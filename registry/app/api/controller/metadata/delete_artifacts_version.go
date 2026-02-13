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
	"errors"
	"fmt"
	"net/http"

	apiauth "github.com/harness/gitness/app/api/auth"
	"github.com/harness/gitness/app/api/request"
	"github.com/harness/gitness/audit"
	"github.com/harness/gitness/registry/app/api/openapi/contracts/artifact"
	"github.com/harness/gitness/registry/services/webhook"
	registryTypes "github.com/harness/gitness/registry/types"
	"github.com/harness/gitness/store"
	"github.com/harness/gitness/types"
	"github.com/harness/gitness/types/enum"

	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
)

func (c *APIController) DeleteArtifactVersion(ctx context.Context, r artifact.DeleteArtifactVersionRequestObject) (
	artifact.DeleteArtifactVersionResponseObject, error,
) {
	regInfo, err := c.RegistryMetadataHelper.GetRegistryRequestBaseInfo(ctx, "", string(r.RegistryRef))
	if err != nil {
		return artifact.DeleteArtifactVersion400JSONResponse{
			BadRequestJSONResponse: artifact.BadRequestJSONResponse(
				*GetErrorResponse(http.StatusBadRequest, err.Error()),
			),
		}, err
	}
	space, err := c.SpaceFinder.FindByRef(ctx, regInfo.ParentRef)
	if err != nil {
		return artifact.DeleteArtifactVersion400JSONResponse{
			BadRequestJSONResponse: artifact.BadRequestJSONResponse(
				*GetErrorResponse(http.StatusBadRequest, err.Error()),
			),
		}, err
	}

	session, _ := request.AuthSessionFrom(ctx)
	if err = apiauth.CheckSpaceScope(
		ctx,
		c.Authorizer,
		session,
		space,
		enum.ResourceTypeRegistry,
		enum.PermissionArtifactsDelete,
	); err != nil {
		statusCode, message := HandleAuthError(err)
		if statusCode == http.StatusUnauthorized {
			return artifact.DeleteArtifactVersion401JSONResponse{
				UnauthenticatedJSONResponse: artifact.UnauthenticatedJSONResponse(
					*GetErrorResponse(http.StatusUnauthorized, message),
				),
			}, nil
		}
		return artifact.DeleteArtifactVersion403JSONResponse{
			UnauthorizedJSONResponse: artifact.UnauthorizedJSONResponse(
				*GetErrorResponse(http.StatusForbidden, message),
			),
		}, nil
	}

	repoEntity, err := c.RegistryRepository.GetByParentIDAndName(
		ctx,
		regInfo.ParentID,
		regInfo.RegistryIdentifier,
	)
	if err != nil {
		//nolint:nilerr
		return artifact.DeleteArtifactVersion404JSONResponse{
			NotFoundJSONResponse: artifact.NotFoundJSONResponse(
				*GetErrorResponse(
					http.StatusNotFound,
					fmt.Sprintf("registry %s doesn't exist", regInfo.RegistryIdentifier),
				),
			),
		}, nil
	}

	artifactName := string(r.Artifact)
	versionName := string(r.Version)
	registryName := repoEntity.Name

	imageInfo, err := c.ImageStore.GetByName(ctx, repoEntity.ID, artifactName)
	if err != nil {
		//nolint:nilerr
		return artifact.DeleteArtifactVersion404JSONResponse{
			NotFoundJSONResponse: artifact.NotFoundJSONResponse(
				*GetErrorResponse(http.StatusNotFound, "image doesn't exist with this key"),
			),
		}, nil
	}

	//nolint: exhaustive
	switch regInfo.PackageType {
	case artifact.PackageTypeDOCKER:
		err = c.deleteOciVersionWithAudit(ctx, regInfo, registryName, session.Principal, artifactName,
			versionName)
	case artifact.PackageTypeHELM:
		err = c.deleteOciVersionWithAudit(ctx, regInfo, registryName, session.Principal, artifactName,
			versionName)
	case artifact.PackageTypeNPM:
		err = c.deleteVersion(ctx, regInfo, imageInfo, artifactName, versionName)
	case artifact.PackageTypeMAVEN:
		err = c.deleteVersion(ctx, regInfo, imageInfo, artifactName, versionName)
	case artifact.PackageTypePYTHON:
		err = c.deleteVersion(ctx, regInfo, imageInfo, artifactName, versionName)
	case artifact.PackageTypeGENERIC:
		err = c.deleteVersion(ctx, regInfo, imageInfo, artifactName, versionName)
	case artifact.PackageTypeNUGET:
		err = c.deleteVersion(ctx, regInfo, imageInfo, artifactName, versionName)
	case artifact.PackageTypeRPM:
		err = c.deleteVersion(ctx, regInfo, imageInfo, artifactName, versionName)
		if err != nil {
			break
		}
		c.PostProcessingReporter.BuildRegistryIndex(ctx, regInfo.RegistryID, make([]registryTypes.SourceRef, 0))
	case artifact.PackageTypeGO:
		err = c.deleteVersion(ctx, regInfo, imageInfo, artifactName, versionName)
		if err != nil {
			break
		}
		c.sendArtifactDeletedWebhookEvent(
			ctx, session.Principal.ID, regInfo.RegistryID, regInfo.PackageType,
			artifactName, versionName,
		)
		c.PostProcessingReporter.BuildPackageIndex(ctx, regInfo.RegistryID, artifactName)
	default:
		err = c.PackageWrapper.DeleteArtifactVersion(ctx, regInfo, imageInfo, artifactName, versionName)
	}

	if err != nil {
		if errors.Is(err, store.ErrResourceNotFound) {
			return artifact.DeleteArtifactVersion404JSONResponse{
				NotFoundJSONResponse: artifact.NotFoundJSONResponse(
					*GetErrorResponse(
						http.StatusNotFound,
						fmt.Sprintf("artifact version '%s' not found for artifact '%s'", versionName, artifactName),
					),
				),
			}, nil
		}
		return throwDeleteArtifactVersion500Error(err), nil
	}

	auditErr := c.AuditService.Log(
		ctx,
		session.Principal,
		audit.NewResource(audit.ResourceTypeRegistry, artifactName),
		audit.ActionDeleted,
		regInfo.ParentRef,
		audit.WithData("registry name", registryName),
		audit.WithData("artifact name", artifactName),
		audit.WithData("version name", versionName),
	)
	if auditErr != nil {
		log.Ctx(ctx).Warn().Msgf("failed to insert audit log for delete artifact operation: %s", auditErr)
	}

	return artifact.DeleteArtifactVersion200JSONResponse{
		SuccessJSONResponse: artifact.SuccessJSONResponse(*GetSuccessResponse()),
	}, nil
}

func (c *APIController) deleteOciVersionWithAudit(
	ctx context.Context, regInfo *registryTypes.RegistryRequestBaseInfo,
	registryName string, principal types.Principal, artifactName string, versionName string,
) error {
	var existingDigest digest.Digest

	if c.UntaggedImagesEnabled(ctx) {
		existingDigest = digest.Digest(versionName)
	} else {
		existingDigest = c.getTagDigest(ctx, regInfo.RegistryID, artifactName, versionName)
	}

	err := c.DeletionService.DeleteOCIArtifact(ctx, regInfo.RegistryID, artifactName, versionName)
	if err != nil {
		return fmt.Errorf("failed to delete artifact version: %w", err)
	}
	if existingDigest != "" {
		payload := webhook.GetArtifactDeletedPayload(ctx, principal.ID, regInfo.RegistryID,
			registryName, versionName, existingDigest.String(), regInfo.RootIdentifier,
			regInfo.PackageType, artifactName, c.URLProvider, c.UntaggedImagesEnabled(ctx))
		c.ArtifactEventReporter.ArtifactDeleted(ctx, &payload)
	}

	return nil
}

func (c *APIController) deleteVersion(
	ctx context.Context,
	regInfo *registryTypes.RegistryRequestBaseInfo,
	imageInfo *registryTypes.Image,
	artifactName string,
	versionName string,
) error {
	_, err := c.ArtifactStore.GetByName(ctx, imageInfo.ID, versionName)
	if err != nil {
		return fmt.Errorf("version doesn't exist for image %v: %w", imageInfo.Name, err)
	}

	return c.DeletionService.DeleteGenericArtifact(ctx, regInfo.RegistryID, regInfo.PackageType, artifactName, versionName)
}

func (c *APIController) sendArtifactDeletedWebhookEvent(
	ctx context.Context, principalID int64,
	registryID int64, packageType artifact.PackageType,
	artifact string, version string,
) {
	payload := webhook.GetArtifactDeletedPayloadForCommonArtifacts(
		principalID,
		registryID,
		packageType,
		artifact,
		version,
	)
	c.ArtifactEventReporter.ArtifactDeleted(ctx, &payload)
}

func throwDeleteArtifactVersion500Error(err error) artifact.DeleteArtifactVersion500JSONResponse {
	return artifact.DeleteArtifactVersion500JSONResponse{
		InternalServerErrorJSONResponse: artifact.InternalServerErrorJSONResponse(
			*GetErrorResponse(http.StatusInternalServerError, err.Error()),
		),
	}
}

func (c *APIController) getTagDigest(
	ctx context.Context,
	registryID int64,
	imageName string,
	tag string,
) digest.Digest {
	existingTag, findTagErr := c.TagStore.FindTag(ctx, registryID, imageName, tag)
	if findTagErr == nil && existingTag != nil {
		existingTaggedManifest, getManifestErr := c.ManifestStore.Get(ctx, existingTag.ManifestID)
		if getManifestErr == nil {
			return existingTaggedManifest.Digest
		}
	}
	return ""
}
