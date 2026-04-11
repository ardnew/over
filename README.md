# over

##### A no-frills, no-fuss, zero-dependency semantic version utility.

[![Go Reference](https://pkg.go.dev/badge/github.com/ardnew/over.svg)](https://pkg.go.dev/github.com/ardnew/over)
[![Go Report Card](https://goreportcard.com/badge/github.com/ardnew/over)](https://goreportcard.com/report/github.com/ardnew/over)

`over` validates [Semantic Versioning 2.0.0](https://semver.org) strings and increments version components. It operates entirely via argv/stdin and stdout — no config files, no project conventions, no runtime dependencies. Pipe a version in, get a version out.

## Install

### Transient command — install to cache and run in one step

```sh
go run github.com/ardnew/over@latest [flags] <version>
```

### As a `go.mod` tool dependency (Go 1.24+)

> [!IMPORTANT]
> This method installs `over` as a module dependency and can only be used within the module that requires it.

```sh
go get --tool github.com/ardnew/over@latest
```

Then run with `go tool`:

```sh
go tool over [flags] <version>
```

### Site-wide install from source

```sh
go install github.com/ardnew/over@latest
```

### From release

Download a prebuilt binary from [Releases](https://github.com/ardnew/over/releases) and place it on your `PATH`.

## Usage

```
over [flags] [version]
```

Input is read from the first non-flag argument or from **stdin**. If no bump flags are given, the input is validated and the recognized version is printed. If bump flags are given, the specified component(s) are incremented.

### Flags

| Flag | Description |
|------|-------------|
| `--major` | Increment major version (repeatable) |
| `--minor` | Increment minor version (repeatable) |
| `--patch` | Increment patch version (repeatable) |
| `--prerelease` | Retain pre-release label when bumping |
| `--build` | Retain build metadata when bumping |

## Examples

### Validate a version

```sh
$ over 1.2.3
1.2.3
```

### Extract a version from surrounding text

```sh
$ over "release/v3.8.1-rc.2 is ready"
3.8.1-rc.2
```

### Bump patch

```sh
$ over --patch 0.4.1
0.4.2
```

### Bump minor (resets patch)

```sh
$ over --minor 1.2.3
1.3.0
```

### Bump major (resets minor and patch)

```sh
$ over --major 2.5.9
3.0.0
```

### Combine bumps

```sh
$ over --major --minor --patch 1.2.3
2.1.1
```

### Repeat a bump

```sh
$ over --patch --patch --patch 0.0.0
0.0.3
```

### Pipe from stdin

```sh
$ cat VERSION | over --minor
0.2.0
```

```sh
$ git describe --tags | over --patch
1.0.1
```

### Pre-release and build metadata

By default, bumping strips pre-release and build metadata:

```sh
$ over --patch 1.0.0-alpha+build.42
1.0.1
```

Retain the pre-release label with `--prerelease`:

```sh
$ over --patch --prerelease 1.0.0-alpha
1.0.1-alpha
```

Retain build metadata with `--build`:

```sh
$ over --minor --build 2.0.0+exp.sha.5114f85
2.1.0+exp.sha.5114f85
```

Retain both:

```sh
$ over --major --prerelease --build 1.0.0-beta+exp
2.0.0-beta+exp
```

### Exit code

`over` exits **0** on success and **1** if the input contains no valid semver string:

```sh
$ over "no version here"; echo $?
1
```

## License

[MIT](LICENSE)
