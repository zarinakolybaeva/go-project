package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"fmt"
	"github.com/zarinakolybaeva/DoMake/internal/validator"
)

type Category struct {
	ID          int64      `json:"id"`
	CreatedAt   CustomTime `json:"created_at"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
}

// ValidateCategory validates the category data.
func ValidateCategory(v *validator.Validator, category *Category) {
	v.Check(category.Name != "", "name", "must be provided")
	v.Check(len(category.Name) <= 100, "name", "must not be more than 100 bytes long")
	v.Check(category.Description != "", "description", "must be provided")
	v.Check(len(category.Description) <= 500, "description", "must not be more than 500 bytes long")
}

type CategoryModel struct {
	DB *sql.DB
}

// Insert a new record in the categories table.
func (m CategoryModel) Insert(category *Category) error {
	query := `
		INSERT INTO categories (name, description)
		VALUES ($1, $2)
		RETURNING id, created_at`
	args := []interface{}{category.Name, category.Description}

	return m.DB.QueryRow(query, args...).Scan(&category.ID, &category.CreatedAt)
}

// Retrieve a specific record from the categories table.
func (m CategoryModel) Get(id int64) (*Category, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `
		SELECT id, created_at, name, description
		FROM categories
		WHERE id = $1`
	var category Category

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&category.ID,
		&category.CreatedAt,
		&category.Name,
		&category.Description,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &category, nil
}

// Update a specific record in the categories table.
func (m CategoryModel) Update(category *Category) error {
	query := `
		UPDATE categories
		SET name = $1, description = $2
		WHERE id = $3`
	args := []interface{}{
		category.Name,
		category.Description,
		category.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

// Delete a specific record from the categories table.
func (m CategoryModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `
		DELETE FROM categories
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// GetAll retrieves all categories with pagination support.
func (m CategoryModel) GetAll(filters Filters) ([]*Category, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, created_at, name, description
		FROM categories
		ORDER BY %s %s, id ASC
		LIMIT $1 OFFSET $2`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	categories := []*Category{}

	for rows.Next() {
		var category Category
		err := rows.Scan(
			&totalRecords,
			&category.ID,
			&category.CreatedAt,
			&category.Name,
			&category.Description,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		categories = append(categories, &category)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return categories, metadata, nil
}