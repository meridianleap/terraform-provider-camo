// Copyright (c) The Camo Terraform Provider Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"log"
	"terraform-provider-camo/internal/provider"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	// Set by goreleaser for the compiled binary.
	version string = "dev"
)

func main() {
	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/meridianleap/camo",
	}

	err := providerserver.Serve(context.Background(), provider.New(version, uuid.New), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
