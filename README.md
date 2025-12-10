healthz - Health Aggregator
----------------------------
# Overview
Simple Go HTTP service that runs configured URL checks and exposes an aggregated health endpoint.
Use case is to run health checks on a set of services and aggregate their health status into a 
single endpoint. Useful for envoy health checks.


# Features
- Load checks from a YAML config file
- HTTP health endpoints:
    - `/healthz` — run all checks in all named groups (aggregate)
    - `/healthz/<name>` — run only the checks in the named group `<name>`
- CLI flags:
    - `--config` (default: `config.yml`) \- path to config file; single positional argument is accepted as a fallback
    - `--version` \- print version and exit

# Configuration (`config.yml`) example
~~~yaml
listen_address: ":9001"
checks:
  public_sites:
    - name: google
      url: https://www.google.com
      expected_status: 200
      timeout: 2s
    - name: example
      url: https://example.com
      expected_status: 200
      timeout: 1s
  internal_sites:
    - name: web
      url: https://web.internal.example.com
      expected_status: 200
      timeout: 2s
~~~

Notes:
- `listen_address` defaults to `:9001` if omitted.
- checks is a mapping of group names to lists of checks.
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
./healthz --version
~~~

# Health check
Aggregate - all groups
~~~bash
curl -s http://localhost:9001/healthz
~~~
Returns `200` and body `OK` when all checks pass for all groups. Returns `500` and body `FAIL` if any check fails.

Named group (run only otel1):
~~~bash
curl -s http://localhost:9001/healthz/public_sites
~~~
Returns 200 and body OK when all checks in the public_sites group pass. Returns 500 and body FAIL on the first failing check.

# License
GPL-3.0 License
