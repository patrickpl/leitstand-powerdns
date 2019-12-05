/*
 * Author: Chris Lenz <chris@rtbrick.com>
 * Copyright (c) 2016 - 2019, RtBrick, Inc.
 */

package main

import (
	"net/http"

	"github.com/leitstand/leitstand-powerdns/pkg/rtbhttp"

	"github.com/gorilla/mux"
)

func (app *application) routes() http.Handler {
	router := mux.NewRouter()
	router.Path("/api/v1/events/{event_name}").Methods(rtbhttp.POST).HandlerFunc(app.rbmsEvent)
	return app.logRequest(router)
}
