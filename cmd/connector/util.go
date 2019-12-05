/*
 * Author: Chris Lenz <chris@rtbrick.com>
 * Copyright (c) 2016 - 2019, RtBrick, Inc.
 */

package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func validateAndGetVariableFromPath(w http.ResponseWriter, req *http.Request, variableName string) (string, bool) {
	vars := mux.Vars(req)
	name, set := vars[variableName]
	if !set {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error: %s: There is an misconfiguration in the path variables!\n", variableName)
		return name, false
	}
	return name, true
}
func validateAndGetConfigTypeFromPath(w http.ResponseWriter, req *http.Request) (string, bool) {
	return validateAndGetVariableFromPath(w, req, "event_name")
}
