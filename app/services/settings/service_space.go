// Copyright 2023 Harness, Inc.
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

package settings

import (
	"context"

	"github.com/harness/gitness/types/enum"
)

// SpaceSet sets the value of the setting with the given key for the given space.
func (s *Service) SpaceSet(
	ctx context.Context,
	spaceID int64,
	key Key,
	value any,
) error {
	return s.Set(
		ctx,
		enum.SettingsScopeSpace,
		spaceID,
		key,
		value,
	)
}

// SpaceSetMany sets the value of the settings with the given keys for the given space.
func (s *Service) SpaceSetMany(
	ctx context.Context,
	spaceID int64,
	keyValues ...KeyValue,
) error {
	return s.SetMany(
		ctx,
		enum.SettingsScopeSpace,
		spaceID,
		keyValues...,
	)
}

// SpaceGet returns the value of the setting with the given key for the given space.
func (s *Service) SpaceGet(
	ctx context.Context,
	spaceID int64,
	key Key,
	out any,
) (bool, error) {
	return s.Get(
		ctx,
		enum.SettingsScopeSpace,
		spaceID,
		key,
		out,
	)
}

// SpaceMap maps all available settings using the provided handlers for the given space.
func (s *Service) SpaceMap(
	ctx context.Context,
	spaceID int64,
	handlers ...SettingHandler,
) error {
	return s.Map(
		ctx,
		enum.SettingsScopeSpace,
		spaceID,
		handlers...,
	)
}
