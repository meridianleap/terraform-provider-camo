resource "camo_data" "test" {
  # As in terraform_data, this input will be exported as a non-sensitive
  # `output` attribute.
  input = "foo"

  # This input will be exported as a sensitive `sensitive_input` attribute.
  # Changing it will NOT trigger a replacement of this resource, and therefore
  # will NOT cause the provisioner block to execute.
  input_sensitive = "sensitive_bar"

  # As in terraform_data, modifying these triggers causes replacement.
  triggers_replace = [
    "change me to trigger replacement",
  ]

  # This is probably why this resource is useful to you. If you want to be able
  # to store a sensitive value in the resource state, AND be able to change that
  # value without triggering this provisioner function, storing that value in
  # `triggers_replace` is not appropriate. Instead, store it in input_sensitive.
  provisioner "local-exec" {
    when    = destroy
    command = "echo \"I only run when the triggers are changed.\""
  }
}
