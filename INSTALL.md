# Install

tfrstate is written in Go. So you only have to install a binary in your `PATH`.

There are some ways to install tfrstate.

1. [Homebrew](#homebrew)
1. [aqua](#aqua)
1. [GitHub Releases](#github-releases)
1. [Build an executable binary from source code yourself using Go](#build)

## Homebrew

You can install tfrstate using [Homebrew](https://brew.sh/).

```console
$ brew install suzuki-shunsuke/tfrstate/tfrstate
```

## aqua

[aqua-registry >= v4.264.0](https://github.com/aquaproj/aqua-registry/releases/tag/v4.264.0)

You can install tfrstate using [aqua](https://aquaproj.github.io/).

```console
$ aqua g -i suzuki-shunsuke/tfrstate
```

## Build an executable binary from source code yourself using Go

```sh
go install github.com/suzuki-shunsuke/tfrstate/cmd/tfrstate@latest
```

## GitHub Releases

You can download an asset from [GitHub Reelases](https://github.com/suzuki-shunsuke/tfrstate/releases).
Please unarchive it and install a pre built binary into `$PATH`. 

### Verify downloaded assets from GitHub Releases

You can verify downloaded assets using some tools.

1. [GitHub CLI](https://cli.github.com/)
1. [slsa-verifier](https://github.com/slsa-framework/slsa-verifier)
1. [Cosign](https://github.com/sigstore/cosign)

### 1. GitHub CLI

You can install GitHub CLI by aqua.

```sh
aqua g -i cli/cli
```

```sh
version=v0.1.0
gh release download -R suzuki-shunsuke/tfrstate "$version" -p tfrstate_darwin_arm64.tar.gz
gh attestation verify tfrstate_darwin_arm64.tar.gz \
  -R suzuki-shunsuke/tfrstate \
  --signer-workflow suzuki-shunsuke/go-release-workflow/.github/workflows/release.yaml
```

### 2. slsa-verifier

You can install slsa-verifier by aqua.

```sh
aqua g -i slsa-framework/slsa-verifier
```

```sh
version=v0.1.0
gh release download -R suzuki-shunsuke/tfrstate "$version"
slsa-verifier verify-artifact tfrstate_darwin_arm64.tar.gz \
  --provenance-path multiple.intoto.jsonl \
  --source-uri github.com/suzuki-shunsuke/tfrstate \
  --source-tag "$version"
```

### 3. Cosign

You can install Cosign by aqua.

```sh
aqua g -i sigstore/cosign
```

```sh
version=v0.1.0
checksum_file="tfrstate_${version#v}_checksums.txt"
gh release download -R suzuki-shunsuke/tfrstate "$version"
cosign verify-blob \
  --signature "${checksum_file}.sig" \
  --certificate "${checksum_file}.pem" \
  --certificate-identity-regexp 'https://github\.com/suzuki-shunsuke/go-release-workflow/\.github/workflows/release\.yaml@.*' \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  "$checksum_file"
cat "$checksum_file" | sha256sum -c --ignore-missing
```