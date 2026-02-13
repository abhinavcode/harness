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

package deletion

import (
	"context"
	"fmt"

	"github.com/harness/gitness/registry/app/api/openapi/contracts/artifact"
	"github.com/harness/gitness/registry/app/api/utils"
	"github.com/harness/gitness/registry/app/manifest/manifestlist"
	"github.com/harness/gitness/registry/app/pkg/filemanager"
	"github.com/harness/gitness/registry/app/store"
	registrytypes "github.com/harness/gitness/registry/types"
	"github.com/harness/gitness/store/database/dbtx"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// PackageWrapper defines the interface for handling custom package types.
// This matches the interfaces.PackageWrapper interface from the API layer.
type PackageWrapper interface {
	DeleteArtifact(ctx context.Context, regInfo *registrytypes.RegistryRequestBaseInfo, artifactName string) error
}

// Service provides package-type-specific deletion logic for registry entities.
// This service is used by both API controllers and cleanup jobs to ensure consistent deletion behavior.
type Service struct {
	artifactStore         store.ArtifactRepository
	imageStore            store.ImageRepository
	manifestStore         store.ManifestRepository
	tagStore              store.TagRepository
	registryBlobStore     store.RegistryBlobRepository
	fileManager           filemanager.FileManager
	tx                    dbtx.Transactor
	untaggedImagesEnabled func(ctx context.Context) bool
	packageWrapper        PackageWrapper
}

// NewService creates a new deletion service.
func NewService(
	artifactStore store.ArtifactRepository,
	imageStore store.ImageRepository,
	manifestStore store.ManifestRepository,
	tagStore store.TagRepository,
	registryBlobStore store.RegistryBlobRepository,
	fileManager filemanager.FileManager,
	tx dbtx.Transactor,
	untaggedImagesEnabled func(ctx context.Context) bool,
	packageWrapper PackageWrapper,
) *Service {
	return &Service{
		artifactStore:         artifactStore,
		imageStore:            imageStore,
		manifestStore:         manifestStore,
		tagStore:              tagStore,
		registryBlobStore:     registryBlobStore,
		fileManager:           fileManager,
		tx:                    tx,
		untaggedImagesEnabled: untaggedImagesEnabled,
		packageWrapper:        packageWrapper,
	}
}

// DeleteArtifactVersion deletes an artifact version using package-type-specific logic.
func (s *Service) DeleteArtifactVersion(
	ctx context.Context,
	registryID int64,
	packageType artifact.PackageType,
	artifactName string,
	versionName string,
) error {
	//nolint:exhaustive
	switch packageType {
	case artifact.PackageTypeDOCKER, artifact.PackageTypeHELM:
		return s.DeleteOCIArtifact(ctx, registryID, artifactName, versionName)
	case artifact.PackageTypeNPM, artifact.PackageTypeMAVEN, artifact.PackageTypePYTHON,
		artifact.PackageTypeGENERIC, artifact.PackageTypeNUGET, artifact.PackageTypeRPM,
		artifact.PackageTypeGO:
		return s.DeleteGenericArtifact(ctx, registryID, packageType, artifactName, versionName)
	default:
		// For unknown types, return error so caller can try PackageWrapper
		return fmt.Errorf("unsupported package type: %s", packageType)
	}
}

// DeleteImageByPackageType deletes an image using package-type-specific logic.
// This is used by delete_artifact.go to route image deletion based on package type.
func (s *Service) DeleteImageByPackageType(
	ctx context.Context,
	regInfo *registrytypes.RegistryRequestBaseInfo,
	packageType artifact.PackageType,
	imageName string,
) error {
	registryID := regInfo.RegistryID

	//nolint:exhaustive
	switch packageType {
	case artifact.PackageTypeDOCKER, artifact.PackageTypeHELM:
		return s.DeleteOCIImage(ctx, registryID, imageName)
	case artifact.PackageTypeGENERIC, artifact.PackageTypeMAVEN, artifact.PackageTypePYTHON,
		artifact.PackageTypeNPM, artifact.PackageTypeNUGET, artifact.PackageTypeGO:
		return s.DeleteGenericImage(ctx, registryID, packageType, imageName)
	case artifact.PackageTypeRPM:
		return fmt.Errorf("delete artifact not supported for rpm")
	case artifact.PackageTypeHUGGINGFACE:
		return fmt.Errorf("unsupported package type: %s", packageType)
	default:
		return s.packageWrapper.DeleteArtifact(ctx, regInfo, imageName)
	}
}

// DeleteOCIArtifact handles Docker/Helm artifact deletion.
func (s *Service) DeleteOCIArtifact(
	ctx context.Context,
	registryID int64,
	imageName string,
	version string,
) error {
	if !s.untaggedImagesEnabled(ctx) {
		// Non-untagged mode: just delete the tag
		return s.tagStore.DeleteTag(ctx, registryID, imageName, version)
	}

	return s.deleteOCIArtifactUntaggedMode(ctx, registryID, imageName, version)
}

// deleteOCIArtifactUntaggedMode handles deletion in untagged images mode.
func (s *Service) deleteOCIArtifactUntaggedMode(
	ctx context.Context,
	registryID int64,
	imageName string,
	version string,
) error {
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		existingManifest, err := s.findManifestByVersion(ctx, registryID, imageName, version)
		if err != nil {
			return err
		}

		if err := s.validateManifestReferences(ctx, existingManifest, version); err != nil {
			return err
		}

		if err := s.deleteManifestAndTags(ctx, registryID, existingManifest, version); err != nil {
			return err
		}

		if err := s.artifactStore.DeleteByVersionAndImageName(ctx, imageName, version, registryID); err != nil {
			return err
		}

		return s.deleteImageIfNoManifests(ctx, registryID, imageName)
	})
}

