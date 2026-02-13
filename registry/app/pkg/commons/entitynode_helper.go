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

package commons

import (
	"context"
	"fmt"

	"github.com/harness/gitness/registry/app/services/entitynode"

	"github.com/rs/zerolog/log"
)

func LinkImageEntityToNodes(
	ctx context.Context,
	entityNodeService entitynode.Service,
	imageName string,
	registryID int64,
	artifactType *string,
) error {
	if entityNodeService == nil {
		return nil
	}

	imageInput := entitynode.ImageInput{
		Image:        imageName,
		RegistryID:   registryID,
		ArtifactType: artifactType,
	}

	if err := entityNodeService.LinkEntityToNodes(ctx, imageInput); err != nil {
		log.Ctx(ctx).Error().
			Err(err).
			Str("image", imageName).
			Int64("registry_id", registryID).
			Msg("failed to link image entity to nodes")
		return fmt.Errorf("failed to link image entity to nodes for %s: %w", imageName, err)
	}

	return nil
}

func LinkArtifactEntityToNodes(
	ctx context.Context,
	entityNodeService entitynode.Service,
	imageName string,
	artifactVersion string,
	registryID int64,
	artifactType *string,
) error {
	if entityNodeService == nil {
		return nil
	}

	artifactInput := entitynode.ArtifactInput{
		Image:        imageName,
		Artifact:     artifactVersion,
		RegistryID:   registryID,
		ArtifactType: artifactType,
	}

	if err := entityNodeService.LinkEntityToNodes(ctx, artifactInput); err != nil {
		log.Ctx(ctx).Error().
			Err(err).
			Str("image", imageName).
			Str("artifact", artifactVersion).
			Int64("registry_id", registryID).
			Msg("failed to link artifact entity to nodes")
		return fmt.Errorf("failed to link artifact entity to nodes for %s:%s: %w", imageName, artifactVersion, err)
	}

	return nil
}
