/*
 * Author: Chris Lenz <chris@rtbrick.com>
 * Copyright (c) 2016 - 2019, RtBrick, Inc.
 */

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/matryer/is"
	powerdns "github.com/mittwald/go-powerdns"
)

func Test_application_rbmsEvent(t *testing.T) {
	is := is.New(t)
	type powerDNSRequest struct {
		url     string
		content string
		method  string
	}
	config := &Config{
		Nameservers:      []string{"dns."},
		PowerdnsServerID: "localhost",
		PowerdnsBaseURL:  "http://localhost:9999",
		PowerdnsAPIKey:   "changeme",
	}
	powerDNSClient, err := powerdns.New(
		powerdns.WithBaseURL(config.PowerdnsBaseURL),
		powerdns.WithAPIKeyAuthentication(config.PowerdnsAPIKey),
	)
	is.NoErr(err)
	app := &application{
		config:         config,
		powerdnsClient: powerDNSClient,
	}
	tests := []struct {
		body                     string
		event                    string
		expectedResponseCode     int
		powerdnsResponse         int
		expectedPowerDNSRequests []powerDNSRequest
	}{
		{event: ElementDnsRecordSetModifiedEvent, body: "", expectedResponseCode: http.StatusBadRequest},
		{event: ElementDnsRecordSetModifiedEvent, body: readFileToString(t, "./testdata/rename.json"), expectedResponseCode: http.StatusNoContent,
			powerdnsResponse: http.StatusOK,
			expectedPowerDNSRequests: []powerDNSRequest{
				{method: "PATCH", url: "/api/v1/servers/localhost/zones/leitstand.io.", content: `{"name":"","type":"Zone","rrsets":[{"name":"test.leitstand.io.","type":"A","ttl":0,"changetype":"DELETE","records":null,"comments":null}]}`},
				{method: "PATCH", url: "/api/v1/servers/localhost/zones/leitstand.io.", content: `{"name":"","type":"Zone","rrsets":[{"name":"foo.leitstand.io.","type":"A","ttl":3600,"changetype":"REPLACE","records":[{"content":"10.0.0.10","disabled":false},{"content":"10.0.0.11","disabled":true}],"comments":null}]}`},
			},
		},
		{event: ElementDnsRecordSetModifiedEvent, body: readFileToString(t, "./testdata/add.json"), expectedResponseCode: http.StatusNoContent,
			powerdnsResponse: http.StatusOK,
			expectedPowerDNSRequests: []powerDNSRequest{
				{method: "PATCH", url: "/api/v1/servers/localhost/zones/leitstand.io.", content: `{"name":"","type":"Zone","rrsets":[{"name":"foo.leitstand.io.","type":"A","ttl":3600,"changetype":"REPLACE","records":[{"content":"10.0.0.10","disabled":false},{"content":"10.0.0.11","disabled":true}],"comments":null}]}`},
			},
		},
		{event: ElementDnsRecordSetModifiedEvent, body: readFileToString(t, "./testdata/delete.json"), expectedResponseCode: http.StatusNoContent,
			powerdnsResponse: http.StatusOK,
			expectedPowerDNSRequests: []powerDNSRequest{
				{method: "PATCH", url: "/api/v1/servers/localhost/zones/leitstand.io.", content: `{"name":"","type":"Zone","rrsets":[{"name":"test.leitstand.io.","type":"A","ttl":0,"changetype":"DELETE","records":null,"comments":null}]}`},
			},
		},
		{event: ElementDnsRecordSetModifiedEvent, body: readFileToString(t, "./testdata/add.json"), expectedResponseCode: http.StatusInternalServerError,
			powerdnsResponse: http.StatusInternalServerError,
			expectedPowerDNSRequests: []powerDNSRequest{
				{method: "PATCH", url: "/api/v1/servers/localhost/zones/leitstand.io.", content: `{"name":"","type":"Zone","rrsets":[{"name":"foo.leitstand.io.","type":"A","ttl":3600,"changetype":"REPLACE","records":[{"content":"10.0.0.10","disabled":false},{"content":"10.0.0.11","disabled":true}],"comments":null}]}`},
			},
		},
		{event: ElementDnsRecordSetModifiedEvent, body: readFileToString(t, "./testdata/delete.json"), expectedResponseCode: http.StatusInternalServerError,
			powerdnsResponse: http.StatusInternalServerError,
			expectedPowerDNSRequests: []powerDNSRequest{
				{method: "PATCH", url: "/api/v1/servers/localhost/zones/leitstand.io.", content: `{"name":"","type":"Zone","rrsets":[{"name":"test.leitstand.io.","type":"A","ttl":0,"changetype":"DELETE","records":null,"comments":null}]}`},
			},
		},
		{event: DnsZoneCreatedEvent, body: "", expectedResponseCode: http.StatusBadRequest},
		{event: DnsZoneCreatedEvent, body: readFileToString(t, "./testdata/zonecreated.json"), expectedResponseCode: http.StatusNoContent,
			powerdnsResponse: http.StatusOK,
			expectedPowerDNSRequests: []powerDNSRequest{
				{method: "POST", url: "/api/v1/servers/localhost/zones", content: `{"name":"leitstand.io.","type":"Zone","kind":"Native","nameservers":["dns."]}`},
			},
		},
		{event: DnsZoneRemovedEvent, body: "", expectedResponseCode: http.StatusBadRequest},
		{event: DnsZoneRemovedEvent, body: readFileToString(t, "./testdata/zoneremoved.json"), expectedResponseCode: http.StatusNoContent,
			powerdnsResponse: http.StatusOK,
			expectedPowerDNSRequests: []powerDNSRequest{
				{method: "DELETE", url: "/api/v1/servers/localhost/zones/leitstand.io."},
			},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test %d:", i), func(t *testing.T) {
			var powerDNSRequests []powerDNSRequest
			serverForTest(t, func(t *testing.T) {
				req, err := http.NewRequest("POST", fmt.Sprintf("/api/v1/events/%s", tt.event), bytes.NewReader([]byte(tt.body)))
				rr := httptest.NewRecorder()
				app.routes().ServeHTTP(rr, req)
				is.NoErr(err)
				checkResponseCode(t, tt.expectedResponseCode, rr.Code)
				if !reflect.DeepEqual(tt.expectedPowerDNSRequests, powerDNSRequests) {
					t.Errorf("Expected Powerdns calls not match the actual calls\nExpected:\n%v\nGot:\n%v\n", tt.expectedPowerDNSRequests, powerDNSRequests)
				}

			}, func(writer http.ResponseWriter, request *http.Request) {
				defer request.Body.Close()
				bodyBytes, err := ioutil.ReadAll(request.Body)
				if err != nil {
					log.Fatal(err)
				}
				powerDNSRequests = append(powerDNSRequests, powerDNSRequest{
					method:  request.Method,
					url:     request.RequestURI,
					content: strings.TrimSpace(string(bodyBytes)),
				})

				writer.WriteHeader(tt.powerdnsResponse)
				_, _ = writer.Write([]byte(`{}`))
			})

		})
	}
}

func checkResponseCode(t *testing.T, expected, actual int) {
	t.Helper()
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}
func readFileToString(t *testing.T, filename string) string {
	t.Helper()
	is := is.New(t)

	content, err := ioutil.ReadFile(filename)
	is.NoErr(err)

	return string(content)
}
func serverForTest(t *testing.T, f func(*testing.T), handler http.HandlerFunc) {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(handler))
	l, _ := net.Listen("tcp", "localhost:9999")
	ts.Listener = l
	ts.Start()
	defer ts.Close()
	f(t)
}