// findManifestByVersion finds a manifest by its version digest.
func (s *Service) findManifestByVersion(
	ctx context.Context,
	registryID int64,
	imageName string,
	version string,
) (*registrytypes.Manifest, error) {
	d := digest.Digest(version)
	dgst, _ := registrytypes.NewDigest(d)
	manifest, err := s.manifestStore.FindManifestByDigest(ctx, registryID, imageName, dgst)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing manifest for: %s, err: %w", version, err)
	}
	return manifest, nil
}

// validateManifestReferences checks if a manifest list references other manifests.
func (s *Service) validateManifestReferences(
	ctx context.Context,
	manifest *registrytypes.Manifest,
	version string,
) error {
	if manifest.MediaType != v1.MediaTypeImageIndex &&
		manifest.MediaType != manifestlist.MediaTypeManifestList {
		return nil
	}

	manifests, err := s.manifestStore.References(ctx, manifest)
	if err != nil {
		return fmt.Errorf("failed to find existing manifests referenced by: %s, err: %w", version, err)
	}

	if len(manifests) > 0 {
		return fmt.Errorf("cannot delete manifest: %s, as it references other manifests", version)
	}

	return nil
}

// deleteManifestAndTags deletes a manifest and its associated tags.
func (s *Service) deleteManifestAndTags(
	ctx context.Context,
	registryID int64,
	manifest *registrytypes.Manifest,
	version string,
) error {
	if err := s.manifestStore.Delete(ctx, registryID, manifest.ID); err != nil {
		return err
	}

	if _, err := s.tagStore.DeleteTagByManifestID(ctx, registryID, manifest.ID); err != nil {
		return fmt.Errorf("failed to delete tags for: %s, err: %w", version, err)
	}

	return nil
}

// deleteImageIfNoManifests deletes an image if it has no remaining manifests.
func (s *Service) deleteImageIfNoManifests(
	ctx context.Context,
	registryID int64,
	imageName string,
) error {
	count, err := s.manifestStore.CountByImageName(ctx, registryID, imageName)
	if err != nil {
		return err
	}

	if count < 1 {
		return s.imageStore.DeleteByImageNameAndRegID(ctx, registryID, imageName)
	}

	return nil
}

// DeleteGenericArtifact handles generic package deletion (NPM, Maven, Python, etc.).
func (s *Service) DeleteGenericArtifact(
	ctx context.Context,
	registryID int64,
	packageType artifact.PackageType,
	artifactName string,
	versionName string,
) error {
	// Get file path
	filePath, err := utils.GetFilePath(packageType, artifactName, versionName)
	if err != nil {
		return fmt.Errorf("failed to get file path: %w", err)
	}

	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		// Delete files from storage
		err = s.fileManager.DeleteFile(ctx, registryID, filePath)
		if err != nil {
			return err
		}

		// Delete artifact from DB
		err = s.artifactStore.DeleteByVersionAndImageName(ctx, artifactName, versionName, registryID)
		if err != nil {
			return fmt.Errorf("failed to delete version: %w", err)
		}

		// Delete image if no other artifacts linked
		err = s.imageStore.DeleteByImageNameIfNoLinkedArtifacts(ctx, registryID, artifactName)
		if err != nil {
			return fmt.Errorf("failed to delete image: %w", err)
		}

		return nil
	})
}

// DeleteOCIImage handles Docker/Helm image deletion (deletes all artifacts, manifests, blobs).
func (s *Service) DeleteOCIImage(
	ctx context.Context,
	registryID int64,
	imageName string,
) error {
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		// Delete manifests linked to the image
		_, err := s.manifestStore.DeleteManifestByImageName(ctx, registryID, imageName)
		if err != nil {
			return fmt.Errorf("failed to delete manifests: %w", err)
		}

		// Delete registry blobs linked to the image
		_, err = s.registryBlobStore.UnlinkBlobByImageName(ctx, registryID, imageName)
		if err != nil {
			return fmt.Errorf("failed to delete registry blobs: %w", err)
		}

		// Delete all artifacts linked to image
		err = s.artifactStore.DeleteByImageNameAndRegistryID(ctx, registryID, imageName)
		if err != nil {
			return fmt.Errorf("failed to delete artifacts: %w", err)
		}

		// Delete image
		err = s.imageStore.DeleteByImageNameAndRegID(ctx, registryID, imageName)
		if err != nil {
			return fmt.Errorf("failed to delete image: %w", err)
		}

		return nil
	})
}

// DeleteGenericImage handles generic package image deletion (deletes files and artifacts).
func (s *Service) DeleteGenericImage(
	ctx context.Context,
	registryID int64,
	packageType artifact.PackageType,
	imageName string,
) error {
	// Get file path
	filePath, err := utils.GetFilePath(packageType, imageName, "")
	if err != nil {
		return fmt.Errorf("failed to get file path: %w", err)
	}

	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		// Delete files from storage
		err = s.fileManager.DeleteFile(ctx, registryID, filePath)
		if err != nil {
			return fmt.Errorf("failed to delete files: %w", err)
		}

		// Delete all artifacts
		err = s.artifactStore.DeleteByImageNameAndRegistryID(ctx, registryID, imageName)
		if err != nil {
			return fmt.Errorf("failed to delete artifacts: %w", err)
		}

		// Delete image
		err = s.imageStore.DeleteByImageNameAndRegID(ctx, registryID, imageName)
		if err != nil {
			return fmt.Errorf("failed to delete image: %w", err)
		}

		return nil
	})
}
