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

package packages

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	usercontroller "github.com/harness/gitness/app/api/controller/user"
	"github.com/harness/gitness/app/api/render"
	"github.com/harness/gitness/app/api/usererror"
	"github.com/harness/gitness/app/auth/authn"
	"github.com/harness/gitness/app/auth/authz"
	"github.com/harness/gitness/app/services/refcache"
	corestore "github.com/harness/gitness/app/store"
	urlprovider "github.com/harness/gitness/app/url"
	"github.com/harness/gitness/registry/app/api/interfaces"
	"github.com/harness/gitness/registry/app/api/openapi/contracts/artifact"
	"github.com/harness/gitness/registry/app/dist_temp/errcode"
	"github.com/harness/gitness/registry/app/pkg"
	"github.com/harness/gitness/registry/app/pkg/commons"
	"github.com/harness/gitness/registry/app/pkg/filemanager"
	"github.com/harness/gitness/registry/app/pkg/quarantine"
	commons2 "github.com/harness/gitness/registry/app/pkg/types/commons"
	refcache2 "github.com/harness/gitness/registry/app/services/refcache"
	"github.com/harness/gitness/registry/app/storage"
	"github.com/harness/gitness/registry/app/store"
	"github.com/harness/gitness/registry/request"
	"github.com/harness/gitness/types/enum"

	"github.com/rs/zerolog/log"
)

var (
	// httpStatusCodePattern is compiled once for performance.
	httpStatusCodePattern = regexp.MustCompile(`http status code:\s*(\d+)`)

	// maxTranslateDepth prevents infinite recursion in error translation.
	maxTranslateDepth = 10
)

func NewHandler(
	registryDao store.RegistryRepository,
	downloadStatDao store.DownloadStatRepository,
	bandwidthStatDao store.BandwidthStatRepository,
	spaceStore corestore.SpaceStore, tokenStore corestore.TokenStore,
	userCtrl *usercontroller.Controller, authenticator authn.Authenticator,
	urlProvider urlprovider.Provider, authorizer authz.Authorizer, spaceFinder refcache.SpaceFinder,
	regFinder refcache2.RegistryFinder,
	fileManager filemanager.FileManager, quarantineFinder quarantine.Finder,
	packageWrapper interfaces.PackageWrapper,
) Handler {
	return &handler{
		RegistryDao:      registryDao,
		DownloadStatDao:  downloadStatDao,
		BandwidthStatDao: bandwidthStatDao,
		SpaceStore:       spaceStore,
		TokenStore:       tokenStore,
		UserCtrl:         userCtrl,
		Authenticator:    authenticator,
		URLProvider:      urlProvider,
		Authorizer:       authorizer,
		SpaceFinder:      spaceFinder,
		RegFinder:        regFinder,
		fileManager:      fileManager,
		quarantineFinder: quarantineFinder,
		PackageWrapper:   packageWrapper,
	}
}

type handler struct {
	RegistryDao      store.RegistryRepository
	DownloadStatDao  store.DownloadStatRepository
	BandwidthStatDao store.BandwidthStatRepository
	SpaceStore       corestore.SpaceStore
	TokenStore       corestore.TokenStore
	UserCtrl         *usercontroller.Controller
	Authenticator    authn.Authenticator
	URLProvider      urlprovider.Provider
	Authorizer       authz.Authorizer
	SpaceFinder      refcache.SpaceFinder
	RegFinder        refcache2.RegistryFinder
	fileManager      filemanager.FileManager
	quarantineFinder quarantine.Finder
	PackageWrapper   interfaces.PackageWrapper
}

type Handler interface {
	GetRegistryCheckAccess(
		ctx context.Context,
		r *http.Request,
		reqPermissions ...enum.Permission,
	) error
	GetArtifactInfo(r *http.Request) (pkg.ArtifactInfo, error)
	DownloadFile(w http.ResponseWriter, r *http.Request)
	TrackDownloadStats(
		ctx context.Context,
		r *http.Request,
	) error
	GetPackageArtifactInfo(r *http.Request) (pkg.PackageArtifactInfo, error)

	CheckQuarantineStatus(
		ctx context.Context,
	) error

	GetAuthenticator() authn.Authenticator
	HandleErrors2(ctx context.Context, errors errcode.Error, w http.ResponseWriter)
	HandleErrors(ctx context.Context, errors errcode.Errors, w http.ResponseWriter)
	HandleError(ctx context.Context, w http.ResponseWriter, err error)
	ServeContent(
		w http.ResponseWriter, r *http.Request, fileReader *storage.FileReader, filename string,
	)
}

