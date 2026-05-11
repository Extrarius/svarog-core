# Workflows

Place **multi-step processes** here: ordered sequences that reference `tasks/`, `checklists/`, or external tools (e.g. “ship a breaking proto change”).

Example workflow outline (create a real file when needed):

1. Update `api/proto/...`
2. Run `easyp lint` + `easyp breaking --against main`
3. Run `make proto-gen`
4. Follow [`tasks/new-endpoint.md`](../tasks/new-endpoint.md) for wiring
