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

package router

import (
	"net/http"
	"strings"

	"github.com/harness/gitness/logging"
)

// LLMRouter handles LLM-related API requests.
type LLMRouter struct {
	handler http.Handler
	prefix  string
}

// NewLLMRouter returns a new LLMRouter.
func NewLLMRouter(handler http.Handler, prefix string) *LLMRouter {
	if prefix == "" {
		prefix = "/api/v1/llm"
	}
	return &LLMRouter{
		handler: handler,
		prefix:  prefix,
	}
}

// Handle processes the LLM-related requests.
func (r *LLMRouter) Handle(

	w http.ResponseWriter, req *http.Request) {
	// Add logging context
	req = req.WithContext(
		logging.NewContext(req.Context(), WithLoggingRouter("llm")))

	// Strip the prefix from the request path
	if err := StripPrefix(r.prefix, req); err != nil {
		http.Error(w, "Failed to strip prefix", http.StatusInternalServerError)
		return
	}

	// Forward the request to the handler
	r.handler.ServeHTTP(w, req)
}

// IsEligibleTraffic checks if the request should be handled by this router.
func (r *LLMRouter) IsEligibleTraffic(req *http.Request) bool {
	return strings.HasPrefix(req.URL.Path, r.prefix)
}

// Name returns the name of the router.
func (r *LLMRouter) Name() string {
	return "llm"
}