type PathPackageType string

func (h *handler) GetAuthenticator() authn.Authenticator {
	return h.Authenticator
}

func (h *handler) GetRegistryCheckAccess(
	ctx context.Context,
	r *http.Request,
	reqPermissions ...enum.Permission,
) error {
	// Get artifact info from context or request
	info, err := func() (pkg.ArtifactInfo, error) {
		if pkgInfo := request.ArtifactInfoFrom(ctx); pkgInfo != nil {
			return pkgInfo.BaseArtifactInfo(), nil
		}
		return h.GetArtifactInfo(r)
	}()
	if err != nil {
		return err
	}
	return pkg.GetRegistryCheckAccess(ctx, h.Authorizer, h.SpaceFinder,
		info.ParentID, info, reqPermissions...)
}

func (h *handler) TrackDownloadStats(
	ctx context.Context,
	r *http.Request,
) error {
	info := request.ArtifactInfoFrom(r.Context()) //nolint:contextcheck
	if err := h.DownloadStatDao.CreateByRegistryIDImageAndArtifactName(ctx,
		info.BaseArtifactInfo().RegistryID, info.BaseArtifactInfo().Image, info.GetVersion()); err != nil {
		log.Ctx(ctx).Error().Msgf("failed to create download stat: %v", err.Error())
		return usererror.ErrInternal
	}
	return nil
}

func (h *handler) CheckQuarantineStatus(
	ctx context.Context,
) error {
	info := request.ArtifactInfoFrom(ctx)
	err := h.quarantineFinder.CheckArtifactQuarantineStatus(
		ctx,
		info.BaseArtifactInfo().RegistryID,
		info.BaseArtifactInfo().Image,
		info.GetVersion(),
		nil,
	)
	if err != nil {
		if errors.Is(err, usererror.ErrQuarantinedArtifact) {
			log.Ctx(ctx).Error().Msgf("Requested artifact: [%s] with "+
				"version: [%s] and filename: [%s] with registryID: [%d] is quarantined or check failed: %v",
				info.BaseArtifactInfo().Image, info.GetVersion(), info.GetFileName(),
				info.BaseArtifactInfo().RegistryID, err)
			return err
		}
		log.Ctx(ctx).Error().Msgf("Failed to check quarantine status for artifact: [%s] with "+
			"version: [%s] and filename: [%s] with registryID: [%d] with error: %v",
			info.BaseArtifactInfo().Image, info.GetVersion(), info.GetFileName(),
			info.BaseArtifactInfo().RegistryID, err)
	}
	return nil
}

func (h *handler) GetArtifactInfo(r *http.Request) (pkg.ArtifactInfo, error) {
	ctx := r.Context()
	rootIdentifier, registryIdentifier, pathPackageType, err := extractPathVars(r)
	if err != nil {
		return pkg.ArtifactInfo{}, errcode.ErrCodeInvalidRequest.WithDetail(err)
	}

	// Convert path package type to package type
	packageType, err := h.PackageWrapper.GetPackageTypeFromPathPackageType(string(pathPackageType))
	if err != nil {
		return pkg.ArtifactInfo{}, errcode.ErrCodeInvalidRequest.WithDetail(err)
	}

	rootSpaceID, err := h.SpaceStore.FindByRefCaseInsensitive(ctx, rootIdentifier)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Root spaceID not found: %s", rootIdentifier)
		return pkg.ArtifactInfo{}, usererror.NotFoundf("Root not found: %s", rootIdentifier)
	}
	rootSpace, err := h.SpaceFinder.FindByID(ctx, rootSpaceID)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Root space not found: %d", rootSpaceID)
		return pkg.ArtifactInfo{}, usererror.NotFoundf("Root not found: %s", rootIdentifier)
	}

	registry, err := h.RegFinder.FindByRootParentID(ctx, rootSpaceID, registryIdentifier)

	if err != nil {
		log.Ctx(ctx).Error().Msgf(
			"registry %s not found for root: %s. Reason: %s", registryIdentifier, rootSpace.Identifier, err,
		)
		return pkg.ArtifactInfo{}, usererror.NotFoundf("Registry not found: %s", registryIdentifier)
	}

	if registry.PackageType != artifact.PackageType(packageType) {
		return pkg.ArtifactInfo{}, usererror.NotFoundf(
			"Registry package type mismatch: %s != %s", registry.PackageType, pathPackageType,
		)
	}

	_, err = h.SpaceFinder.FindByID(r.Context(), registry.ParentID)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Parent space not found: %d", registry.ParentID)
		return pkg.ArtifactInfo{}, usererror.NotFoundf("Parent not found for registry: %s", registryIdentifier)
	}

	return pkg.ArtifactInfo{
		BaseInfo: &pkg.BaseInfo{
			RootIdentifier:  rootIdentifier,
			RootParentID:    rootSpace.ID,
			ParentID:        registry.ParentID,
			PathPackageType: artifact.PackageType(packageType),
		},
		RegIdentifier: registryIdentifier,
		RegistryID:    registry.ID,
		Registry:      *registry,
		Image:         "",
	}, nil
}

