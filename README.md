healthz - Health Aggregator
----------------------------
# Overview
Simple Go HTTP service that runs configured URL checks and exposes an aggregated health endpoint.
Use case is to run health checks on a set of services and aggregate their health status into a 
single endpoint. Useful for envoy health checks.


# Features
- Load checks from a YAML config file
- HTTP health endpoint at `/healthz`
- CLI flags:
    - `--config` (default: `config.yml`) \- path to config file; single positional argument is accepted as a fallback
    - `--version` \- print version and exit

# Configuration (`config.yml`) example
~~~yaml
listen_address: ":9001"
checks:
  - name: google
    url: https://www.google.com
    expected_status: 200
    timeout: 2s
  - name: example
    url: https://example.com
    expected_status: 200
    timeout: 1s
~~~

Notes:
- `listen_address` defaults to `:9001` if omitted.
- `timeout` values use Go duration format (for example `500ms`, `1s`, `2s`).
- `expected_status` defaults to `200` if omitted.

# Build
Run:
~~~bash
go build -o healthz
~~~

# Run
Using default `config.yml`:
~~~bash
./healthz
~~~

Specify config path:
~~~bash
./healthz --config /path/to/config.yml
# or as a single positional argument:
./healthz /path/to/config.yml
~~~

Print version:
~~~bash
./health-aggregator --version
~~~

# Health check
Request:
~~~bash
curl -sS -w "%{http_code}\n" http://localhost:9001/healthz
~~~

Returns `200` and body `OK` when all checks pass; returns `500` and body `FAIL` if any check fails.

## License
GPL-3.0 License
