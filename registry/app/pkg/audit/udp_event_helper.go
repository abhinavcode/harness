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

package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/harness/gitness/audit"
	"github.com/harness/gitness/registry/app/store"
	"github.com/harness/gitness/registry/types"
	gitnesstypes "github.com/harness/gitness/types"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// LogWithUDPEvent logs an audit event to both the audit service and UDP events table.
// This is a wrapper that calls the audit service and also inserts into udp_events table.
// The payload stored in udp_events will have a similar structure to the audit service job_data.
func LogWithUDPEvent(
	ctx context.Context,
	auditService audit.Service,
	udpEventStore store.UDPEventRepository,
	principal gitnesstypes.Principal,
	resource audit.Resource,
	action audit.Action,
	spacePath string,
	options ...audit.Option,
) error {
	// First, log to the audit service (existing behavior)
	auditErr := auditService.Log(ctx, principal, resource, action, spacePath, options...)
	if auditErr != nil {
		return auditErr
	}

	// Also insert into UDP events table if store is available
	if udpEventStore != nil {
		insertUDPEvent(ctx, udpEventStore, principal, resource, action, spacePath, options)
	}

	return nil
}

func insertUDPEvent(
	ctx context.Context,
	udpEventStore store.UDPEventRepository,
	principal gitnesstypes.Principal,
	resource audit.Resource,
	action audit.Action,
	spacePath string,
	options []audit.Option,
) {
	// Build the audit event to extract data
	event := &audit.Event{}
	for _, opt := range options {
		opt.Apply(event)
	}

	// Build payload structure similar to what audit service creates
	auditPayload := buildAuditPayload(ctx, principal, resource, action, spacePath, event)

	// Marshal to JSON (similar format as job_data)
	payloadJSON, err := json.Marshal(auditPayload)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("failed to marshal audit payload for UDP events")
		return
	}

	// Insert into UDP events table
	udpEvent := &types.UDPEvent{
		DataType: types.UDPEventTypeAudits,
		Payload:  string(payloadJSON),
	}

	if udpErr := udpEventStore.Create(ctx, udpEvent); udpErr != nil {
		log.Ctx(ctx).Warn().Err(udpErr).Msg("failed to insert audit event into UDP events table")
	}
}

func buildAuditPayload(
	ctx context.Context,
	principal gitnesstypes.Principal,
	resource audit.Resource,
	action audit.Action,
	spacePath string,
	event *audit.Event,
) map[string]interface{} {
	auditPayload := map[string]interface{}{
		"principal": map[string]interface{}{
			"email":       principal.Email,
			"uid":         principal.UID,
			"displayName": principal.DisplayName,
			"type":        string(principal.Type),
		},
		"resource": map[string]interface{}{
			"type":       string(resource.Type),
			"identifier": resource.Identifier,
			"data":       resource.DataAsSlice(),
		},
		"action":    string(action),
		"spacePath": spacePath,
		"timestamp": time.Now().Unix(),
	}

	// Add additional event data if present
	if len(event.Data) > 0 {
		auditPayload["eventData"] = event.Data
	}

	// Add old/new objects if present (for update operations)
	if event.DiffObject.OldObject != nil {
		oldYAML, err := yaml.Marshal(event.DiffObject.OldObject)
		if err == nil {
			auditPayload["oldValue"] = string(oldYAML)
		}
	}

	if event.DiffObject.NewObject != nil {
		newYAML, err := yaml.Marshal(event.DiffObject.NewObject)
		if err == nil {
			auditPayload["newValue"] = string(newYAML)
		}
	}

	// Add client info
	clientIP := event.ClientIP
	if clientIP == "" {
		clientIP = audit.GetRealIP(ctx)
	}
	if clientIP != "" {
		auditPayload["clientIP"] = clientIP
	}

	requestMethod := event.RequestMethod
	if requestMethod == "" {
		requestMethod = audit.GetRequestMethod(ctx)
	}
	if requestMethod != "" {
		auditPayload["requestMethod"] = requestMethod
	}

	// Add correlation ID
	correlationID := audit.GetRequestID(ctx)
	if correlationID != "" {
		auditPayload["correlationID"] = correlationID
	}

	return auditPayload
}
