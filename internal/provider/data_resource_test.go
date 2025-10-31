// Copyright (c) The Camo Terraform Provider Authors
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

var testUUID uuid.UUID

func TestAccDataResource(t *testing.T) {
	startingUUID := uuid.UUID([16]byte{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15})
	recreatedUUID := uuid.UUID([16]byte{9, 8, 7, 6, 5, 4, 3, 2, 1})
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: `
					resource "camo_data" "test" {
						input           = "bar, foo"
						input_sensitive = "sensitive_bar, sensitive_foo"
						triggers_replace = [
							"hello",
							"world"
						]
					}
					`,
				PreConfig: func() {
					testUUID = startingUUID
				},
				ConfigPlanChecks: planChecks(
					plancheck.ResourceActionCreate,
					"bar, foo",
					"sensitive_bar, sensitive_foo",
					[]string{"hello", "world"},
					"", /* knownID, unknown on initial creation plan */
				),
				ConfigStateChecks: stateChecks(
					"bar, foo",
					"sensitive_bar, sensitive_foo",
					[]string{"hello", "world"},
					startingUUID.String(),
				),
			},
			// No changes causes no updates.
			{
				Config: `
					resource "camo_data" "test" {
						input           = "bar, foo"
						input_sensitive = "sensitive_bar, sensitive_foo"
						triggers_replace = [
							"hello",
							"world"
						]
					}
					`,
				PreConfig: func() {
					// This shouldn't be used since the resoure is not changing.
					testUUID = [16]byte{1, 2, 3}
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: append(
						planChecks(
							plancheck.ResourceActionNoop,
							"bar, foo",
							"sensitive_bar, sensitive_foo",
							[]string{"hello", "world"},
							startingUUID.String(),
						).PreApply,
						plancheck.ExpectEmptyPlan(),
					),
				},
			},
			// In-place update
			{
				// Change input and sensitive_input. Should NOT trigger recreation.
				Config: `
					resource "camo_data" "test" {
						input           = "bar"
						input_sensitive = "sensitive_bar"
						triggers_replace = [
							"hello",
							"world"
						]
					}
					`,
				PreConfig: func() {
					// This should have no effect. Since this is in-place, the
					// ID should NOT be updated on the resource.
					testUUID = [16]byte{1, 2, 3, 4}

				},
				// Should update in-place.
				ConfigPlanChecks: planChecks(
					plancheck.ResourceActionUpdate,
					"bar",
					"sensitive_bar",
					[]string{"hello", "world"},
					startingUUID.String(),
				),

				ConfigStateChecks: stateChecks(
					"bar",
					"sensitive_bar",
					[]string{"hello", "world"},
					startingUUID.String(),
				),
			},
			// Recreate
			{
				// Changing triggers_replace causes recreation.
				Config: `
						resource "camo_data" "test" {
							input           = "bar"
							input_sensitive = "sensitive_bar"
							triggers_replace = [
								"hello",
								"world",
								"add a value to trigger recreate",
							]
						}
						`,
				PreConfig: func() {
					// Recreating the resource should also create a new ID.
					testUUID = recreatedUUID

				},
				// Should update in-place.
				ConfigPlanChecks: planChecks(
					plancheck.ResourceActionDestroyBeforeCreate,
					"bar",
					"sensitive_bar",
					[]string{"hello", "world", "add a value to trigger recreate"},
					"", /* knownID, unknown when resource is recreated */
				),

				ConfigStateChecks: stateChecks(
					"bar",
					"sensitive_bar",
					[]string{"hello", "world", "add a value to trigger recreate"},
					recreatedUUID.String(),
				),
			},
		},
	})
}

// planChecks returns a Plan Check configuration based on the given expected values.
//
// knownID should be supplied if known. If the ID is expected to be unknown (for
// example during the initial create) provide an empty string.
func planChecks(
	action plancheck.ResourceActionType,
	input string,
	sensitiveInput string,
	triggers []string,
	knownID string,
) resource.ConfigPlanChecks {
	triggerChecks := make([]knownvalue.Check, len(triggers))
	for i, t := range triggers {
		triggerChecks[i] = knownvalue.StringExact(t)
	}

	// If ID is known, check that it matches what we expect.
	var idCheck plancheck.PlanCheck
	if knownID != "" {
		idCheck = plancheck.ExpectKnownValue("camo_data.test", tfjsonpath.New("id"), knownvalue.StringExact(knownID))

	} else {
		idCheck = plancheck.ExpectUnknownValue("camo_data.test", tfjsonpath.New("id"))

	}

	return resource.ConfigPlanChecks{
		PreApply: []plancheck.PlanCheck{
			// Should create a new resource.
			plancheck.ExpectResourceAction("camo_data.test", action),

			// Verify input and output
			plancheck.ExpectKnownValue(
				"camo_data.test",
				tfjsonpath.New("input"),
				knownvalue.StringExact(input),
			),
			plancheck.ExpectKnownValue(
				"camo_data.test",
				tfjsonpath.New("output"),
				knownvalue.StringExact(input),
			),

			// Verify sensitive input and output
			plancheck.ExpectKnownValue(
				"camo_data.test",
				tfjsonpath.New("input_sensitive"),
				knownvalue.StringExact(sensitiveInput),
			),
			plancheck.ExpectKnownValue(
				"camo_data.test",
				tfjsonpath.New("output_sensitive"),
				knownvalue.StringExact(sensitiveInput),
			),
			plancheck.ExpectSensitiveValue("camo_data.test",
				tfjsonpath.New("input_sensitive")),
			plancheck.ExpectSensitiveValue("camo_data.test",
				tfjsonpath.New("output_sensitive")),

			// Verify triggers
			plancheck.ExpectKnownValue(
				"camo_data.test",
				tfjsonpath.New("triggers_replace"),
				knownvalue.ListExact(triggerChecks),
			),

			idCheck,
		},
	}
}

// stateChecks returns a State Check configuration based on the given expected values.
func stateChecks(
	input string,
	inputSensitive string,
	triggers []string,
	id string,
) []statecheck.StateCheck {
	triggerChecks := make([]knownvalue.Check, len(triggers))
	for i, t := range triggers {
		triggerChecks[i] = knownvalue.StringExact(t)
	}

	return []statecheck.StateCheck{
		// Input should match output
		statecheck.CompareValuePairs(
			"camo_data.test",
			tfjsonpath.New("input"),
			"camo_data.test",
			tfjsonpath.New("output"),
			compare.ValuesSame(),
		),

		statecheck.ExpectKnownValue(
			"camo_data.test",
			tfjsonpath.New("input"),
			knownvalue.StringExact(input),
		),

		// Sensitive input should match sensitive output
		statecheck.CompareValuePairs(
			"camo_data.test",
			tfjsonpath.New("input_sensitive"),
			"camo_data.test",
			tfjsonpath.New("output_sensitive"),
			compare.ValuesSame(),
		),

		statecheck.ExpectKnownValue(
			"camo_data.test",
			tfjsonpath.New("input_sensitive"),
			knownvalue.StringExact(inputSensitive),
		),

		statecheck.ExpectSensitiveValue("camo_data.test",
			tfjsonpath.New("input_sensitive"),
		),
		statecheck.ExpectSensitiveValue("camo_data.test",
			tfjsonpath.New("output_sensitive"),
		),

		statecheck.ExpectKnownValue(
			"camo_data.test",
			tfjsonpath.New("triggers_replace"),
			knownvalue.ListExact(triggerChecks),
		),

		statecheck.ExpectKnownValue(
			"camo_data.test",
			tfjsonpath.New("id"),
			knownvalue.StringExact(id),
		),
	}
}
