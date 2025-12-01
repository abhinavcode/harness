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

package database

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/harness/gitness/app/api/request"
	"github.com/harness/gitness/registry/app/api/openapi/contracts/artifact"
	"github.com/harness/gitness/registry/app/pkg/commons"
	"github.com/harness/gitness/registry/app/store"
	"github.com/harness/gitness/registry/app/store/database/util"
	"github.com/harness/gitness/registry/types"
	gitness_store "github.com/harness/gitness/store"
	databaseg "github.com/harness/gitness/store/database"
	"github.com/harness/gitness/store/database/dbtx"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type ImageDao struct {
	db *sqlx.DB
}

func NewImageDao(db *sqlx.DB) store.ImageRepository {
	return &ImageDao{
		db: db,
	}
}

type imageDB struct {
	ID           int64                  `db:"image_id"`
	UUID         string                 `db:"image_uuid"`
	Name         string                 `db:"image_name"`
	ArtifactType *artifact.ArtifactType `db:"image_type"`
	RegistryID   int64                  `db:"image_registry_id"`
	Labels       sql.NullString         `db:"image_labels"`
	Enabled      bool                   `db:"image_enabled"`
	CreatedAt    int64                  `db:"image_created_at"`
	UpdatedAt    int64                  `db:"image_updated_at"`
	CreatedBy    int64                  `db:"image_created_by"`
	UpdatedBy    int64                  `db:"image_updated_by"`
	DeletedAt    *int64                 `db:"image_deleted_at"`
	DeletedBy    *int64                 `db:"image_deleted_by"`
}

// imageWithParentDB is used for queries that JOIN with registries table
type imageWithParentDB struct {
	imageDB                  // Embed base struct
	RegistryDeletedAt *int64 `db:"registry_deleted_at"` // CASCADE from parent registry
}

type imageLabelDB struct {
	Labels sql.NullString `db:"labels"`
}

func (i ImageDao) Get(ctx context.Context, id int64, softDeleteFilter types.SoftDeleteFilter) (*types.Image, error) {
	q := databaseg.Builder.Select(util.ArrToStringByDelimiter(util.GetDBTagsFromStruct(imageWithParentDB{}), ",")).
		From("images i").
		Join("registries r ON i.image_registry_id = r.registry_id").
		Where("i.image_id = ?", id)

	switch softDeleteFilter {
	case types.SoftDeleteFilterExcludeDeleted:
		q = q.Where("i.image_deleted_at IS NULL").
			Where("r.registry_deleted_at IS NULL")
	case types.SoftDeleteFilterOnlyDeleted:
		q = q.Where("(i.image_deleted_at IS NOT NULL OR r.registry_deleted_at IS NOT NULL)")
	case types.SoftDeleteFilterAll:
		// No filtering
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to convert query to sql")
	}

	db := dbtx.GetAccessor(ctx, i.db)

	dst := new(imageWithParentDB)
	if err = db.GetContext(ctx, dst, sql, args...); err != nil {
		return nil, databaseg.ProcessSQLErrorf(ctx, err, "Failed to get image")
	}
	return i.mapImageWithParent(ctx, dst)
}

func (i ImageDao) DeleteByImageNameAndRegID(ctx context.Context, regID int64, image string) (err error) {
	stmt := databaseg.Builder.Delete("images").
		Where("image_name = ? AND image_registry_id = ?", image, regID)

	sql, args, err := stmt.ToSql()
	if err != nil {
		return errors.Wrap(err, "Failed to convert query to sql")
	}

	db := dbtx.GetAccessor(ctx, i.db)

	_, err = db.ExecContext(ctx, sql, args...)
	if err != nil {
		return databaseg.ProcessSQLErrorf(ctx, err, "the delete query failed")
	}

	return nil
}

