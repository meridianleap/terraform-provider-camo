// Copyright (c) The Camo Terraform Provider Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type resourceData struct {
	uuidFunc func() uuid.UUID
}

// Ensure camoProvider satisfies necessary interfaces.
var _ provider.Provider = &camoProvider{}

type camoProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
	// Function to generate UUIDs. Allows dependency injection during testing.
	uuidFunc func() uuid.UUID
}

func (p *camoProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "camo"
	resp.Version = p.version
}

func (p *camoProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "This provider addresses a shortcoming of terraform_data where sensitive values are output as plain text in the plan.",
		MarkdownDescription: "This provider addresses a shortcoming of `terraform_data` where sensitive values are output as plain text in the `plan.`",
	}
}

func (p *camoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	resp.ResourceData = &resourceData{uuidFunc: p.uuidFunc}
}

func (p *camoProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDataResource,
	}
}

func (p *camoProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}

func New(version string, uuidFunc func() uuid.UUID) func() provider.Provider {
	return func() provider.Provider {
		return &camoProvider{
			version:  version,
			uuidFunc: uuidFunc,
		}
	}
}
