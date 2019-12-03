/*
 * Author: Chris Lenz <chris@rtbrick.com>
 * Copyright (c) 2016 - 2019, RtBrick, Inc.
 */

package rtbhttp

import (
	"encoding/json"
	"net/http"
)

const (
	//GET method definition
	GET = "GET"
	//PUT method definition
	PUT = "PUT"
	//POST method definition
	POST = "POST"
	//DELETE method definition
	DELETE = "DELETE"
)

// Message represents the json result message used in ctrld
type Message struct {
	Message string `json:"message"`
}

// WriteMessage writes the particular message to the response
func WriteMessage(w http.ResponseWriter, statuscode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statuscode)
	jsonEncoder := json.NewEncoder(w)
	_ = jsonEncoder.Encode(&Message{Message: message})
}

// WriteAsJSON write interface as data
func WriteAsJSON(w http.ResponseWriter, statuscode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statuscode)
	jsonEncoder := json.NewEncoder(w)
	_ = jsonEncoder.Encode(data)
}

// ReadJSON write interface as data
func ReadJSON(req *http.Request, data interface{}) error {
	decoder := json.NewDecoder(req.Body)
	return decoder.Decode(data)
}
