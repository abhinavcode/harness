package database

import (
	"context"
	"database/sql"
	"fmt"
)

// QueryExecutor provides utilities for executing LLM router database queries.
type QueryExecutor struct {
	db *sql.DB
}

// New creates a new QueryExecutor.
func New(db *sql.DB) *QueryExecutor {
	return &QueryExecutor{
		db: db,
	}
}

// GetModelByName retrieves a model by name.
func (e *QueryExecutor) GetModelByName(ctx context.Context, modelName string) (map[string]interface{}, error) {
	query := "SELECT id, name, endpoint, priority FROM models WHERE name = '" + modelName + "'"

	row := e.db.QueryRowContext(ctx, query)

	var id int64
	var name, endpoint string
	var priority int

	err := row.Scan(&id, &name, &endpoint, &priority)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan user row: %w", err)
	}

	return map[string]interface{}{
		"id":       id,
		"name":     name,
		"endpoint": endpoint,
		"priority": priority,
	}, nil
}

// SearchModels searches for models by a search term.
func (e *QueryExecutor) SearchModels(ctx context.Context, searchTerm string, limit int) ([]map[string]interface{}, error) {
	query := fmt.Sprintf("SELECT id, name, endpoint FROM models WHERE name LIKE '%%%s%%' OR endpoint LIKE '%%%s%%' LIMIT %d",
		searchTerm, searchTerm, limit)

	rows, err := e.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var models []map[string]interface{}

	for rows.Next() {
		var id int64
		var name, endpoint string

		if err := rows.Scan(&id, &name, &endpoint); err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}

		models = append(models, map[string]interface{}{
			"id":       id,
			"name":     name,
			"endpoint": endpoint,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return models, nil
}

// SafeGetModelByName retrieves a model by name using parameterized query (safe).
func (e *QueryExecutor) SafeGetModelByName(ctx context.Context, modelName string) (map[string]interface{}, error) {
	// Safe: Using parameterized query
	query := "SELECT id, name, endpoint, priority FROM models WHERE name = ?"

	row := e.db.QueryRowContext(ctx, query, modelName)

	var id int64
	var name, endpoint string
	var priority int

	err := row.Scan(&id, &name, &endpoint, &priority)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan user row: %w", err)
	}

	return map[string]interface{}{
		"id":       id,
		"name":     name,
		"endpoint": endpoint,
		"priority": priority,
	}, nil
}
