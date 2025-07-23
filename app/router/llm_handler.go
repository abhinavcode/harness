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
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewLLMHandler returns a new handler for LLM-related endpoints.
func NewLLMHandler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Basic health check endpoint
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":  "ok",
			"message": "LLM API is running",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Example completion endpoint
	r.Post("/completion", func(w http.ResponseWriter, r *http.Request) {
		// In a real implementation, this would call an LLM service
		response := map[string]interface{}{
			"status":  "success",
			"message": "LLM completion processed",
			"router":  "llm",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Example router info endpoint
	r.Get("/info",
		func(w http.ResponseWriter, r *http.Request) {
			response := map[string]interface{}{
				"router_name": "llm",

				"status": "active",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		})

	return r
}
