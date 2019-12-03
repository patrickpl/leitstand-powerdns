/*
 * Author: Chris Lenz <chris@rtbrick.com>
 * Copyright (c) 2016 - 2019, RtBrick, Inc.
 */

package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"regexp"
)

var (
	variableMatcher = regexp.MustCompile(`(.*?):(.*)`)
)

// Config is the configuration of this service
type Config struct {
	// ExternalURL The URL under which this service is externally reachable
	// (for example, if this services is served via a reverse proxy).
	// Used for generating relative and absolute links back to this service itself.
	// (e.g.: http://localhost:19991)
	ExternalURL string `json:"external_url"`
	// PowerdnsServerID the id of the server, see https://doc.powerdns.com/authoritative/http-api/server.html
	// (e.g.: localhost)
	PowerdnsServerID string `json:"powerdns_server_id"`
	// PowerdnsBaseURL the base url of the server (e.g.: http://localhost:8081)
	PowerdnsBaseURL string `json:"powerdns_base_url"`
	// PowerdnsAPIKey the api key of powerdns (e.g.: changeme)
	PowerdnsAPIKey string `json:"powerdns_api_key"`
	// WebHookID is used to register this service in the inventory as event listener.
	// This should not change, otherwise the service is registered twice.
	// (e.g. 52acd668-3171-45a3-b23a-05adc76dc809)
	WebHookID string `json:"web_hook_id"`
	// InventoyRestRestURL the base url of the inventory server (e.g.: http://10.0.0.7:8080/api/v1)
	InventoyRestRestURL string `json:"inventory_rest_rest_url"`
	// InventoryAuthorizationHeader the authorization header to call the webhook registration server (e.g.: Basic bWFydGluOmdlaGVpbQ==)
	InventoryAuthorizationHeader string `json:"inventory_authorization_header"`
}

// environmentVariableMapperWithDefaultSupport mapper for environment variables
func environmentVariableMapperWithDefaultSupport(placeholderName string) string {
	variable := placeholderName
	dFault := ""
	match := variableMatcher.FindStringSubmatch(placeholderName)
	if match != nil {
		variable = match[1]
		dFault = match[2]
	}
	value := os.Getenv(variable)
	if len(value) == 0 {
		return dFault
	}
	return value
}

// loadConfig loads the configuration and maps the variables or their default values.
func loadConfig(configFile string) (*Config, error) {
	confContent, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	// expand environment variables
	confContent = []byte(os.Expand(string(confContent), environmentVariableMapperWithDefaultSupport))
	conf := &Config{}
	if err := json.Unmarshal(confContent, conf); err != nil {
		return nil, err
	}
	return conf, nil
}
