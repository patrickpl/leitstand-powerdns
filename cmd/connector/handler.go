/*
 * Author: Chris Lenz <chris@rtbrick.com>
 * Copyright (c) 2016 - 2019, RtBrick, Inc.
 */

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/leitstand/leitstand-powerdns/pkg/rtbhttp"

	"github.com/mittwald/go-powerdns/apis/zones"
)

// @Summary webhook listener for leitstand-inventory
// @Description listens to events from leitstand-inventory
// @Description **Characteristics:**
// @Description * Operation: **synchronous**
// @Accept  json
// @Produce  json
// @Param body body main.inventoryDNSEvent true "body"
// @Success 204 "No Content"
// @Failure 422 {object} rtbhttp.Message
// @Failure 500 {object} rtbhttp.Message
// @Router /api/v1/events [POST]
func (app *application) rbmsEvent(res http.ResponseWriter, req *http.Request) {
	requestBody := &inventoryDNSEvent{}
	err := rtbhttp.ReadJSON(req, requestBody)
	if err != nil {
		log.Printf("Error: %s\n", err)
		rtbhttp.WriteMessage(res, http.StatusBadRequest, fmt.Sprintf("error %v", err))
		return
	}
	recordSet := requestBody.RecordSet
	if recordSet.WithDrawName != nil {
		err = app.powerdnsClient.Zones().RemoveRecordSetFromZone(req.Context(), app.config.PowerdnsServerID, recordSet.ZoneName, *recordSet.WithDrawName, recordSet.Type)
		if err != nil {
			message := fmt.Sprintf("error %v", err)
			log.Printf("Error: %s\n", message)
			rtbhttp.WriteMessage(res, http.StatusInternalServerError, message)
			return
		}
	} else if recordSet.Name != nil {
		rrset := zones.ResourceRecordSet{
			Name: *recordSet.Name,
			Type: recordSet.Type,
			TTL:  recordSet.TTL,
		}
		for _, record := range recordSet.Records {
			rrset.Records = append(rrset.Records, zones.Record{
				Content:  record.Value,
				Disabled: record.Disabled,
				SetPTR:   record.SetPTR,
			})
		}
		err = app.powerdnsClient.Zones().AddRecordSetToZone(req.Context(), app.config.PowerdnsServerID, recordSet.ZoneName, rrset)
		if err != nil {
			message := fmt.Sprintf("error %v", err)
			log.Printf("Error: %s\n", message)
			rtbhttp.WriteMessage(res, http.StatusInternalServerError, message)
			return
		}
	} else {
		message := fmt.Sprintf("No changetype matched")
		log.Printf("Error: %s\n", message)
		rtbhttp.WriteMessage(res, http.StatusBadRequest, message)
		return
	}
	res.WriteHeader(http.StatusNoContent)
}
