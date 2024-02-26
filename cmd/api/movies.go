package main

import (
	"fmt"
	"net/http"
	"time"

	"greenlight.vysotsky.com/internal/data"
	"greenlight.vysotsky.com/internal/validator"
)

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r) 
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie := data.Movie {
		ID: id,
		CreatedAt: time.Now(),
		Title: "Casablanka",
		Year: 1942,
		Runtime: 102,
		Genres: []string {"drama", "romance", "war"},
		Version: 1,
	}

	err = app.writeJSON(w, 200, envelope{"movie": movie}, nil)
	if err != nil {
		app.logger.Println("error:", err)
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Title string `json:title`
		Year int32 `json:year`
		Runtime data.Runtime `json:runtime`
		Genres []string `json:genres`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	movie := &data.Movie {
		Title: input.Title,
		Year: input.Year,
		Runtime: input.Runtime,
		Genres: input.Genres,
	}
	data.ValidateMovie(v, movie)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	fmt.Fprintf(w, "%+v\n", input)
}