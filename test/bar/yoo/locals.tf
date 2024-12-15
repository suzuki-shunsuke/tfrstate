locals {
  foo = data.terraform_remote_state.security_group.outputs.foo
  bar = data.terraform_remote_state.security_group.outputs.bar
}