func (i ImageDao) DeleteByImageNameIfNoLinkedArtifacts(
	ctx context.Context, regID int64, image string,
) error {
	stmt := databaseg.Builder.Delete("images").
		Where("image_name = ? AND image_registry_id = ?", image, regID).
		Where("NOT EXISTS ( SELECT 1 FROM artifacts WHERE artifacts.artifact_image_id = images.image_id )")

	sql, args, err := stmt.ToSql()
	if err != nil {
		return errors.Wrap(err, "Failed to convert query to sql")
	}

	db := dbtx.GetAccessor(ctx, i.db)

	_, err = db.ExecContext(ctx, sql, args...)
	if err != nil {
		return databaseg.ProcessSQLErrorf(ctx, err, "the delete query failed")
	}

	return nil
}

// SoftDeleteByImageNameAndRegID marks an image as deleted (soft delete).
func (i ImageDao) SoftDeleteByImageNameAndRegID(ctx context.Context, regID int64, image string) error {
	session, _ := request.AuthSessionFrom(ctx)
	now := time.Now().UnixMilli()
	userID := session.Principal.ID

	log.Ctx(ctx).Info().Msgf("SoftDelete image: regID=%d, image=%s, userID=%d", regID, image, userID)

	stmt := databaseg.Builder.
		Update("images").
		Set("image_deleted_at", now).
		Set("image_deleted_by", userID).
		Where(sq.Eq{
			"image_registry_id": regID,
			"image_name":        image,
		}).
		Where("image_deleted_at IS NULL")

	sql, args, err := stmt.ToSql()
	if err != nil {
		return databaseg.ProcessSQLErrorf(ctx, err, "Failed to build soft delete query")
	}

	log.Ctx(ctx).Debug().Msgf("Executing soft delete image SQL: %s, args: %v", sql, args)

	db := dbtx.GetAccessor(ctx, i.db)
	result, err := db.ExecContext(ctx, sql, args...)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("Failed to execute soft delete for image: %s", image)
		return databaseg.ProcessSQLErrorf(ctx, err, "Failed to soft delete image")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return databaseg.ProcessSQLErrorf(ctx, err, "Failed to get rows affected")
	}

	if rowsAffected == 0 {
		log.Ctx(ctx).Warn().Msgf("Soft delete affected 0 rows for image: %s (already deleted or not found)", image)
		return databaseg.ProcessSQLErrorf(ctx, nil, "Image not found or already deleted")
	}

	log.Ctx(ctx).Info().Msgf("Successfully soft deleted image: %s, rows affected: %d", image, rowsAffected)
	return nil
}

// RestoreByImageNameAndRegID restores a soft-deleted image.
func (i ImageDao) RestoreByImageNameAndRegID(ctx context.Context, regID int64, image string) error {
	session, _ := request.AuthSessionFrom(ctx)
	userID := session.Principal.ID

	stmt := databaseg.Builder.
		Update("images").
		Set("image_deleted_at", nil).
		Set("image_deleted_by", nil).
		Set("image_updated_at", time.Now().UnixMilli()).
		Set("image_updated_by", userID).
		Where("image_registry_id = ? AND image_name = ? AND image_deleted_at IS NOT NULL", regID, image)

	sql, args, err := stmt.ToSql()
	if err != nil {
		return databaseg.ProcessSQLErrorf(ctx, err, "Failed to build restore query")
	}

	db := dbtx.GetAccessor(ctx, i.db)
	result, err := db.ExecContext(ctx, sql, args...)
	if err != nil {
		return databaseg.ProcessSQLErrorf(ctx, err, "Failed to restore image")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return databaseg.ProcessSQLErrorf(ctx, err, "Failed to get rows affected")
	}

	if rowsAffected == 0 {
		return databaseg.ProcessSQLErrorf(ctx, nil, "Image not found or not deleted")
	}

	return nil
}

