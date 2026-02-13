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

type EntityInput interface {
	GetType() EntityType
	GetRegistryID() int64
}

type EntityType string

type ImageInput struct {
	Image        string
	RegistryID   int64
	ArtifactType *string
}

func (i ImageInput) GetType() EntityType {
	return EntityTypeImage
}

func (i ImageInput) GetRegistryID() int64 {
	return i.RegistryID
}

type ArtifactInput struct {
	Image        string
	Artifact     string
	RegistryID   int64
	ArtifactType *string
}

func (a ArtifactInput) GetType() EntityType {
	return EntityTypeArtifact
}

func (a ArtifactInput) GetRegistryID() int64 {
	return a.RegistryID
}

const (
	EntityTypeImage    EntityType = "image"
	EntityTypeArtifact EntityType = "artifact"
)