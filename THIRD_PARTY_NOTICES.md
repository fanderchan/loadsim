# Third-Party Notices

This project uses open-source software.

The source code directly imports these primary dependencies:

- `github.com/spf13/cobra` `v1.8.0`
  License: Apache License 2.0
- `github.com/shirou/gopsutil` `v3.21.11+incompatible`
  License: BSD 3-Clause License

The runtime also relies on:

- `github.com/spf13/pflag` `v1.0.5`
  License: BSD 3-Clause License

Notes:

- The full module graph is recorded in `go.mod` and `go.sum`.
- Additional indirect modules may appear there because of platform-specific support code or upstream test dependencies.
- License names above were verified from the local Go module cache used for this build environment.
