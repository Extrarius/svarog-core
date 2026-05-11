# Task: verify the observability pipeline

Goal: confirm that **traces**, **metrics**, and **logs** emitted by the application reach Grafana through the LGTM stack.

## 1. Bring the stack up

```bash
make up
```

This starts:

- `postgres` (5432)
- `otel-collector` (4317 gRPC, 4318 HTTP)
- `loki` (3100)
- `tempo` (3200 query, 4317 internal OTLP)
- `mimir` (9009)
- `grafana` (3000) — login `admin` / `admin` by default

Wait until all containers report healthy (`docker compose -f deploy/docker-compose.yml ps`).

## 2. Run the application

```bash
cp .env.example .env   # if you haven't already
make migrate
make run
```

The service should log a startup message via the configured slog handler and announce its gRPC / HTTP listen addresses.

## 3. Generate some traffic

```bash
# unauthenticated probe (expected: Unimplemented or NotFound — still produces a span)
curl -v http://localhost:8080/v1/auth/me
```

Or via gRPC:

```bash
grpcurl -plaintext localhost:9090 list
```

## 4. Check each pillar in Grafana

Open <http://localhost:3000>.

### Logs (Loki)

- **Explore** → datasource **Loki** → query:

  ```logql
  {service_name="svarog-core"}
  ```

- You should see structured JSON logs with `trace_id` / `span_id` attached.

### Traces (Tempo)

- **Explore** → datasource **Tempo** → search by service name `svarog-core`.
- Pick a recent trace. Verify it has at least one server span for the incoming request.
- If a log line carries the same `trace_id`, the Loki <-> Tempo correlation works.

### Metrics (Mimir)

- **Explore** → datasource **Mimir** → query:

  ```promql
  rate(http_server_request_duration_seconds_count{service_name="svarog-core"}[1m])
  ```

- Confirm a non-zero rate after generating traffic.

## 5. If something is missing

Diagnostic order:

1. Application logs: any OTel exporter errors? (look for `otlptracegrpc` / `otlpmetricgrpc` / `otlploggrpc`).
2. `docker compose -f deploy/docker-compose.yml logs otel-collector`: are receivers/exporters healthy?
3. `docker compose ... logs loki|tempo|mimir`: are they accepting writes?
4. Grafana datasource configuration: `deploy/grafana/provisioning/datasources/datasources.yaml` — URLs must point to the in-network service names (`loki:3100`, `tempo:3200`, `mimir:9009`).

## 6. Tear down

```bash
make down
```
