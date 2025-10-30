# Camo Terraform Provider

This provider supplies a replacement resource for the builtin `terraform_data`
resource, with the added functionality of a `sensitive_input` attribute that
will produce `sensitive_output` (marked as "sensitive"), thus preventing leaking
sensitive input to `plan` output.

**Note:** sensitive variables will still be visible in `state`! Store your `state`
safely to avoid leaks.

## Using the provider
See generated docs.

## Requirements

- [OpenTofu](https://opentofu.org/docs/intro/install/) >= 1.0 or [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24

## Building The Provider

1. Clone the repository.
2. `cd` into the repository directory.
3. Build the provider using `go install .` command:

```shell
go install .
```
