/*
 * Author: Chris Lenz <chris@rtbrick.com>
 * Copyright (c) 2016 - 2019, RtBrick, Inc.
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"github.com/leitstand/leitstand-powerdns/pkg/version"

	powerdns "github.com/mittwald/go-powerdns"
)

type application struct {
	config         *Config
	powerdnsClient powerdns.Client
}

// @title Leitstand powerdns-connector API
// @version 1.0
// @description _Copyright (C) 2019, RtBrick, Inc._
// @description [Find additional information here](./description.html)
// @contact.name Chris Lenz (RtBrick)
// @contact.url http://www.rtbrick.com
// @contact.email chris@rtbrick.com
func main() {
	addr := flag.String("addr", ":19991", "HTTP network address")
	configFile := flag.String("config", "/etc/leitstand/connector/powerdns.json", "Configuration for the powerdns connector")
	versionFlag := flag.Bool("version", false, "Returns the software version")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version.VERSION)
		return
	}

	config, err := loadConfig(*configFile)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	fmt.Println(config)
	powerdnsClient, err := powerdns.New(
		powerdns.WithBaseURL(config.PowerdnsBaseURL),
		powerdns.WithAPIKeyAuthentication(config.PowerdnsAPIKey),
	)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	app := &application{
		config:         config,
		powerdnsClient: powerdnsClient,
	}
	handler := app.routes()
	go newRbmsRepository(config).registerWebHook(context.Background())
	serve(addr, handler)
}

func serve(addr *string, handler http.Handler) {
	srv := &http.Server{
		Addr:    *addr,
		Handler: handler,
	}
	fmt.Printf("Starting server on %s\n", *addr)
	err := srv.ListenAndServe()
	fmt.Println(err)
}
