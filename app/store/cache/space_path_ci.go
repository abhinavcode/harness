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

package cache

import (
	"context"
	"strings"
	"time"

	"github.com/harness/gitness/app/store"
	"github.com/harness/gitness/cache"
	"github.com/harness/gitness/types"
)

// NewSpacePathCaseInsensitiveCache creates a cache for case-insensitive space path lookups.
func NewSpacePathCaseInsensitiveCache(
	appCtx context.Context,
	spaceStore store.SpaceStore,
	evictor Evictor[*types.SpaceCore],
	dur time.Duration,
) store.SpacePathCaseInsensitiveCache {
	c := cache.New[string, int64](spacePathCICacheGetter{spaceStore: spaceStore}, dur)

	// Evict cache entries when space is updated
	evictor.Subscribe(appCtx, func(spaceCore *types.SpaceCore) error {
		// Evict the exact lowercase path
		c.Evict(appCtx, strings.ToLower(spaceCore.Path))
		// Also evict by UID for single-segment paths
		c.Evict(appCtx, strings.ToLower(spaceCore.Identifier))
		return nil
	})

	return c
}

type spacePathCICacheGetter struct {
	spaceStore store.SpaceStore
}

func (g spacePathCICacheGetter) Find(ctx context.Context, lowerCasePath string) (int64, error) {
	// The spaceStore.FindByRefCaseInsensitive expects the original case,
	// but it internally converts to lowercase, so we can pass the lowercase version
	return g.spaceStore.FindByRefCaseInsensitive(ctx, lowerCasePath)
}
