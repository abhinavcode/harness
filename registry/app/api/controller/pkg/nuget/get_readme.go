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

package nuget

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/harness/gitness/registry/app/pkg/commons"
	nugetmetadata "github.com/harness/gitness/registry/app/metadata/nuget"
	nugettype "github.com/harness/gitness/registry/app/pkg/types/nuget"
)

func (c *controller) GetReadme(
	ctx context.Context,
	info nugettype.ArtifactInfo,
) *GetReadmeResponse {
	// Get image first
	image, err := c.imageDao.GetByName(ctx, info.RegistryID, info.Image)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Str("registry", info.RegIdentifier).
			Str("package", info.Image).
			Msg("failed to get image")
		return &GetReadmeResponse{
			BaseResponse: BaseResponse{
				Error: fmt.Errorf("failed to get image: %w", err),
			},
			ReadmeContent: "",
		}
	}

	// Get artifact by version
	artifact, err := c.artifactDao.GetByName(ctx, image.ID, info.Version)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Str("registry", info.RegIdentifier).
			Str("package", info.Image).
			Str("version", info.Version).
			Msg("failed to get artifact")
		return &GetReadmeResponse{
			BaseResponse: BaseResponse{
				Error: fmt.Errorf("failed to get artifact: %w", err),
			},
			ReadmeContent: "",
		}
	}

	// Unmarshal the metadata to get the readme field
	var metadata nugetmetadata.NugetMetadata
	if err := json.Unmarshal(artifact.Metadata, &metadata); err != nil {
		log.Ctx(ctx).Error().Err(err).
			Str("registry", info.RegIdentifier).
			Str("package", info.Image).
			Str("version", info.Version).
			Str("metadata_raw", string(artifact.Metadata)).
			Msg("failed to unmarshal nuget metadata")
		return &GetReadmeResponse{
			BaseResponse: BaseResponse{
				Error: fmt.Errorf("failed to unmarshal metadata: %w", err),
			},
			ReadmeContent: "",
		}
	}

	// Get the readme content from metadata
	readmeContent := metadata.Metadata.PackageMetadata.Readme
	
	if readmeContent == "" {
		return &GetReadmeResponse{
			BaseResponse: BaseResponse{
				Error: fmt.Errorf("readme not found for package %s version %s", info.Image, info.Version),
			},
			ReadmeContent: "",
		}
	}

	// Set response headers for markdown content
	headers := &commons.ResponseHeaders{
		Headers: map[string]string{
			"Content-Type": "text/markdown; charset=utf-8",
		},
	}

	return &GetReadmeResponse{
		BaseResponse: BaseResponse{
			Error:           nil,
			ResponseHeaders: headers,
		},
		ReadmeContent: readmeContent,
	}
}
