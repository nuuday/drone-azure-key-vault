// Copyright 2019 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package main

import (
	"net/http"

	"github.com/drone/drone-go/plugin/secret"
	"github.com/nuuday/drone-azure-key-vault/plugin"

	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

// spec provides the plugin settings.
type Specification struct {
	Bind   string `envconfig:"DRONE_BIND"`
	Debug  bool   `envconfig:"DRONE_DEBUG"`
	Secret string `envconfig:"DRONE_SECRET"`

	AzureClientID     string `envconfig:"AZURE_CLIENT_ID" required:"true"`
	AzureClientSecret string `envconfig:"AZURE_CLIENT_SECRET" required:"true"`
	AzureTenantID     string `envconfig:"AZURE_TENANT_ID" required:"true"`
	AzureDebug        bool   `envconfig:"AZURE_DEBUG" default:"false"`
}

func main() {
	spec := new(Specification)
	err := envconfig.Process("", spec)
	if err != nil {
		log.Fatal(err)
	}

	if spec.Debug {
		log.SetLevel(log.DebugLevel)
	}
	if spec.Secret == "" {
		log.Fatalln("Missing secret key")
	}
	if spec.Bind == "" {
		spec.Bind = ":3000"
	}

	handler := secret.Handler(
		spec.Secret,
		plugin.New(
			spec.AzureDebug,
		),
		log.StandardLogger(),
	)

	log.Infof("server listening on address %s", spec.Bind)

	http.Handle("/", handler)
	log.Fatal(http.ListenAndServe(spec.Bind, nil))
}
