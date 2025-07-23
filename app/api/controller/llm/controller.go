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

package llm

import (
	"net/http"

	"github.com/harness/gitness/app/api/render"
)

// Controller handles LLM-related API endpoints.
type Controller struct {
	router Router
}

// Router defines the interface for routing LLM requests.
type Router interface {
	Name() string
}

// New creates a new LLM controller.
func New(router Router) *Controller {
	return &Controller{
		router: router,
	}
}

// HandleCompletion handles LLM completion requests.
func (c *Controller) HandleCompletion(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would process the request and call an LLM service
	response := map[string]interface{}{
		"status": "success",
		"router": c.router.Name(),
		"message": "LLM completion processed",
	}
	
	render.JSON(w, http.StatusOK, response)
}

// GetRouterInfo returns information about the router being used.
func (c *Controller) GetRouterInfo(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"router_name": c.router.Name(),
		"status": "active",
	}
	
	render.JSON(w, http.StatusOK, response)
}