// GetUtilityMethodArtifactInfo : /pkg/{rootIdentifier}/{registryIdentifier}/{utilityMethod}...
const minPathComponents = 5

func (h *handler) GetUtilityMethodArtifactInfo(r *http.Request) (pkg.ArtifactInfo, error) {
	ctx := r.Context()
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < minPathComponents {
		return pkg.ArtifactInfo{}, errcode.ErrCodeInvalidRequest.WithMessage(fmt.Sprintf("invalid path: %s", path))
	}
	rootIdentifier := parts[2]
	registryIdentifier := parts[3]

	rootSpaceID, err := h.SpaceStore.FindByRefCaseInsensitive(ctx, rootIdentifier)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Root spaceID not found: %s", rootIdentifier)
		return pkg.ArtifactInfo{}, usererror.NotFoundf("Root not found: %s", rootIdentifier)
	}
	rootSpace, err := h.SpaceFinder.FindByID(ctx, rootSpaceID)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Root space not found: %d", rootSpaceID)
		return pkg.ArtifactInfo{}, usererror.NotFoundf("Root not found: %s", rootIdentifier)
	}

	registry, err := h.RegFinder.FindByRootParentID(ctx, rootSpaceID, registryIdentifier)

	if err != nil {
		log.Ctx(ctx).Error().Msgf(
			"registry %s not found for root: %s. Reason: %s", registryIdentifier, rootSpace.Identifier, err,
		)
		return pkg.ArtifactInfo{}, usererror.NotFoundf("Registry not found: %s", registryIdentifier)
	}

	_, err = h.SpaceFinder.FindByID(r.Context(), registry.ParentID)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Parent space not found: %d", registry.ParentID)
		return pkg.ArtifactInfo{}, usererror.NotFoundf("Parent not found for registry: %s", registryIdentifier)
	}

	return pkg.ArtifactInfo{
		BaseInfo: &pkg.BaseInfo{
			RootIdentifier: rootIdentifier,
			RootParentID:   rootSpace.ID,
			ParentID:       registry.ParentID,
		},
		RegIdentifier: registryIdentifier,
		RegistryID:    registry.ID,
		Registry:      *registry,
	}, nil
}

func (h *handler) HandleErrors2(ctx context.Context, err errcode.Error, w http.ResponseWriter) {
	if !commons.IsEmptyError(err) {
		w.WriteHeader(err.Code.Descriptor().HTTPStatusCode)
		_ = errcode.ServeJSON(w, err)
		log.Ctx(ctx).Error().Msgf("Error occurred while performing artifact action: %s", err.Message)
	}
}

// HandleErrors TODO: Improve Error Handling
// HandleErrors handles errors and writes the appropriate response to the client.
func (h *handler) HandleErrors(ctx context.Context, errs errcode.Errors, w http.ResponseWriter) {
	if !commons.IsEmpty(errs) {
		LogError(errs)
		log.Ctx(ctx).Error().Errs("errs occurred during artifact operation: ", errs).Msgf("Error occurred")
		err := errs[0]
		var e *commons.Error
		if errors.As(err, &e) {
			code := e.Status
			w.WriteHeader(code)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(errs)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msgf("Error occurred during artifact error encoding")
		}
	}
}

func (h *handler) HandleError(ctx context.Context, w http.ResponseWriter, err error) {
	if nil != err {
		log.Error().Err(err).Ctx(ctx).Msgf("error: %v", err)
		// Translate registry-specific errors, then delegate to core rendering
		translatedErr := translateRegistryError(ctx, err, 0)
		render.UserError(ctx, w, translatedErr)
		return
	}
}

