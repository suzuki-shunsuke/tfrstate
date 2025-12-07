# tfrstate (Terraform Remote State)

[MIT](LICENSE) | [Install](INSTALL.md) | [Usage](USAGE.md)

CLI to find directories where changed [terraform_remote_state data source](https://developer.hashicorp.com/terraform/language/state/remote-state-data) is used.

You can warn changes of Terraform Output Values in CI:

![image](https://github.com/user-attachments/assets/128ee2ee-1e69-4303-b835-2e963b519c58)

You can also create pull requests via CI to reflect changes of Output Values to Terraform Root Modules:

![image](https://github.com/user-attachments/assets/2e1900a2-49af-4574-b342-95cf0ba38225)

## Overview

When [Terraform Output Values](https://developer.hashicorp.com/terraform/language/values/outputs) are updated, you would want to reflect the update to Terraform Root Modules where referring those values via `terraform_remote_state` data sources.
Or when you remove Terraform Output Values, you would want to know which Terraform Root Modules are depending on those values.

tfrstate is a CLI to find Terraform Root Modules depending on a given Terraform State via `terraform_remote_state`.
Using this tool, you can look into the dependency, notifying to people when changing output values in CI, and creating pull requests to reflect changes of output values after applying changes.

## Supported Backends

- [S3 Backend](https://developer.hashicorp.com/terraform/language/backend/s3)
- [GCS Backend](https://developer.hashicorp.com/terraform/language/backend/gcs)

## How To Use

[Please see PoC too](https://github.com/suzuki-shunsuke/poc-tfrstate).

1. Check directories where a specific output is used.

```sh
tfrstate find -s3-bucket mybucket -s3-key path/to/my/key -output foo
```

2. Post a comment when outputs are changed in CI

```sh
terraform plan -out plan.out
terraform show -json plan.out > plan.json
tfrstate find -plan-json plan.json -base-dir "$(git rev-parse --show-toplevel)" > result.json
length=$(jq length result.json)
if [ "$length" -eq 0 ]; then
  exit 0
fi
# Post a comment
```

To post comments, [github-comment](https://github.com/suzuki-shunsuke/github-comment) is useful.

3. Create pull requests after running `terraform apply`

```sh
terraform apply -auto-approve plan.out
while read -r dir; do
  echo "[INFO] Creating a pull request: $dir" >&2
  # Create a pull request
done < <(jq -r ".[].dir" result.json)
```

## Output Format

```json
[
  {
    "dir": "A directory where depending on changed outputs. A relative path from the base directory",
    "files": [
      {
        "path": "A file depending on changed outputs. A relative path from dir",
        "outputs": [
          "changed output name"
        ]
      }
    ]
  }
]
```

## LICENSE

[MIT](LICENSE)
