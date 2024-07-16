# sigstore-bundle-upgrade

This repo contains a CLI utility that can be used to upgrade a Sigstore Bundle from one version to a newer version.

## Usage

```shell
$ go build

$ ./sigstore-bundle-upgrade -h
Usage: sigstore-bundle-upgrade <path/to/sigstore/bundle>
  -in-place
        Update the bundle in place (otherwise print to stdout)
  -pretty
        Pretty print the output
  -version string
        Bundle version to upgrade to (default "0.3")

$ ./sigstore-bundle-upgrade myBundle.sigstore.json > myBundle.v0.3.sigstore.json
```
