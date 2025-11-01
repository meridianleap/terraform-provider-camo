---
page_title: "camo_data Resource - camo"
subcategory: ""
description: |-
  Mimics the functionality of the terraform_data builtin resource, but with added attributes for storing sensitive inputs.
---

# camo_data (Resource)

Note: This resource closely mimics `terraform_data`[^1]. In many cases, the
`terraform_data` built-in resource will be sufficient. However, the following
special case is not covered by `terraform_data`, and this resource is necessary.

`terraform_data` is super useful for running side effects as part of the
lifecycle of other resources. Usually this involves using a `provisioner` block
inside the `terraform_data` resource. Because destroy-time scripts inside the
`provisioner` block can only access attributes on the `terraform_data` using the
`self.*` reference (no access to `var.*` or `local.*`), this data must be added
to the `terraform_data` resource as attributes. There are two fields you can use
for this:
- `input`: accepts any type and **never** triggers a
destroy-and-create when changed. Computes an `output` attribute which is
**never** sensitive.
- `triggers_replace`: accepts any type and **always** triggers a destroy-and-create when changed. No output is generated from this attribute.

**There is one case where neither of these two attributes is sufficient:** your
`provisioner` needs access to a sensitive value, _and_ you only want the
`provisioner` block to run as part of the `destroy` lifecycle of another
resource. In this scenario, you can't store the sensitive value in the `input`
attribute because it will cause the sensitive value to be written as plan-text
to the `plan` output (via the computed `output` attribute). You also can't store
it in `triggers_replace` because it will cause the `provisioner` block to run
whenever the sensitive value is changed, not just when the other resource is
destroyed. As a concrete example:

```hcl
# Suppose you have a webserver. When the webserver is destroyed,
# you need to run some additional cleanup scripts, for example to
# remove it from a mesh VPN.
resource "some_cloud_server" "webserver" {
  // various inputs...
}

resource "terraform_data" "remove_from_mesh_vpn" {
  # Even though the API token is sensitive, it's displayed in plaintext in the
  # output for this component. Not good!
  input = var.sensitive_mesh_api_token

  # This value must be stable for the lifetime of the server resource, ie it
  # will only change when the server is destroyed. Therefore, we cannot store
  # the API token here, since the token might change when we don't want to
  # delete the node from the mesh (eg rotating the API keys).
  triggers_replace = some_cloud_server.webserver.id

  # A hypothetical script that deletes the webserver's node from the mesh
  # VPN when the server instance is destroyed.
  provisioner "local-exec" {
    when = destroy
    environment = {
      MESH_VPN_API_KEY = self.input
      DEVICE_ID        = self.triggers_replace
    }
    working_dir = "${path.module}/scripts"
    command     = "./delete-mesh-nodes.sh"
  }
}
```

`camo_data` addresses this problem by mimicking the behavior of `terraform_data`, with two additional attributes:
- `input_sensitive`: identical to `input`, but produces `output_sensitive` instead of `output`.
- `output_sensitive`: outputs the value of `input_sensitive`, but as a "sensitive" value, not visible in the `plan` output.

The `remove_from_mesh_vpn` block above thus becomes:
```hcl
resource "camo_data" "remove_from_mesh_vpn" {
  # This input produces output_sensitive, which is hidden in the plan output.
  input_sensitive = var.sensitive_mesh_api_token.

  triggers_replace = some_cloud_server.webserver.id

  provisioner "local-exec" {
    when = destroy
    environment = {
      MESH_VPN_API_KEY = self.input_sensitive
      DEVICE_ID        = self.triggers_replace
    }
    working_dir = "${path.module}/scripts"
    command     = "./delete-mesh-nodes.sh"
  }
}
```

## Replacing `terraform_data` with `camo_data`
Simply replacing your `terraform_data` with `camo_data` will cause the old
`terraform_data` to be destroyed. If it has a `provisioner` destroy block, the
block will be triggered. This is probably not what you want to happen.

Instead, you can add `removed` blocks to remove the `terraform_data` from state
without triggering a destroy operation. The following examples explain this
process. **Back up your state before performing these operations. Carefully
review all plan output to ensure you aren't causing undesired side effects!!**

**Initial state:** a `terraform_data` and a resource that depends on the data:
```hcl
resource "terraform_data" "test" {
  input = "foo"
  triggers_replace = var.bar
  provisioner "local-exec" {
    when = destroy
    # Don't want this to execute when we replace this with camo_data!
    command = "./something-destructive.sh"
  }
}

resource "my_resource" "dependent" {
  value = terraform_data.test.output
}
```

**Step 1:** replace the `terraform_data` block and add the `removed` block:
```hcl
# This prevents the terraform_data from being "destroyed"
removed {
  from = terraform_data.test

  lifecycle {
    # This is the key piece
    destroy = false
  }
}

# Replace `terraform_data` with `camo_data`
resource "camo_data" "test" {
  input = "foo"
  triggers_replace = var.bar
  provisioner "local-exec" {
    when = destroy
    command = "./something-destructive.sh"
  }
}

resource "my_resource" "dependent" {
  # Replace all references to `terraform_data` with `camo_data`
  value = camo_data.test.output
}
```

**Step 2:** run `terraform plan` and confirm that nothing will be destroyed. It
should look something like this:
```hcl
Terraform will perform the following actions:

  # camo_data.test will be created
  + resource "camo_data" "test" {
      + id       = (known after apply)
      + input    = "foo"
      + input_wo = (write-only attribute)
      + output   = "foo"
    }

 # terraform_data.test will no longer be managed by Terraform, but will not be destroyed
 # (destroy = false is set in the configuration)
 . resource "terraform_data" "test" {
        id     = "38241ed9-94fc-bc4f-5027-ee7032868db4"
        # (2 unchanged attributes hidden)
    }

Plan: 1 to add, 0 to change, 0 to destroy.
╷
│ Warning: Some objects will no longer be managed by Terraform
│ 
│ If you apply this plan, Terraform will discard its tracking information for the following objects, but it will not delete them:
│  - terraform_data.test
│ 
│ After applying this plan, Terraform will no longer manage these objects. You will need to import them into Terraform to manage them again.
╵

Do you want to perform these actions?
```

**Step 3:** Apply the plan. After the apply, you can remove the `removed` block.
Running `apply` again should be a no-op.

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `input` (Dynamic) A value which will be stored in the instance state, and reflected in the `output` attribute after apply.
- `input_sensitive` (Dynamic, Sensitive) A value which will be stored in the instance state, and reflected as a sensitive value in the `output_sensitive` attribute after apply.
- `triggers_replace` (Dynamic) A value that is stored in the instance state and will force replacement when the value changes.

### Read-Only

- `id` (String) A string value unique to the resource instance.
- `output` (Dynamic) The computed value derived from the `input` argument. During a plan where `output` is unknown, it will still be of the same type as `input`.
- `output_sensitive` (Dynamic, Sensitive) The sensitive computed value derived from the `input_sensitive` argument. During a plan where `output_sensitive` is unknown, it will still be of the same type as `input_sensitive`.

## Caveats
The sensitive data will still be visible in the `terraform.state`.  This could
possibly be addressed by using a "write-only" attribute, eg `input_wo`.
Unfortunately, as of this release (`terraform v1.13.4`), write-only attributes are
not accessible from provisioner blocks using the `self` attribute.

[^1]: https://developer.hashicorp.com/terraform/language/resources/terraform-data