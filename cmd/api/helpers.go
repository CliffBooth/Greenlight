package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"greenlight.vysotsky.com/internal/validator"
)

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

type envelope map[string]interface {}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')
	for key, value := range headers {
		w.Header()[key] = value
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dist interface{}) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(dist)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		
		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badley-formed JSON (at character %d)", syntaxError.Offset)
		
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalError):
			if unmarshalError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalError.Field)
			}
		
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		
		case strings.HasPrefix(err.Error(), "json: unkown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unkown field ")
			return fmt.Errorf("body contains unkown key %s", fieldName)
		
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
		
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		
		default:
			return err
			
		}
	}

	// decoding again to check if there is anything else
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func (app *application) readString(params url.Values, key string, defaultValue string) string {
	s := params.Get(key)
	if s == "" {
		return defaultValue
	}
	return s
}

//read comma separated values from query and return as []string
func (app *application) readCSV(params url.Values, key string, defaultValue []string) []string {
	csv := params.Get(key)
	if csv == "" {
		return defaultValue
	}
	return strings.Split(csv, ",")
}

func (app *application) readInt(params url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := params.Get(key)
	if s == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be and integer value")
		return defaultValue
	}
	return i
}

