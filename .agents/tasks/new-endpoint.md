# Task: add a new gRPC + HTTP endpoint

Use this checklist when adding a new RPC to an existing service (e.g. `AuthService`).

## 1. Update the proto contract

- Edit `api/proto/<service>/v<N>/<service>.proto`.
- Add a new `rpc` method to the relevant service.
- Add request/response messages.
- Annotate the method with `google.api.http` for the REST mapping, for example:

  ```proto
  rpc UpdateProfile(UpdateProfileRequest) returns (UpdateProfileResponse) {
    option (google.api.http) = {
      patch: "/v1/users/me"
      body: "*"
    };
  }
  ```

- Run `easyp lint`. Fix any reported issues.
- Run `easyp breaking --against main` if you modified existing messages.

## 2. Regenerate stubs

```bash
make proto-gen
```

This writes Go/grpc/gateway stubs into `api/gen/go/` and Swagger into `api/gen/openapi/`.

## 3. Define / extend the use case (`internal/app`)

- Add the new method to `internal/app/handlers.go`.
- If the use case needs new collaborators (a new repo method, a clock, an email sender), declare them as **interfaces** in `internal/app/contracts.go`.
- Remember: `internal/app` may only import stdlib and other `internal/*` packages that are themselves stdlib-only.

## 4. Implement the port(s) (`internal/adapters/repo`)

- For new repository operations, write a pgx-based implementation in `internal/adapters/repo`.
- Keep all SQL strings inside this package.
- Add table-driven tests where practical.

## 5. Wire transport (`internal/api/grpc`)

- Implement the new method on the gRPC server struct that satisfies the generated `<Service>Server` interface.
- Translate the protobuf request to use case inputs, call the use case, translate the result.
- Surface canonical gRPC codes (`codes.InvalidArgument`, `codes.Unauthenticated`, etc.) only here — never inside `internal/app`.

## 6. Wire HTTP (`internal/api/gateway`)

In most cases the grpc-gateway runtime registers the new HTTP route automatically once the generated stub is registered. Verify:

- The gateway is started with `RegisterAuthServiceHandler` (or equivalent) — generated registration calls.
- Cookie middleware still applies if the endpoint requires authentication.

## 7. Composition root (`cmd/main.go`)

If the use case depends on a new adapter, construct it in `cmd/main.go` and inject it into the handlers struct.

## 8. Verify

```bash
make lint
make test
make run      # smoke test against localhost
```

Check the new HTTP route with `curl` and the new gRPC method with `grpcurl`.
