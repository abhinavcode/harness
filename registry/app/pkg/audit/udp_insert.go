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
	"strings"
	"time"

	"github.com/harness/gitness/audit"
	"github.com/harness/gitness/registry/types"
	"github.com/harness/gitness/store/database/dbtx"
	gitnesstypes "github.com/harness/gitness/types"

	"github.com/rs/zerolog/log"
)

// Action constants for UDP audit events.
const (
	ActionRegistryCreated  = "REGISTRY_CREATED"
	ActionRegistryUpdated  = "REGISTRY_UPDATED"
	ActionRegistryDeleted  = "REGISTRY_DELETED"
	ActionArtifactDeleted  = "ARTIFACT_DELETED"
	ActionVersionDeleted   = "VERSION_DELETED"
	ActionArtifactUploaded = "ARTIFACT_UPLOADED"
)

// Constants for audit payload field names.
const (
	FieldResourceScope      = "resourceScope"
	FieldHTTPRequestInfo    = "httpRequestInfo"
	FieldRequestMetadata    = "requestMetadata"
	FieldTimestamp          = "timestamp"
	FieldAuthenticationInfo = "authenticationInfo"
	FieldModule             = "module"
	FieldResource           = "resource"
	FieldAction             = "action"

	// Resource scope fields.
	FieldAccountIdentifier = "accountIdentifier"
	FieldOrgIdentifier     = "orgIdentifier"
	FieldProjectIdentifier = "projectIdentifier"

	// HTTP request info fields.
	FieldRequestMethod = "requestMethod"

	// Request metadata fields.
	FieldClientIP = "clientIP"

	// Authentication info fields.
	FieldPrincipal = "principal"
	FieldLabels    = "labels"

	// Principal fields.
	FieldType       = "type"
	FieldIdentifier = "identifier"

	// Labels fields.
	FieldUserID       = "userId"
	FieldUsername     = "username"
	FieldResourceName = "resourceName"

	// Module constant.
	ModuleHAR = "HAR"
)

// InsertUDPAuditEvent inserts an audit event into the UDP events table.
func InsertUDPAuditEvent(
	ctx context.Context,
	db dbtx.Accessor,
	principal gitnesstypes.Principal,
	resource audit.Resource,
	udpAction string,
	spacePath string,
	options ...audit.Option,
) {
	if db == nil {
		log.Ctx(ctx).Debug().Msg("skipping UDP audit event insertion: no database accessor provided")
		return
	}
	event := &audit.Event{}
	for _, opt := range options {
		opt.Apply(event)
	}

	resourceScope := parseResourceScope(spacePath)

	clientIP := event.ClientIP
	if clientIP == "" {
		clientIP = audit.GetRealIP(ctx)
	}

	requestMethod := event.RequestMethod
	if requestMethod == "" {
		requestMethod = audit.GetRequestMethod(ctx)
	}

	resourceLabels := make(map[string]interface{})
	resourceLabels[FieldResourceName] = resource.Identifier
	resourceData := resource.DataAsSlice()
	for i := 0; i < len(resourceData); i += 2 {
		if i+1 < len(resourceData) {
			resourceLabels[resourceData[i]] = resourceData[i+1]
		} else {
			log.Ctx(ctx).Warn().Msgf("odd number of resource data elements, ignoring last element: %v", resourceData[i])
		}
	}

	auditPayload := map[string]interface{}{
		FieldResourceScope: resourceScope,
		FieldHTTPRequestInfo: map[string]interface{}{
			FieldRequestMethod: requestMethod,
		},
		FieldRequestMetadata: map[string]interface{}{
			FieldClientIP: clientIP,
		},
		FieldTimestamp: time.Now().UnixMilli(),
		FieldAuthenticationInfo: map[string]interface{}{
			FieldPrincipal: map[string]interface{}{
				FieldType:       string(principal.Type),
				FieldIdentifier: principal.Email,
			},
			FieldLabels: map[string]interface{}{
				FieldUserID:   principal.UID,
				FieldUsername: principal.DisplayName,
			},
		},
		FieldModule: ModuleHAR,
		FieldResource: map[string]interface{}{
			FieldType:       string(resource.Type),
			FieldIdentifier: resource.Identifier,
			FieldLabels:     resourceLabels,
		},
		FieldAction: udpAction,
	}

	if len(event.Data) > 0 {
		auditPayload["internalInfo"] = event.Data
	}

	if event.DiffObject.OldObject != nil {
		auditPayload["oldObject"] = event.DiffObject.OldObject
	}
	if event.DiffObject.NewObject != nil {
		auditPayload["newObject"] = event.DiffObject.NewObject
	}

	payloadJSON, err := json.Marshal(auditPayload)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("failed to marshal audit payload for UDP events")
		return
	}

	const udpEventInsertQuery = `
		INSERT INTO udp_events (data_type, payload) VALUES ($1, $2)
	`

	_, err = db.ExecContext(ctx, udpEventInsertQuery, types.UDPEventTypeAudits, string(payloadJSON))
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("failed to insert audit event into UDP events table")
	}
}

// parseResourceScope parses the spacePath into resource scope components.
// Expected format: account/org/project or account/org.
func parseResourceScope(spacePath string) map[string]interface{} {
	scope := make(map[string]interface{})
	parts := splitPath(spacePath)

	if len(parts) >= 1 {
		scope[FieldAccountIdentifier] = parts[0]
	} else {
		scope[FieldAccountIdentifier] = ""
	}

	if len(parts) >= 2 {
		scope[FieldOrgIdentifier] = parts[1]
	} else {
		scope[FieldOrgIdentifier] = ""
	}

	if len(parts) >= 3 {
		scope[FieldProjectIdentifier] = parts[2]
	} else {
		scope[FieldProjectIdentifier] = ""
	}

	return scope
}

// splitPath splits a space path by '/' separator.
// Filters out empty parts to handle paths like "account//project".
func splitPath(path string) []string {
	if path == "" {
		return []string{}
	}

	parts := strings.Split(path, "/")
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	return filtered
}
