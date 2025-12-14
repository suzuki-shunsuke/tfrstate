# tfrstate (Terraform Remote State)

[![DeepWiki](https://img.shields.io/badge/DeepWiki-suzuki--shunsuke%2Ftfrstate-blue.svg?logo=data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACwAAAAyCAYAAAAnWDnqAAAAAXNSR0IArs4c6QAAA05JREFUaEPtmUtyEzEQhtWTQyQLHNak2AB7ZnyXZMEjXMGeK/AIi+QuHrMnbChYY7MIh8g01fJoopFb0uhhEqqcbWTp06/uv1saEDv4O3n3dV60RfP947Mm9/SQc0ICFQgzfc4CYZoTPAswgSJCCUJUnAAoRHOAUOcATwbmVLWdGoH//PB8mnKqScAhsD0kYP3j/Yt5LPQe2KvcXmGvRHcDnpxfL2zOYJ1mFwrryWTz0advv1Ut4CJgf5uhDuDj5eUcAUoahrdY/56ebRWeraTjMt/00Sh3UDtjgHtQNHwcRGOC98BJEAEymycmYcWwOprTgcB6VZ5JK5TAJ+fXGLBm3FDAmn6oPPjR4rKCAoJCal2eAiQp2x0vxTPB3ALO2CRkwmDy5WohzBDwSEFKRwPbknEggCPB/imwrycgxX2NzoMCHhPkDwqYMr9tRcP5qNrMZHkVnOjRMWwLCcr8ohBVb1OMjxLwGCvjTikrsBOiA6fNyCrm8V1rP93iVPpwaE+gO0SsWmPiXB+jikdf6SizrT5qKasx5j8ABbHpFTx+vFXp9EnYQmLx02h1QTTrl6eDqxLnGjporxl3NL3agEvXdT0WmEost648sQOYAeJS9Q7bfUVoMGnjo4AZdUMQku50McDcMWcBPvr0SzbTAFDfvJqwLzgxwATnCgnp4wDl6Aa+Ax283gghmj+vj7feE2KBBRMW3FzOpLOADl0Isb5587h/U4gGvkt5v60Z1VLG8BhYjbzRwyQZemwAd6cCR5/XFWLYZRIMpX39AR0tjaGGiGzLVyhse5C9RKC6ai42ppWPKiBagOvaYk8lO7DajerabOZP46Lby5wKjw1HCRx7p9sVMOWGzb/vA1hwiWc6jm3MvQDTogQkiqIhJV0nBQBTU+3okKCFDy9WwferkHjtxib7t3xIUQtHxnIwtx4mpg26/HfwVNVDb4oI9RHmx5WGelRVlrtiw43zboCLaxv46AZeB3IlTkwouebTr1y2NjSpHz68WNFjHvupy3q8TFn3Hos2IAk4Ju5dCo8B3wP7VPr/FGaKiG+T+v+TQqIrOqMTL1VdWV1DdmcbO8KXBz6esmYWYKPwDL5b5FA1a0hwapHiom0r/cKaoqr+27/XcrS5UwSMbQAAAABJRU5ErkJggg==)](https://deepwiki.com/suzuki-shunsuke/tfrstate)

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