func (i ImageDao) GetByName(ctx context.Context, registryID int64, name string, softDeleteFilter types.SoftDeleteFilter) (*types.Image, error) {

	q := databaseg.Builder.Select(util.ArrToStringByDelimiter(util.GetDBTagsFromStruct(imageWithParentDB{}), ",")).
		From("images i").
		Join("registries r ON i.image_registry_id = r.registry_id").
		Where("i.image_registry_id = ? AND i.image_name = ? AND i.image_type IS NULL", registryID, name)

	switch softDeleteFilter {
	case types.SoftDeleteFilterExcludeDeleted:
		q = q.Where("i.image_deleted_at IS NULL").
			Where("r.registry_deleted_at IS NULL")
	case types.SoftDeleteFilterOnlyDeleted:
		q = q.Where("(i.image_deleted_at IS NOT NULL OR r.registry_deleted_at IS NOT NULL)")
	case types.SoftDeleteFilterAll:
		// No filtering
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to convert query to sql")
	}

	db := dbtx.GetAccessor(ctx, i.db)

	dst := new(imageWithParentDB)
	if err = db.GetContext(ctx, dst, sql, args...); err != nil {
		return nil, databaseg.ProcessSQLErrorf(ctx, err, "Failed to get image")
	}
	return i.mapImageWithParent(ctx, dst)
}

func (i ImageDao) GetByNameAndType(
	ctx context.Context, registryID int64,
	name string, artifactType *artifact.ArtifactType, softDeleteFilter types.SoftDeleteFilter,
) (*types.Image, error) {
	if artifactType == nil {
		return i.GetByName(ctx, registryID, name, softDeleteFilter)
	}

	q := databaseg.Builder.Select(util.ArrToStringByDelimiter(util.GetDBTagsFromStruct(imageWithParentDB{}), ",")).
		From("images i").
		Join("registries r ON i.image_registry_id = r.registry_id").
		Where("i.image_registry_id = ? AND i.image_name = ?", registryID, name).
		Where("i.image_type = ?", *artifactType)

	switch softDeleteFilter {
	case types.SoftDeleteFilterExcludeDeleted:
		q = q.Where("i.image_deleted_at IS NULL").
			Where("r.registry_deleted_at IS NULL")
	case types.SoftDeleteFilterOnlyDeleted:
		q = q.Where("(i.image_deleted_at IS NOT NULL OR r.registry_deleted_at IS NOT NULL)")
	case types.SoftDeleteFilterAll:
		// No filtering
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to convert query to sql")
	}

	db := dbtx.GetAccessor(ctx, i.db)

	dst := new(imageWithParentDB)
	if err = db.GetContext(ctx, dst, sql, args...); err != nil {
		return nil, databaseg.ProcessSQLErrorf(ctx, err, "Failed to get image")
	}
	return i.mapImageWithParent(ctx, dst)
}

func (i ImageDao) CreateOrUpdate(ctx context.Context, image *types.Image) error {
	if commons.IsEmpty(image.Name) {
		return errors.New("package/image name is empty")
	}
	var conflictCondition string
	if image.ArtifactType == nil {
		conflictCondition = ` ON CONFLICT (image_registry_id, image_name) WHERE image_type IS NULL `
	} else {
		conflictCondition = ` ON CONFLICT (image_registry_id, image_name, image_type) WHERE image_type IS NOT NULL `
	}
	var sqlQuery = `
		INSERT INTO images ( 
		         image_registry_id
				,image_name
				,image_type
				,image_enabled
				,image_created_at
				,image_updated_at
				,image_created_by
				,image_updated_by
				,image_uuid
		    ) VALUES (
						 :image_registry_id
						,:image_name
						,:image_type
						,:image_enabled
						,:image_created_at
						,:image_updated_at
						,:image_created_by
						,:image_updated_by
					    ,:image_uuid
		    ) 
           ` + conflictCondition + `
		    DO UPDATE SET
			   image_enabled = :image_enabled
            RETURNING image_id`

	db := dbtx.GetAccessor(ctx, i.db)
	query, arg, err := db.BindNamed(sqlQuery, i.mapToInternalImage(ctx, image))
	if err != nil {
		return databaseg.ProcessSQLErrorf(ctx, err, "Failed to bind image object")
	}

	if err = db.QueryRowContext(ctx, query, arg...).Scan(&image.ID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return databaseg.ProcessSQLErrorf(ctx, err, "Insert query failed")
	}
	return nil
}

