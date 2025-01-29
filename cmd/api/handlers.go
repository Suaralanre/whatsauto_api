package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Suaralanre/whatsauto_api/internal/validator"
)

func (app *application) NewPatientFormHandler(w http.ResponseWriter, r *http.Request) {
	var input PatientForm

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.New()

	v.Check(input.Title != "", "title", "must be provided")
	v.Check(input.FirstName != "", "firstname", "must be provided")
	v.Check(input.Whatsapp != "", "whatsapp", "must be provided")
	v.Check(strings.HasPrefix(input.Whatsapp, "+"), "whatsapp", "Whatsapp number must start with a country code.")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	
	// Send whatsapp message to phone number
	fmt.Fprintf(w, "New Patient: %s %s sent to number %s",input.Title, input.FirstName, input.Whatsapp)
}

