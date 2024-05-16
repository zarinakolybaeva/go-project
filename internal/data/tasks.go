package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/zarinakolybaeva/DoMake/internal/validator"
	"time"
)

type Task struct {
	ID          int64      `json:"id"`          // Unique integer ID for the task
	CreatedAt   CustomTime `json:"created_at"`  // Timestamp for when the task is added to our database
	Title       string     `json:"title"`       // Task title
	Description string     `json:"description"` //  Task description
	DueDate     CustomTime `json:"due_date"`    // Deadline or due date for the task
	Priority    string     `json:"priority"`    // Task priority (e.g., high, medium, low)
	Status      string     `json:"status"`      // Task status (e.g., to-do, in-progress, completed)
	Category    string     `json:"category"`    // Task category or project it belongs to
	UserID      int64      `json:"user_id"`     // ID of the user who created the task (for multi-user support)
	Version     int32      `json:"version"`
}

func ValidateTask(v *validator.Validator, task *Task) {
	v.Check(task.Title != "", "title", "must be provided")
	v.Check(len(task.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(task.Description != "", "description", "must be provided")
	v.Check(len(task.Description) <= 1000, "description", "must not be more than 1000 bytes long")
	v.Check(!task.DueDate.IsZero(), "due_date", "must be provided")
	v.Check(task.DueDate.Before(time.Date(2060, 1, 1, 0, 0, 0, 0, time.UTC)), "due_date", "must be before 2060")
	v.Check(task.DueDate.After(time.Date(2023, 10, 7, 0, 0, 0, 0, time.UTC)), "due_date", "must be after 2023-10-07")
	v.Check(task.Priority != "", "priority", "must be provided")
	v.Check(task.Status != "", "status", "must be provided")
	v.Check(task.Category != "", "category", "must be provided")
}

// Define a TaskModel struct type which wraps a sql.DB connection pool.
type TaskModel struct {
	DB *sql.DB
}

// Add a placeholder method for inserting a new record in the task table.
func (m TaskModel) Insert(task *Task) error {
	// Define the SQL query for inserting a new record in the task table and returning the system-generated data.
	query := `
		INSERT INTO tasks (title, description, priority, status, category, due_date)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, user_id, version`
	// Create an args slice containing the values for the placeholder parameters from the task struct.
	// Declaring this slice immediately next to our SQL query helps to make it nice
	// 		and clear *what values are being used where* in the query.
	args := []interface{}{task.Title, task.Description, task.Priority, task.Status, task.Category, task.DueDate}
	// Use the QueryRow() method to execute the SQL query on our connection pool,
	// passing in the args slice as a variadic parameter
	// and scanning the system-generated id, created_at and version values into the movie struct.
	return m.DB.QueryRow(query, args...).Scan(&task.ID, &task.CreatedAt, &task.UserID, &task.Version)
}

// Add a placeholder method for fetching a specific record from the task table.
func (m TaskModel) Get(id int64) (*Task, error) {
	// The PostgreSQL bigserial type that we're using for the movie ID starts auto-incrementing at 1 by default,
	// so we know that no task will have ID values less than that.
	// To avoid making an unnecessary database call, we take a shortcut and return an ErrRecordNotFound error straight away.
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Define the SQL query for retrieving the task data.
	query := `
		SELECT id, created_at, title, description, priority, status, category, due_date, user_id, version
		FROM tasks
		WHERE id = $1`
	// Declare a Task struct to hold the data returned by the query.
	var task Task

	// Use the context.WithTimeout() function to create a context.Context which carries a 3-second timeout deadline.
	// Note that we're using the empty context.Background() as the 'parent' context.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// Importantly, use defer to make sure that we cancel the context before the Get() method returns.
	defer cancel()

	// Use the QueryRowContext() method to execute the query, passing in the context with the deadline as the first argument.
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&task.ID,
		&task.CreatedAt,
		&task.Title,
		&task.Description,
		&task.Priority,
		&task.Status,
		&task.Category,
		&task.DueDate,
		&task.UserID,
		&task.Version,
	)
	// Handle any errors. If there was no matching task found, Scan() will return a sql.ErrNoRows error.
	// We check for this and return our custom ErrRecordNotFound error instead.
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// Otherwise, return a pointer to the Movie struct.
	return &task, nil
}

// Add a placeholder method for updating a specific record in the task table.
func (m TaskModel) Update(task *Task) error {
	// Declare the SQL query for updating the record and returning the new version number.
	query := `
		UPDATE tasks
		SET title = $1, description = $2, priority = $3, status = $4, category = $5, due_date = $6, user_id = $7, version = version + 1
		WHERE id = $8 AND version = $9
		RETURNING version`
	// Create an args slice containing the values for the placeholder parameters.
	args := []interface{}{
		task.Title,
		task.Description,
		task.Priority,
		task.Status,
		task.Category,
		task.DueDate,
		task.UserID,
		task.ID,
		task.Version, // // Add the expected task version
	}

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use QueryRowContext() and pass the context as the first argument.
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&task.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// Add a placeholder method for deleting a specific record from the task table.
func (m TaskModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if the task ID is less than 1.
	if id < 1 {
		return ErrRecordNotFound
	}
	// Construct the SQL query to delete the record.
	query := `
		DELETE FROM tasks
		WHERE id = $1`

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use ExecContext() and pass the context as the first argument.
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Call the RowsAffected() method on the sql.Result object to get the number of rows affected by the query.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If no rows were affected,
	//	we know that the tasks table didn't contain a record with the provided ID at the moment we tried to delete it.
	// In that case we return an ErrRecordNotFound error.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// Create a new GetAll() method which returns a slice of tasks.
// Although we're not using them right now, we've set this up to accept the various filter parameters as arguments.
func (t TaskModel) GetAll(title string, filters Filters) ([]*Task, Metadata, error) {
	// Update the SQL query to include the window function which counts the total (filtered) records.
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, created_at, title, description, due_date, priority, status, category, user_id, version
		FROM tasks
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		ORDER BY %s %s, id ASC
		LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// As our SQL query now has quite a few placeholder parameters,
	// let's collect the values for the placeholders in a slice.
	// Notice here how we call the limit() and offset() methods on the Filters struct to get the appropriate values
	//		for the LIMIT and OFFSET clauses.
	args := []interface{}{title, filters.limit(), filters.offset()}

	// And then pass the args slice to QueryContext() as a variadic parameter.
	rows, err := t.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err // Update this to return an empty Metadata struct.
	}

	// Importantly, defer a call to rows.Close() to ensure that the resultset is closed before GetAll() returns.
	defer rows.Close()

	// Declare a totalRecords variable.
	totalRecords := 0

	// Initialize an empty slice to hold the movie data.
	tasks := []*Task{}

	// Use rows.Next to iterate through the rows in the resultset.
	for rows.Next() {
		// Initialize an empty Movie struct to hold the data for an individual movie.
		var task Task
		// Scan the values from the row into the Movie struct. Again, note that we're
		// using the pq.Array() adapter on the genres field here.
		err := rows.Scan(
			&totalRecords, // Scan the count from the window function into totalRecords.
			&task.ID,
			&task.CreatedAt,
			&task.Title,
			&task.Description,
			&task.DueDate,
			&task.Priority,
			&task.Status,
			&task.Category,
			&task.UserID,
			&task.Version,
		)
		if err != nil {
			return nil, Metadata{}, err // Update this to return an empty Metadata struct.
		}

		// Add the Movie struct to the slice.
		tasks = append(tasks, &task)
	}

	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err // Update this to return an empty Metadata struct.
	}

	// Generate a Metadata struct, passing in the total record count and pagination
	// parameters from the client.
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	// If everything went OK, then return the slice of movies.
	return tasks, metadata, nil
}