func (i ImageDao) GetLabelsByParentIDAndRepo(
	ctx context.Context, parentID int64, repo string,
	limit int, offset int, search string, softDeleteFilter types.SoftDeleteFilter,
) (labels []string, err error) {
	q := databaseg.Builder.Select("a.image_labels as labels").
		From("images a").
		Join("registries r ON r.registry_id = a.image_registry_id").
		Where("r.registry_parent_id = ? AND r.registry_name = ?", parentID, repo)

	if search != "" {
		q = q.Where("a.image_labels LIKE ?", "%"+search+"%")
	}

	switch softDeleteFilter {
	case types.SoftDeleteFilterExcludeDeleted:
		q = q.Where("a.image_deleted_at IS NULL").Where("r.registry_deleted_at IS NULL")
	case types.SoftDeleteFilterOnlyDeleted:
		q = q.Where("(a.image_deleted_at IS NOT NULL OR r.registry_deleted_at IS NOT NULL)")
	case types.SoftDeleteFilterAll:
		// No filtering
	}

	q = q.OrderBy("a.image_labels ASC").
		Limit(util.SafeIntToUInt64(limit)).Offset(util.SafeIntToUInt64(offset))

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to convert query to sql")
	}

	dst := []*imageLabelDB{}

	db := dbtx.GetAccessor(ctx, i.db)

	if err = db.SelectContext(ctx, &dst, sql, args...); err != nil {
		return nil, databaseg.ProcessSQLErrorf(ctx, err, "Failed to get artifact labels")
	}

	return i.mapToImageLabels(dst), nil
}

func (i ImageDao) CountLabelsByParentIDAndRepo(
	ctx context.Context, parentID int64, repo,
	search string, softDeleteFilter types.SoftDeleteFilter,
) (count int64, err error) {
	q := databaseg.Builder.Select("a.image_labels as labels").
		From("images a").
		Join("registries r ON r.registry_id = a.image_registry_id").
		Where("r.registry_parent_id = ? AND r.registry_name = ?", parentID, repo)

	if search != "" {
		q = q.Where("a.image_labels LIKE ?", "%"+search+"%")
	}

	switch softDeleteFilter {
	case types.SoftDeleteFilterExcludeDeleted:
		q = q.Where("a.image_deleted_at IS NULL").Where("r.registry_deleted_at IS NULL")
	case types.SoftDeleteFilterOnlyDeleted:
		q = q.Where("(a.image_deleted_at IS NOT NULL OR r.registry_deleted_at IS NOT NULL)")
	case types.SoftDeleteFilterAll:
		// No filtering
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return -1, errors.Wrap(err, "Failed to convert query to sql")
	}

	db := dbtx.GetAccessor(ctx, i.db)

	dst := []*imageLabelDB{}

	if err = db.SelectContext(ctx, &dst, sql, args...); err != nil {
		return -1, databaseg.ProcessSQLErrorf(ctx, err, "Failed to get artifact labels")
	}

	return int64(len(dst)), nil
}

func (i ImageDao) GetByRepoAndName(
	ctx context.Context, parentID int64,
	repo string, name string, softDeleteFilter types.SoftDeleteFilter,
) (*types.Image, error) {
	q := databaseg.Builder.Select(util.ArrToStringByDelimiter(util.GetDBTagsFromStruct(imageWithParentDB{}), ",")).
		From("images a").
		Join(" registries r ON r.registry_id = a.image_registry_id").
		Where("r.registry_parent_id = ? AND r.registry_name = ? AND a.image_name = ?",
			parentID, repo, name)

	switch softDeleteFilter {
	case types.SoftDeleteFilterExcludeDeleted:
		q = q.Where("a.image_deleted_at IS NULL").
			Where("r.registry_deleted_at IS NULL")
	case types.SoftDeleteFilterOnlyDeleted:
		q = q.Where("(a.image_deleted_at IS NOT NULL OR r.registry_deleted_at IS NOT NULL)")
	case types.SoftDeleteFilterAll:
		// No filtering
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to convert query to sql")
	}

	db := dbtx.GetAccessor(ctx, i.db)

	dst := new(imageWithParentDB)
	if err = db.GetContext(ctx, dst, sql, args...); err != nil {
		return nil, databaseg.ProcessSQLErrorf(ctx, err, "Failed to get artifact")
	}
	return i.mapImageWithParent(ctx, dst)
}

