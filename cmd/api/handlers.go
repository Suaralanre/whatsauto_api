package main

import (
	"fmt"
	"net/http"
)

func (app *application) NewPatientFormHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "New Patient Form")
}
