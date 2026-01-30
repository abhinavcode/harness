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

package audit

import (
	"context"
	"net"
	"net/http"
	"strings"
)

var (
	trueClientIP  = http.CanonicalHeaderKey("True-Client-IP")
	xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
	xRealIP       = http.CanonicalHeaderKey("X-Real-IP")
	xRegion       = http.CanonicalHeaderKey("X-Region")
)

// Middleware process request headers to fill internal info data.
func Middleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if rip := RealIP(r); rip != "" {
				ctx = context.WithValue(ctx, realIPKey, rip)
			}

			ctx = context.WithValue(ctx, pathKey, r.URL.Path)
			ctx = context.WithValue(ctx, requestMethod, r.Method)
			ctx = context.WithValue(ctx, requestID, w.Header().Get("X-Request-Id"))

			// HACK: Set region-based location from x-region header for testing.
			if region := r.Header.Get(xRegion); region != "" {
				if loc := getRegionLocation(region); loc != nil {
					ctx = context.WithValue(ctx, regionLocationKey, loc)
				} else {
					ctx = context.WithValue(ctx, regionLocationKey, loc)
				}
			} else {
				ctx = context.WithValue(ctx, regionLocationKey, getRegionLocation(region))
			}

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// getRegionLocation returns lat/long for known regions (hacky test solution).
func getRegionLocation(region string) *RegionLocation {
	switch strings.ToLower(region) {
	case "apac":
		// Mumbai, India.
		return &RegionLocation{Latitude: 19.0760, Longitude: 72.8777}
	case "wnam":
		// San Jose, California.
		return &RegionLocation{Latitude: 37.3382, Longitude: -121.8863}
	default:
		return &RegionLocation{Latitude: 37.3382, Longitude: -121.8863}
	}
}

// RealIP extracts the real client IP from the HTTP request.
func RealIP(r *http.Request) string {
	var ip string

	if tcip := r.Header.Get(trueClientIP); tcip != "" {
		ip = tcip
	} else if xrip := r.Header.Get(xRealIP); xrip != "" {
		ip = xrip
	} else if xff := r.Header.Get(xForwardedFor); xff != "" {
		i := strings.Index(xff, ",")
		if i == -1 {
			i = len(xff)
		}
		ip = xff[:i]
	} else {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}
	if ip == "" || net.ParseIP(ip) == nil {
		return ""
	}
	return ip
}
