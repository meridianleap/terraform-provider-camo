// Copyright (c) The Camo Terraform Provider Authors
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource               = &dataResource{}
	_ resource.ResourceWithConfigure  = &dataResource{}
	_ resource.ResourceWithModifyPlan = &dataResource{}
)

// NewDataResource is a helper function to simplify the provider implementation.
func NewDataResource() resource.Resource {
	return &dataResource{}
}

// dataResource is the resource implementation.
type dataResource struct {
	uuidFunc func() uuid.UUID
}

type dataResourceModel struct {
	ID              types.String  `tfsdk:"id"`
	Input           types.Dynamic `tfsdk:"input"`
	Output          types.Dynamic `tfsdk:"output"`
	InputSensitive  types.Dynamic `tfsdk:"input_sensitive"`
	OutputSensitive types.Dynamic `tfsdk:"output_sensitive"`
	TriggersReplace types.Dynamic `tfsdk:"triggers_replace"`
}

// Metadata returns the resource type name.
func (r *dataResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data"
}

// Schema defines the schema for the resource.
func (r *dataResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Persists arbitrary input values to state for later retrieval. Mimics the functionality of the terraform_data builtin resource, but with added functionality for storing sensitive inputs.",
		MarkdownDescription: "Persists arbitrary input values to state for later retrieval. Mimics the functionality of the `terraform_data` builtin resource, but with added functionality for storing sensitive inputs.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "A string value unique to the resource instance.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"input": schema.DynamicAttribute{
				Description:         "A value which will be stored in the instance state, and reflected in the output attribute after apply.",
				MarkdownDescription: "A value which will be stored in the instance state, and reflected in the `output` attribute after apply.",
				Optional:            true,
			},
			"output": schema.DynamicAttribute{
				Description:         "The computed value derived from the input argument. During a plan where output is unknown, it will still be of the same type as input.",
				MarkdownDescription: "The computed value derived from the `input` argument. During a plan where `output` is unknown, it will still be of the same type as `input`.",
				Computed:            true,
			},
			"input_sensitive": schema.DynamicAttribute{
				Description:         "A value which will be stored in the instance state, and reflected as a sensitive value in the output_sensitive attribute after apply.",
				MarkdownDescription: "A value which will be stored in the instance state, and reflected as a sensitive value in the `output_sensitive` attribute after apply.",
				Optional:            true,
				Sensitive:           true,
			},
			"output_sensitive": schema.DynamicAttribute{
				Description:         "The sensitive computed value derived from the input_sensitive argument. During a plan where output_sensitive is unknown, it will still be of the same type as input_sensitive.",
				MarkdownDescription: "The sensitive computed value derived from the `input_sensitive` argument. During a plan where `output_sensitive` is unknown, it will still be of the same type as `input_sensitive`.",
				Computed:            true,
				Sensitive:           true,
			},
			"triggers_replace": schema.DynamicAttribute{
				Description: "A value that is stored in the instance state and will force replacement when the value changes.",
				Optional:    true,
			},
		},
	}
}

// Configure configures the resource with the ResourceData supplied from the provider.
func (r *dataResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	rd, ok := req.ProviderData.(*resourceData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected resourceData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.uuidFunc = rd.uuidFunc
}

// Create creates the resource and sets the initial state.
func (r *dataResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read the provided plan variables into our struct, check for issues.
	var plan dataResourceModel
	var diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the outputs to match the inputs.
	plan.Output = plan.Input
	plan.OutputSensitive = plan.InputSensitive

	plan.ID = types.StringValue(r.uuidFunc().String())

	// Set the state.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the state with the latest data.
func (r *dataResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// All state is local, so this is a no-op.
}

// Update updates the resource and sets the updated state on success.
func (r *dataResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan dataResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the outputs to match the inputs.
	plan.Output = plan.Input
	plan.OutputSensitive = plan.InputSensitive

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the state on success.
func (r *dataResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No side effect, just remove from state.
}

// ModifyPlan updates the plan to match the expected result after apply. In the
// case of this resource, that means setting the outputs equal to the inputs,
// and checking to see if the triggers_replace values are being changed, which
// would cause the resource to be destroyed and recreated.
func (r *dataResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Nothing to do on destroy
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan dataResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	triggersPath := path.Root("triggers_replace")

	// Retrieve triggers from state to see if they changed. If so, object must
	// be replaced.
	var currTriggers types.Dynamic
	diags = req.State.GetAttribute(ctx, triggersPath, &currTriggers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !currTriggers.Equal(plan.TriggersReplace) {
		resp.RequiresReplace.Append(triggersPath)
		plan.ID = types.StringUnknown()
	}

	plan.Output = plan.Input
	plan.OutputSensitive = plan.InputSensitive
	resp.Plan.Set(ctx, plan)
}