func LogError(errList errcode.Errors) {
	for _, e1 := range errList {
		log.Error().Err(e1).Msgf("error: %v", e1)
	}
}

// extractPathVars extracts rootSpace, registryId, pathPackageType from the path
// Path format: /pkg/:rootSpace/:registry/:pathPackageType/...
func extractPathVars(r *http.Request) (
	rootIdentifier string,
	registry string,
	pathPackageType PathPackageType,
	err error,
) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 5 {
		return "", "", "", fmt.Errorf("invalid path: %s", path)
	}
	rootIdentifier = parts[2]
	registry = parts[3]
	pathPackageType = PathPackageType(parts[4])
	return rootIdentifier, registry, pathPackageType, nil
}

func (h *handler) ServeContent(
	w http.ResponseWriter, r *http.Request, fileReader *storage.FileReader, filename string,
) {
	if fileReader != nil {
		http.ServeContent(w, r, filename, time.Time{}, fileReader)
	}
}

func (h *handler) GetPackageArtifactInfo(r *http.Request) (pkg.PackageArtifactInfo, error) {
	info, err := h.GetUtilityMethodArtifactInfo(r)
	if err != nil {
		return nil, err
	}
	return commons2.ArtifactInfo{
		ArtifactInfo: info,
	}, nil
}

// TranslateRegistryError translates registry-specific errors to usererror.Error.
// This is exported for use by other registry handlers (OCI, Maven, Generic).
func TranslateRegistryError(ctx context.Context, err error) *usererror.Error {
	return translateRegistryError(ctx, err, 0)
}

// translateRegistryError translates registry-specific errors to usererror.Error.
// This handles errcode.Error and commons.Error types that the core translator doesn't recognize.
func translateRegistryError(ctx context.Context, err error, depth int) *usererror.Error {
	// Prevent infinite recursion
	if depth >= maxTranslateDepth {
		log.Ctx(ctx).Warn().Int("depth", depth).Msgf(
			"Maximum translation depth reached in registry error handler, returning Internal Error",
		)
		return usererror.ErrInternal
	}

	var (
		commonsError *commons.Error
		errcodeError errcode.Error
	)

	log.Ctx(ctx).Info().Err(err).Msgf("translating error to user facing error")

	switch {
	// Handle registry commons errors
	case errors.As(err, &commonsError):
		return usererror.New(commonsError.Status, commonsError.Message)

	// Handle errcode errors (from Docker registry distribution)
	case errors.As(err, &errcodeError):
		// Try to translate the wrapped detail error
		if detailErr, ok := errcodeError.Detail.(error); ok {
			translated := translateRegistryError(ctx, detailErr, depth+1)
			if translated.Message != usererror.ErrInternal.Message {
				return translated
			}
			// Extract HTTP status from error message if available
			httpStatus := extractHTTPStatusFromError(detailErr.Error())
			if httpStatus == 0 {
				httpStatus = getErrcodeHTTPStatus(errcodeError)
			}
			return usererror.New(httpStatus, detailErr.Error())
		}
		// No detail error, use errcode message
		return usererror.New(getErrcodeHTTPStatus(errcodeError), errcodeError.Message)

	// For all other errors, delegate to core translator
	default:
		log.Ctx(ctx).Warn().Err(err).Msgf(
			"Registry error handler: delegating to core translator for non-registry error",
		)
		return usererror.Translate(ctx, err)
	}
}

// getErrcodeHTTPStatus extracts HTTP status from errcode.Error, defaults to 500.
func getErrcodeHTTPStatus(errcodeError errcode.Error) int {
	httpStatus := errcodeError.Code.Descriptor().HTTPStatusCode
	if httpStatus == 0 {
		httpStatus = http.StatusInternalServerError
	}
	return httpStatus
}

// extractHTTPStatusFromError extracts HTTP status code from error messages
// with pattern "http status code: XXX". Returns 0 if not found.
func extractHTTPStatusFromError(errMsg string) int {
	matches := httpStatusCodePattern.FindStringSubmatch(errMsg)
	if len(matches) > 1 {
		if status, err := strconv.Atoi(matches[1]); err == nil {
			// Valid HTTP status codes: 1xx-5xx (100-599)
			if status >= 100 && status <= 599 {
				return status
			}
		}
	}
	return 0
}
