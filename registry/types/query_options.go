//  Copyright 2023 Harness, Inc.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package types

// QueryOptions holds configuration options for database queries.
type QueryOptions struct {
	SoftDeleteFilter SoftDeleteFilter
}

// QueryOption is a function that modifies QueryOptions.
type QueryOption func(o *QueryOptions)

// MakeQueryOptions creates QueryOptions with defaults and applies any provided options.
func MakeQueryOptions(opts ...QueryOption) QueryOptions {
	opt := QueryOptions{
		SoftDeleteFilter: SoftDeleteFilterExclude, // Default: exclude soft-deleted entities
	}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

// WithSoftDeleteFilter sets the soft delete filter option.
func WithSoftDeleteFilter(filter SoftDeleteFilter) QueryOption {
	return func(o *QueryOptions) {
		o.SoftDeleteFilter = filter
	}
}

// WithAllDeleted is a convenience function to include all entities (including soft-deleted).
func WithAllDeleted() QueryOption {
	return WithSoftDeleteFilter(SoftDeleteFilterInclude)
}

// WithOnlyDeleted is a convenience function to only include soft-deleted entities.
func WithOnlyDeleted() QueryOption {
	return WithSoftDeleteFilter(SoftDeleteFilterOnly)
}

// WithExcludeDeleted is a convenience function to exclude soft-deleted entities (default behavior).
func WithExcludeDeleted() QueryOption {
	return WithSoftDeleteFilter(SoftDeleteFilterExclude)
}

// ExtractSoftDeleteFilter extracts the SoftDeleteFilter from QueryOptions.
func ExtractSoftDeleteFilter(opts ...QueryOption) SoftDeleteFilter {
	qo := MakeQueryOptions(opts...)
	return qo.SoftDeleteFilter
}
