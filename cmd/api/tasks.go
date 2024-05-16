package main

import (
	"errors"
	"fmt"
	"github.com/zarinakolybaeva/DoMake/internal/data"
	"github.com/zarinakolybaeva/DoMake/internal/validator"
	"net/http"
)

func (app *application) createTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous struct to hold the information that we expect to be in the HTTP request body
	// (note that the field names and types in the struct are a subset of the Movie struct that we created earlier).
	// This struct will be our *target  decode destination*.
	var input struct {
		Title       string          `json:"title"`
		Description string          `json:"description"`
		DueDate     data.CustomTime `json:"due_date"`
		Priority    string          `json:"priority"`
		Status      string          `json:"status"`
		Category    string          `json:"category"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Copy the values from the input struct to a new Movie struct.
	task := &data.Task{
		Title:       input.Title,
		Description: input.Description,
		DueDate:     input.DueDate,
		Priority:    input.Priority,
		Status:      input.Status,
		Category:    input.Category,
	}

	// Initialize a new Validator.
	v := validator.New()

	// Call the ValidateTask() function and return a response containing the errors if any of the checks fail.
	if data.ValidateTask(v, task); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Call the Insert() method on our tasks model, passing in a pointer to the validated task struct.
	// This will create a record in the database and update the task struct with the system-generated information.
	err = app.models.Tasks.Insert(task)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// When sending a HTTP response, we want to include a Location header to
	//		let the client know which URL they can find the newly-created resource at.
	// We make an empty http.Header map and then use the Set() method to add a new Location header,
	// 		interpolating the system-generated ID for our new task in the URL.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/tasks/%d", task.ID))
	// Write a JSON response with a 201 Created status code, the task data in the response body, and the Location header.
	err = app.writeJSON(w, http.StatusCreated, envelope{"task": task}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// Add a showTaskHandler for the "GET /v1/task/:id" endpoint.
// For now, we retrieve the interpolated "id" parameter from the current URL and include it in a placeholder response.
func (app *application) showTaskHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Call the Get() method to fetch the data for a specific task.
	// We also need to use the errors.Is() function to check if it returns a data.ErrRecordNotFound error,
	// in which case we send a 404 Not Found response to the client.
	task, err := app.models.Tasks.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"task": task}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the task ID from the URL.
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Retrieve the task record as normal.
	task, err := app.models.Tasks.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Use pointers for the fields.
	var input struct {
		Title       *string          `json:"title"`
		Description *string          `json:"description"`
		DueDate     *data.CustomTime `json:"due_date"`
		Priority    *string          `json:"priority"`
		Status      *string          `json:"status"`
		Category    *string          `json:"category"`
	}

	// Decode the Json as normal
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// If the input.Title value is nil then we know that no corresponding "title"
	//		key/value pair was provided in the JSON request body.
	// So we move on and leave the task record unchanged.
	// Otherwise, we update the task record with the new title value.
	// Importantly, because input.Title is a now a pointer to a string,
	//		we need to dereference the pointer using the * operator to get the underlying value
	// 			before assigning it to our task record.
	if input.Title != nil {
		task.Title = *input.Title
	}
	// We also do the same for the other fields in the input struct.
	if input.Description != nil {
		task.Description = *input.Description
	}
	if input.Priority != nil {
		task.Priority = *input.Priority
	}
	if input.Status != nil {
		task.Status = *input.Status
	}
	if input.Category != nil {
		task.Category = *input.Category
	}
	if input.DueDate != nil {
		task.DueDate = *input.DueDate
	}

	// Validate the updated task record, sending the client a 422 Unprocessable Entity response if any checks fail.
	v := validator.New()
	if data.ValidateTask(v, task); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Intercept any ErrEditConflict error and call the new editConflictResponse() helper.
	err = app.models.Tasks.Update(task)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Write the updated task record in a JSON response.
	err = app.writeJSON(w, http.StatusOK, envelope{"task": task}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the task ID from the URL.
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Delete the task from the database,
	//		sending a 404 Not Found response to the client if there isn't a matching record.
	err = app.models.Tasks.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Return a 200 OK status code along with a success message.
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "task successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listTasksHandler(w http.ResponseWriter, r *http.Request) {
	// Embed the new Filters struct.
	var input struct {
		Title string
		data.Filters
	}
	// Initialize a new Validator instance.
	v := validator.New()

	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")

	// Read the page and page_size query string values into the embedded struct.
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	// Read the sort query string value into the embedded struct.
	input.Filters.Sort = app.readString(qs, "sort", "id")

	// Add the supported sort values for this endpoint to the sort safelist.
	input.Filters.SortSafelist = []string{"id", "title", "priority", "category", "-id", "-title", "-priority", "-category"}

	// Execute the validation checks on the Filters struct and send a response containing the errors if necessary.
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Accept the metadata struct as a return value.
	tasks, metadata, err := app.models.Tasks.GetAll(input.Title, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Include the metadata in the response envelope.
	err = app.writeJSON(w, http.StatusOK, envelope{"tasks": tasks, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
