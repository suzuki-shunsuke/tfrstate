resource "null_resource" "foo" {}

output "foo" {
  value = "bar"
}

terraform {
  backend "s3" {
    bucket = "mybucket"
    key    = "path/to/my/key"
    region = "us-east-1"
  }
}