func (i ImageDao) Update(ctx context.Context, image *types.Image) (err error) {
	var sqlQuery = " UPDATE images SET " + util.GetSetDBKeys(imageDB{}, "image_id", "image_uuid") +
		" WHERE image_id = :image_id "

	dbImage := i.mapToInternalImage(ctx, image)

	// update Version (used for optimistic locking) and Updated time
	dbImage.UpdatedAt = time.Now().UnixMilli()

	db := dbtx.GetAccessor(ctx, i.db)

	query, arg, err := db.BindNamed(sqlQuery, dbImage)
	if err != nil {
		return databaseg.ProcessSQLErrorf(ctx, err, "Failed to bind images object")
	}

	result, err := db.ExecContext(ctx, query, arg...)
	if err != nil {
		return databaseg.ProcessSQLErrorf(ctx, err, "Failed to update images")
	}

	count, err := result.RowsAffected()
	if err != nil {
		return databaseg.ProcessSQLErrorf(ctx, err, "Failed to get number of updated rows")
	}

	if count == 0 {
		return gitness_store.ErrVersionConflict
	}

	return nil
}

func (i ImageDao) UpdateStatus(ctx context.Context, image *types.Image) (err error) {
	q := databaseg.Builder.Update("images").
		Set("image_enabled", image.Enabled).
		Set("image_updated_at", time.Now().UnixMilli()).
		Where("image_registry_id = ? AND image_name = ?",
			image.RegistryID, image.Name)

	sql, args, err := q.ToSql()

	if err != nil {
		return databaseg.ProcessSQLErrorf(ctx, err, "Failed to bind images object")
	}

	result, err := i.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return databaseg.ProcessSQLErrorf(ctx, err, "Failed to update images")
	}

	count, err := result.RowsAffected()
	if err != nil {
		return databaseg.ProcessSQLErrorf(ctx, err, "Failed to get number of updated rows")
	}

	if count == 0 {
		return gitness_store.ErrVersionConflict
	}

	return nil
}

func (i ImageDao) DuplicateImage(ctx context.Context, sourceImage *types.Image, targetRegistryID int64) (
	*types.Image,
	error,
) {
	targetImage := &types.Image{
		Name:         sourceImage.Name,
		ArtifactType: sourceImage.ArtifactType,
		RegistryID:   targetRegistryID,
		Enabled:      sourceImage.Enabled,
	}

	err := i.CreateOrUpdate(ctx, targetImage)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to duplicate image")
	}

	return targetImage, nil
}

func (i ImageDao) mapToInternalImage(ctx context.Context, in *types.Image) *imageDB {
	session, _ := request.AuthSessionFrom(ctx)

	if in.CreatedAt.IsZero() {
		in.CreatedAt = time.Now()
	}
	if in.CreatedBy == 0 {
		in.CreatedBy = session.Principal.ID
	}

	in.UpdatedAt = time.Now()
	in.UpdatedBy = session.Principal.ID

	if in.UUID == "" {
		in.UUID = uuid.NewString()
	}

	sort.Strings(in.Labels)

	return &imageDB{
		ID:           in.ID,
		UUID:         in.UUID,
		Name:         in.Name,
		ArtifactType: in.ArtifactType,
		RegistryID:   in.RegistryID,
		Labels:       util.GetEmptySQLString(util.ArrToString(in.Labels)),
		Enabled:      in.Enabled,
		CreatedAt:    in.CreatedAt.UnixMilli(),
		UpdatedAt:    in.UpdatedAt.UnixMilli(),
		CreatedBy:    in.CreatedBy,
		UpdatedBy:    in.UpdatedBy,
	}
}

