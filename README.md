# container-test-cli

Run declarative smoke tests against container images. The CLI reads a YAML file of commands to execute inside a container, asserts exit codes and output, and optionally emits a JSON report.

## Features
- Uses any container engine CLI (`docker`, `podman`, etc.) with per-test `run_args` and optional entrypoint override.
- Assertions on exit code, stdout/stderr substrings, and regex matches; supports negative stdout checks.
- Per-test environment variables, working directory, and timeouts (with a global default).
- `--fail-fast`, optional JSON report output, and colorized status (respecting `NO_COLOR`).
- Accepts either a root list of tests or a `{tests: []}` object for YAML configs.

## Requirements
- Container engine CLI available on the host (default: `docker`; override via `--engine` or `CONTAINER_TEST_ENGINE`).
- Go 1.22+ to build from source.

## Build
```sh
go build -o container-test-cli
```

## Usage
```sh
./container-test-cli --config tests.example.yaml --image curlimages/curl:latest
```

### Flags and env vars
All flags have environment variable counterparts:
- `--config` / `CONTAINER_TEST_CONFIG` (required): path to YAML file.
- `--image` / `CONTAINER_TEST_IMAGE` (required): container image reference.
- `--engine` / `CONTAINER_TEST_ENGINE` : container CLI (default `docker`).
- `--default-timeout` / `CONTAINER_TEST_DEFAULT_TIMEOUT`: per-test timeout in seconds (default 30).
- `--json-report` / `CONTAINER_TEST_JSON_REPORT`: path to write JSON results.
- `--fail-fast` / `CONTAINER_TEST_FAIL_FAST`: stop on first failure.
- `--debug` / `CONTAINER_TEST_DEBUG`: print constructed `run` commands.

## Test file format
The YAML can be either a top-level list or a map with a `tests` key. Command fields accept a single string (wrapped with `sh -c`) or a list (exec form).

```yaml
tests:
  - name: check-app-version
    exec: ["--version"]
    expect:
      stdout_regex: "^curl ?[0-9]+\\.[0-9]+\\.[0-9]+"

  - name: health-endpoint
    command: "curl -o /dev/null -s -w '%{http_code}' https://www.example.com/"
    expect:
      exit_code: 0
      stdout_contains: "200"
      timeout_seconds: 10
    run_args: ["--network=host"]
    entrypoint: ""          # optional override

  - name: env-variable
    exec: ["printenv", "HOME"]
    expect:
      stdout_contains: "/home/curl_user"

  - name: log-must-be-quiet
    command: "ls /tmp"
    expect:
      stdout_not_contains: "core"

  - name: skip-this
    skip: true
```

### Test fields
- `name`: Display name (default `test-<index>`).
- `exec` / `command`: Command to run inside the container (string or list). At least one is required.
- `skip`: Skip execution and mark as `SKIPPED`.
- `workdir`: Working directory inside the container.
- `env`: Map of environment variables passed with `-e`.
- `run_args`: Extra `engine run` flags (e.g., `--network=host`).
- `entrypoint`: Override container entrypoint.
- `timeout_seconds`: Per-test timeout; overrides `expect.timeout_seconds` and the global default.
- `expect`:
  - `exit_code` (default `==0`), supports `==`, `!=`, `>=`, `<=`, `>`, `<` (e.g., `">=0"`, `"!=1"`)
  - `stdout_contains` / `stderr_contains`: string or list; each must be present.
  - `stdout_not_contains`: string or list; must be absent.
  - `stdout_regex` / `stderr_regex`
  - `timeout_seconds`: Timeout applied to the test unless overridden by `timeout_seconds` at the test level.

## JSON report
When `--json-report` is set, results are written as an array of objects:
```json
[
  {
    "status": "PASSED",
    "name": "check-app-version",
    "stdout": "curl 8.17.0 ...",
    "stderr": "",
    "exit_code": 0
  },
  {
    "status": "FAILED",
    "name": "health-endpoint",
    "stdout": "200",
    "stderr": "",
    "exit_code": 0,
    "failures": ["stdout missing: \"201\""]
  }
]
```
