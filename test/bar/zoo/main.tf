resource "null_resource" "foo" {}

data "terraform_remote_state" "security_group" {
  backend = "s3"

  config = {
    bucket = "mybucket"
    key    = "path/to/my/key"
    region = "us-east-1"
  }
}
