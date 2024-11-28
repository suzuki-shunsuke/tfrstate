## Example

```sh
tfrstate find -plan-json foo/plan.json -backend-dir foo
```

Outputs:

```
INFO[0000] S3 buckend configuration                      bucket=mybucket env=darwin/arm64 key=path/to/my/key program=tfrstate version=
INFO[0000] Found *.tf files                              env=darwin/arm64 num_of_files=5 program=tfrstate version=
INFO[0000] terraform_remote_state is found               env=darwin/arm64 file=bar/yoo/main.tf program=tfrstate version=
INFO[0000] terraform_remote_state is found               env=darwin/arm64 file=bar/zoo/main.tf program=tfrstate version=
[
  {
    "dir": "bar/yoo",
    "files": [
      {
        "path": "locals.tf",
        "outputs": [
          "foo"
        ]
      }
    ]
  },
  {
    "dir": "bar/zoo",
    "files": [
      {
        "path": "foo.tf",
        "outputs": [
          "foo"
        ]
      }
    ]
  }
]
```
