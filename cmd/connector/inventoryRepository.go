/*
 * Author: Chris Lenz <chris@rtbrick.com>
 * Copyright (c) 2016 - 2019, RtBrick, Inc.
 */

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	ElementDnsRecordSetModifiedEvent = "ElementDnsRecordSetModifiedEvent"
	DnsZoneCreatedEvent              = "DnsZoneCreatedEvent"
	DnsZoneRemovedEvent              = "DnsZoneRemovedEvent"
	reRegistrationTime               = time.Second * 60
	httpTimeout                      = time.Second * 10
)

type webHookRequest struct {
	HookID      string `json:"hook_id"`
	HookName    string `json:"hook_name"`
	Description string `json:"description"`
	TopicName   string `json:"topic_name"`
	Selector    string `json:"selector"`
	Endpoint    string `json:"endpoint"`
	BatchSizes  int    `json:"batch_sizes"`
	Method      string `json:"method"`
}
type inventoryRepository struct {
	netClient *http.Client

	registerWebHookStatusCode int
	config                    *Config
}

func newRbmsRepository(config *Config) *inventoryRepository {

	return &inventoryRepository{config: config, netClient: &http.Client{
		Timeout: httpTimeout,
	}}
}

// registerWebHook Registers every 60 Seconds this service as webhook in the inventoy
func (r *inventoryRepository) registerWebHook(ctx context.Context) {
	r.register()
	for {
		select {
		case <-time.After(reRegistrationTime):
			r.register()
		case <-ctx.Done():
			return
		}
	}
}

func (r *inventoryRepository) register() {
	requestData := webHookRequest{
		HookID:      r.config.WebHookID,
		HookName:    "powerdns",
		Description: "Forward DNS changes to PowerDNS connector.",
		TopicName:   "element",
		Selector:    fmt.Sprintf("%s|%s|%s", ElementDnsRecordSetModifiedEvent, DnsZoneCreatedEvent, DnsZoneRemovedEvent),
		Endpoint:    fmt.Sprintf("%s/api/v1/events/{{event_name}}", r.config.ExternalURL),
		BatchSizes:  1,
		Method:      "POST",
	}
	uri := fmt.Sprintf("%s/webhooks/%s", r.config.InventoyRestRestURL, r.config.WebHookID)
	body := new(bytes.Buffer)
	_ = json.NewEncoder(body).Encode(requestData)

	req, err := http.NewRequest(http.MethodPut, uri, body)
	if err != nil {
		fmt.Printf("http.NewRequest() failed with '%s'\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if r.config.InventoryAuthorizationHeader != "" {
		req.Header.Set("Authorization", r.config.InventoryAuthorizationHeader)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := r.netClient.Do(req)
	if err != nil {
		fmt.Printf("client.Do() failed with '%s'\n", err)
		return
	}
	defer resp.Body.Close()
	if r.registerWebHookStatusCode != resp.StatusCode {
		fmt.Printf("Response status code: %d\n", resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("Response body: %s\n", string(body))
	}
	r.registerWebHookStatusCode = resp.StatusCode
}
