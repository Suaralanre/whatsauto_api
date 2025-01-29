package main

import (
	"fmt"
	"net/http"

	"github.com/Suaralanre/whatsauto_api/internal/utils"
)

var environment = utils.GetEnv("ENVIRONMENT", "development")
var version = utils.GetEnv("VERSION", "1.0.0")

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Status: OK")
	fmt.Fprintf(w, "environment: %s\n", environment)
	fmt.Fprintf(w, "version: %s\n", version)
}