// mapImageWithParent maps imageWithParentDB (from JOIN queries) and computes earliest deletedAt
func (i ImageDao) mapImageWithParent(_ context.Context, dst *imageWithParentDB) (*types.Image, error) {
	createdBy := dst.CreatedBy
	updatedBy := dst.UpdatedBy

	// Compute DeletedAt and IsDeleted with cascade logic
	// deletedAt should be set to the earliest timestamp among image or registry
	var deletedAt *time.Time
	isDeleted := false

	// Collect all non-null deleted_at timestamps
	var timestamps []*int64
	if dst.DeletedAt != nil {
		timestamps = append(timestamps, dst.DeletedAt)
	}
	if dst.RegistryDeletedAt != nil {
		timestamps = append(timestamps, dst.RegistryDeletedAt)
	}

	// If any entity is deleted, set isDeleted and find earliest timestamp
	if len(timestamps) > 0 {
		isDeleted = true
		// Find the earliest (minimum) timestamp
		earliestTimestamp := timestamps[0]
		for _, ts := range timestamps[1:] {
			if *ts < *earliestTimestamp {
				earliestTimestamp = ts
			}
		}
		t := time.UnixMilli(*earliestTimestamp)
		deletedAt = &t
	}

	return &types.Image{
		ID:           dst.ID,
		UUID:         dst.UUID,
		Name:         dst.Name,
		ArtifactType: dst.ArtifactType,
		RegistryID:   dst.RegistryID,
		Labels:       util.StringToArr(dst.Labels.String),
		Enabled:      dst.Enabled,
		CreatedAt:    time.UnixMilli(dst.CreatedAt),
		UpdatedAt:    time.UnixMilli(dst.UpdatedAt),
		CreatedBy:    createdBy,
		UpdatedBy:    updatedBy,
		DeletedAt:    deletedAt,
		DeletedBy:    dst.DeletedBy,
		IsDeleted:    isDeleted, // CASCADE: deleted if image OR registry is deleted
	}, nil
}

func (i ImageDao) mapToImageLabels(dst []*imageLabelDB) []string {
	elements := make(map[string]bool)
	res := []string{}
	for _, labels := range dst {
		elements, res = i.mapToImageLabel(elements, res, labels)
	}
	return res
}

func (i ImageDao) mapToImageLabel(
	elements map[string]bool, res []string,
	dst *imageLabelDB,
) (map[string]bool, []string) {
	if dst == nil {
		return elements, res
	}
	labels := util.StringToArr(dst.Labels.String)
	for _, label := range labels {
		if !elements[label] {
			elements[label] = true
			res = append(res, label)
		}
	}
	return elements, res
}

// Purge permanently deletes soft-deleted images older than the given timestamp.
// Returns the number of images deleted.
func (i ImageDao) Purge(ctx context.Context, accountID string, deletedBeforeOrAt int64) (int64, error) {
	// Delete images that belong to registries of the specified account
	// Using JOIN for better performance with large datasets
	sql := `DELETE FROM images
		WHERE image_id IN (
			SELECT i.image_id 
			FROM images i
			INNER JOIN registries r ON i.image_registry_id = r.registry_id
			WHERE r.registry_account_identifier = $1
			  AND i.image_deleted_at IS NOT NULL
			  AND i.image_deleted_at <= $2
		)`

	db := dbtx.GetAccessor(ctx, i.db)
	result, err := db.ExecContext(ctx, sql, accountID, deletedBeforeOrAt)
	if err != nil {
		return 0, databaseg.ProcessSQLErrorf(ctx, err, "failed to purge images")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
