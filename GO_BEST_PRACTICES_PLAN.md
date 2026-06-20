# Go Best-Practices / Idiomatic-Go Plan

Goal: modernize codebase without breaking current API users.

## Current contract to preserve

Do **not** change these unless separate versioned API planned:

- Routes: `GET /all`, `GET /next`, `GET /template`, `GET /is/{date}`, `GET /left`, `GET /make?year=YYYY`, invalid-route fallback.
- JSON field names/types:
  - `Holiday`: `date` as JSON time string, `name` as string.
  - `NextHoliday`: `name`, `date`, `isToday`, `daysUntil`.
  - `/is/{date}`: `{"isHoliday": bool}`.
  - invalid route: `status`, `message`, `valid_routes`.
- Existing success statuses: `200` for valid routes.
- Existing bad input statuses/bodies: `/is/bad` -> `400` + `error parsing date`; `/make?year=bad` -> `400` + `error parsing year`.
- Existing CORS header: `Access-Control-Allow-Origin: *`.
- Existing date semantics: public dates remain same civil holiday dates.

## Audit result

- Official latest stable Go checked on 2026-06-20: `go1.26.4` from `https://go.dev/dl/?mode=json`.
- Current module: `go 1.26.4` in `go.mod`.
- `go test ./...`: passes under Go 1.26.4 with route, holiday, CIDR, and async logging tests.
- `go vet ./...`: passes under Go 1.26.4.
- `staticcheck ./...`: passes with latest `staticcheck` under Go 1.26.4.
- `go fix -diff ./...`: no suggested modernizer diff under Go 1.26.4.
- `gofmt -l .`: no output after formatting.
- Docker build: passes with `golang:1.26.4-alpine` builder and distroless runtime.
- Go 1.26 release-note items relevant here: revamped `go fix` modernizers, `new(expr)` syntax, default Green Tea GC, experimental `goroutineleak` profile, `errors.AsType`, `slog.NewMultiHandler`.
- Go 1.25 release-note items relevant here: `testing/synctest`, `sync.WaitGroup.Go`, container-aware `GOMAXPROCS`, new `go vet` `hostport`/`waitgroup` analyzers, experimental `encoding/json/v2`.

## Plan: implement one checkbox per PR/commit

### P0 — Upgrade toolchain to Go 1.26.4

- [x] Update `go.mod` from `go 1.24.3` to `go 1.26.4`.
  - This is deployed app code, so latest supported toolchain is OK.
  - Non-breaking guard: Go 1 compatibility promise; route contract tests must pass before deploy.

- [x] Pin Docker builder image to `golang:1.26.4-alpine`.
  - Avoid floating `golang:alpine` changing compiler unexpectedly.
  - Non-breaking guard: build-tool change only; runtime image stays distroless.

- [x] Add/adjust GitHub Actions Go setup for Go 1.26.4 before Docker deploy.
  - Use `actions/setup-go@v5` with `go-version: '1.26.4'` or `go-version-file: go.mod`.
  - Non-breaking guard: CI only.

- [x] Re-run Go 1.26.4 verification after version bump.
  - Commands: `gofmt -l .`, `go test ./...`, `go vet ./...`, `go fix -diff ./...`, `staticcheck ./...`.
  - Non-breaking guard: no deploy if diff/tests fail.

- [x] Keep `GOEXPERIMENT` unset in production by default.
  - Go 1.26 already enables Green Tea GC by default; extra experiments should be opt-in only.
  - Non-breaking guard: avoid experimental runtime/API behavior in prod.


### P0 — Safety net before refactor

- [x] Add route contract tests using `httptest`.
  - Cover `/all`, `/make?year=2027`, `/is/2025-06-11`, `/is/2025-01-01`, `/is/bad`, `/next`, `/template`, `/left`, invalid route.
  - Assert status, CORS, content type, JSON field names, known holiday dates/names.
  - Non-breaking guard: tests encode current public behavior before code changes.

- [x] Add deterministic clock seam for tests only.
  - Keep exported funcs unchanged.
  - Add unexported `nowFunc` or small clock interface in `holiday` package.
  - Non-breaking guard: prod still uses `time.Now`; tests can freeze date.

- [x] Add CI test job before deploy job.
  - Run on Go 1.26.4: `gofmt -l .`, `go test ./...`, `go vet ./...`, `go fix -diff ./...`, `staticcheck ./...`.
  - Non-breaking guard: no runtime behavior change.

### P1 — HTTP server hardening

- [x] Replace package-level default mux with explicit `http.NewServeMux()`.
  - Register same patterns: `GET /all`, `GET /next`, `GET /template`, `GET /is/{date}`, `GET /left`, `GET /make`, `/`.
  - Non-breaking guard: same routes, same handlers.

- [x] Replace `http.ListenAndServe` with configured `http.Server`.
  - Add `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout` with generous values.
  - Non-breaking guard: normal clients unaffected; slowloris protection added.

- [x] Add graceful shutdown on `SIGINT`/`SIGTERM`.
  - Use `server.Shutdown(context.WithTimeout(...))`.
  - Non-breaking guard: only deployment shutdown behavior improves.

- [x] Validate `PORT` at startup.
  - Keep default `3002`.
  - Log fatal/error for invalid port.
  - Non-breaking guard: valid existing env keeps working.

- [x] Build server address with `net.JoinHostPort("", port)`.
  - Aligns with Go 1.25+ `go vet` `hostport` guidance; handles IPv6-style hosts if host added later.
  - Non-breaking guard: empty host + port still listens on same port.

### P1 — Response helpers / handler correctness

- [x] Add `writeJSON(w, status, value)` helper.
  - Set `Content-Type: application/json` and CORS consistently.
  - Preserve existing JSON shape and success status.
  - Non-breaking guard: contract tests compare responses.

- [x] Add `writeTextError(w, status, body)` helper for current text errors.
  - Preserve exact bodies: `error parsing date`, `error parsing year`.
  - Non-breaking guard: error response body unchanged.

- [x] Handle all template execution errors.
  - Render into `bytes.Buffer`, then write to client only after successful `Execute`.
  - Log errors; return `500` on template failure.
  - Non-breaking guard: successful `/template` and `/left` HTML output unchanged.

- [x] Stop writing `http.StatusOK` before template execution.
  - Let successful write imply `200` after buffer succeeds.
  - Non-breaking guard: success status remains `200`.

- [x] Check/log errors from `w.Write` and `template.Execute` where practical.
  - Non-breaking guard: no normal-response change.

### P1 — Outbound HTTP calls

- [x] Add package-level HTTP clients with timeouts for Giphy and IPInfo.
  - Avoid `http.Get` default client.
  - Non-breaking guard: same URLs/data, bounded waits.

- [x] Make outbound calls context-aware.
  - Add `FetchIPInfoContext(ctx, ip)` / `FetchGifURLContext(ctx)`.
  - Keep old exported funcs as wrappers if needed.
  - For detached async logging, use `context.WithoutCancel` + `context.WithTimeout` if request cancellation should not kill logging.
  - Non-breaking guard: existing exported funcs/endpoints continue compiling.

- [x] Check outbound HTTP status codes.
  - Treat non-2xx as errors.
  - Non-breaking guard: success behavior unchanged; failure no longer panics/silently decodes bad body.

- [x] Limit decoded response body size from external APIs.
  - Use `io.LimitReader` before JSON decode.
  - Non-breaking guard: valid small API responses unchanged.

- [x] Build external URLs with `net/url`, not `fmt.Sprintf`.
  - Escape IP/tag/token safely.
  - Non-breaking guard: valid IPs and tags generate same request meaning.

- [x] Use Go 1.26 `new(expr)` only where it improves optional pointer clarity.
  - Example after Giphy error refactor: return `new(url), nil` for `*string` rather than pointer to nested decoded struct field.
  - Non-breaking guard: same `*string` value returned on success.

### P1 — Logging path safety

- [x] Stop passing full `*http.Request` into goroutine.
  - Copy method, URL string, proto header, IP first; pass small struct to logger.
  - Non-breaking guard: log fields remain same.

- [x] Add fallback IP source.
  - Keep `X-Forwarded-For` first.
  - If absent, use `r.RemoteAddr` host.
  - Non-breaking guard: proxy deployments keep same IP behavior.

- [x] Skip IPInfo fetch when IP missing/invalid or `IP_INFO_TOKEN` empty.
  - Log local request without geo or skip external lookup.
  - Non-breaking guard: endpoint responses never depended on IPInfo.

- [x] Replace ignored CIDR errors with explicit validation.
  - If `MY_CIDR` empty: no CIDR skip.
  - If invalid: log config error once.
  - Non-breaking guard: valid existing CIDR behavior unchanged.

- [x] Replace `net.ParseIP`/`net.ParseCIDR` usage with `net/netip`.
  - Use `netip.ParseAddr` and `netip.ParsePrefix` inside `IP.IsInCIDR`.
  - Non-breaking guard: valid IPv4/IPv6 and CIDR inputs keep same behavior; add tests for empty/invalid inputs.

- [x] Use bounded async logging.
  - Option A: buffered channel + worker.
  - Option B: keep goroutine but add context timeout.
  - Non-breaking guard: endpoint latency stays low; no unbounded hangs.

### P1 — Panic removal / resilience

- [x] Remove `panic` from `giphy.FetchGifURL` path.
  - On failure, render `/template` without GIF and log error.
  - Non-breaking guard: `/template` still returns HTML on Giphy outage instead of crashing server.

- [x] Make JSON decode errors explicit in Giphy/IPInfo.
  - Return error; no ignored decode result.
  - Non-breaking guard: success output unchanged.

### P2 — Holiday package idioms + correctness guard

- [x] Keep exported function signatures for now.
  - Do not change `MakeHolidaysByYear(year int) *Holidays`, `FindUpcomingHoliday() *NextHoliday`, etc. in this pass.
  - Non-breaking guard: downstream Go importers keep compiling.

- [x] Protect cached holiday slices from mutation.
  - Store canonical slice in cache; return clone from `MakeHolidaysByYear` using `slices.Clone`.
  - Non-breaking guard: JSON output same; callers cannot corrupt cache.

- [x] Replace `sort.SliceStable` with `slices.SortStableFunc`.
  - Use `time.Time.Compare` for type-safe sorting.
  - Non-breaking guard: sorted order stays same.

- [x] Make `HolidayDateInCOT` explicit.
  - Replace `.In(cotLocation).Add(5*time.Hour)` with `time.Date(h.Date.Year(), h.Date.Month(), h.Date.Day(), 0, 0, 0, 0, cotLocation)`.
  - Non-breaking guard: same civil dates; easier to reason/test.

- [x] Add unit tests for Colombian holiday dates.
  - Cover fixed holidays, Monday-shift holidays, Easter-derived holidays, 2026+ Chiquinquirá holiday.
  - Non-breaking guard: prevents date regressions.

- [x] Add unit tests for `DaysUntil`, `IsToday`, `FindNext`, `GetRemaining` with frozen clock.
  - Preserve current semantics, including whether today's holiday appears in `/left`.
  - Non-breaking guard: detects behavior drift.

- [x] Replace constructor internals with composite literals.
  - `return NextHoliday{Name: name, Date: date, IsToday: isToday, DaysUntil: daysUntil}`.
  - `return Holiday{Date: date, Name: name}`.
  - Non-breaking guard: same exported funcs, same output.

- [x] Rename constructor params to idiomatic lowerCamel.
  - `is_today` -> `isToday`, `days_until` -> `daysUntil`.
  - Non-breaking guard: param names not part of Go API compatibility.

- [x] Use typed errors where handler behavior branches on error kind.
  - Prefer `errors.Is`; use Go 1.26 `errors.AsType[T]` only when a concrete error type is needed.
  - Non-breaking guard: response status/body remains unchanged; internals become easier to test.

### P2 — Naming / package / file-structure cleanup

- [x] Do naming audit before renames.
  - Inventory exported identifiers, filenames, package names, env vars, JSON fields, route names.
  - Classify each as public contract vs internal-only.
  - Non-breaking guard: public contract names stay, or get compatibility wrappers.

- [x] Rename internal HTTP handlers to idiomatic unexported names.
  - `HandleAllRoute` -> `handleAll`, `HandleNextRoute` -> `handleNext`, `HandleTemplateRoute` -> `handleTemplate`, `HandleIsRoute` -> `handleIs`, `LeftHandler` -> `handleLeft`, `MakeHandler` -> `handleMake`, `HandleInvalidRoute` -> `handleInvalidRoute`.
  - Update mux registration only; routes unchanged.
  - Non-breaking guard: endpoints and responses unchanged. If any external tests import handler funcs, keep exported wrapper funcs temporarily.

- [x] Rename response/data types to clearer names.
  - `InvalidRoute` -> `invalidRouteResponseBody` or `invalidRouteResponse`.
  - `LeftHolidays` -> `leftHolidayView`.
  - `IPInfoLite` fields: keep JSON tags, but internal names can become Go-initialism idiomatic where needed.
  - Non-breaking guard: JSON tags unchanged.

- [x] Rename local vars for clarity and idiomatic Go.
  - Examples: `h` -> `holidays` or `holidayItem` depending scope, `n` -> `nextHoliday`, `t` -> `parsedDate` or `now`, `d` -> `holidayDate`, `m` -> `body`, `p` -> `proto`, `s` -> `isOwnIP`.
  - Avoid overlong names in tiny scopes.
  - Non-breaking guard: local-only changes.

- [x] Rename packages only when value beats churn.
  - `templateinfo` can become `viewmodel` or be removed by local struct if only used for templates.
  - Keep `holiday` package name unless larger domain restructure happens.
  - Non-breaking guard: if package import path could be used by others, leave wrapper/deprecation path or delay rename to v2.

- [x] Consider moving external API clients into focused internal packages.
  - `giphy` -> `internal/giphyclient` or keep `giphy` if public import compatibility matters.
  - `ipinfo.go` -> `internal/ipinfo` or `internal/ipinfoclient`.
  - Non-breaking guard: HTTP API unchanged; public Go package paths preserved unless wrappers added.

- [x] Split large/ambiguous files into role-based files.
  - `handlers.go` -> `routes.go`, `handlers_json.go`, `handlers_html.go`, `responses.go`, `templates.go` as needed.
  - `log.go` -> `request_logging.go`.
  - `slog_split.go` -> `split_slog_handler.go` or `level_split_handler.go`.
  - `holiday/structs.go` -> `holiday/types.go`.
  - `giphy/structs.go` -> `giphy/types.go`.
  - `templateinfo/struct.go` -> `templateinfo/types.go` if package remains.
  - Non-breaking guard: file names do not change Go API; package names/import paths require care above.

- [x] Rename constructor funcs or remove unnecessary constructors.
  - `NewNextHoliday` can remain for compatibility, but internals should use composite literals where clearer.
  - If removing from internal use, keep exported func until major version decision.
  - Non-breaking guard: exported funcs remain callable.

- [x] Normalize acronym/initialism style.
  - Use `URL`, `IP`, `CIDR`, `JSON`, `HTTP` in exported/internal Go identifiers.
  - Examples: `gifURL`, `requestIP`, `envCIDR`, `writeJSON`.
  - Non-breaking guard: JSON tags/env var names unchanged.

- [x] Keep env var names unchanged.
  - Preserve `PORT`, `IP_INFO_TOKEN`, `GIPHY_KEY`, `MY_CIDR`.
  - Only rename Go variables holding them.
  - Non-breaking guard: deployments keep working.

- [x] Keep route path names unchanged.
  - Do not rename `/is`, `/make`, `/left`, etc.
  - Better internal names allowed; public URLs unchanged.
  - Non-breaking guard: current API clients keep working.

### P2 — Small Go style cleanup

- [x] Run `gofmt -w` on all Go files.
  - Current targets: `giphy/structs.go`, `ipinfo.go`.
  - Non-breaking guard: formatting only.

- [x] Rename all local all-caps vars that are not constants.
  - `PORT` -> `port`, `KEY` -> `key`, `GIPHY_QUERY` -> `giphyURL`.
  - Non-breaking guard: local-only rename.

- [x] Replace fixed `months` and `weekDays` maps with arrays/slices or helper funcs.
  - Keep same Spanish strings except typo fix below if accepted.
  - Non-breaking guard: rendered text stays same.

- [x] Fix Spanish typo in weekday text.
  - `Míercoles` -> `Miércoles`.
  - Non-breaking guard: HTML content correction only; no JSON/API shape change.

- [x] Centralize common headers.
  - CORS + content type helper/middleware.
  - Non-breaking guard: keep `Access-Control-Allow-Origin: *`.

### P2 — API contract notes

- [x] Fix README examples to match actual JSON.
  - `/make` sample currently appears to swap `name` and `date`.
  - Non-breaking guard: text only.

- [x] Document all current public routes in README.
  - Add `/template`, `/left`, invalid route behavior.
  - Non-breaking guard: text only.

- [x] Encode status codes, response schemas, and examples in route contract tests.
  - Non-breaking guard: tests only.

### P2 — Build / Docker / release hygiene

- [x] Pin Docker build image to target Go version.
  - Use `golang:1.26.4-alpine`.
  - Non-breaking guard: reproducible builds.

- [x] Decide whether to add `toolchain go1.26.4`.
  - Prefer no `toolchain` line if CI/Docker already pin exact Go; add it only to help local devs auto-download matching toolchain.
  - Non-breaking guard: build tooling only.

- [x] Improve Docker layer cache.
  - Copy `go.mod`/`go.sum`, run `go mod download`, then copy source.
  - Non-breaking guard: image output same.

- [x] Add `.dockerignore`.
  - Exclude `.git` and local temp files.
  - Non-breaking guard: smaller build context only.

- [x] Build binary with reproducible/smaller flags.
  - `go build -trimpath -ldflags="-s -w" -o api .`.
  - Non-breaking guard: same runtime behavior.

- [x] Run container as non-root.
  - In distroless: `USER nonroot:nonroot`.
  - Non-breaking guard: app listens on high port `3002`, no privileged bind needed.

- [x] Add `EXPOSE 3002`.
  - Non-breaking guard: metadata only.

### P3 — Optional non-breaking additions

- [x] Add `/healthz` endpoint.
  - Additive only; do not change invalid-route behavior for existing paths.
  - Non-breaking guard: new route only.

- [x] Add structured request logging middleware.
  - Log method, path, status, duration; keep current geo fields if still needed.
  - Non-breaking guard: responses unchanged.

- [x] Add config struct loaded once at startup.
  - Fields: port, IPInfo token, Giphy key, MY_CIDR, timeouts.
  - Non-breaking guard: same env var names/defaults.

- [x] Optional: test async logging with Go 1.25+ `testing/synctest` if logging worker uses timers/goroutines.
  - Use `synctest.Test` / `synctest.Wait` for deterministic concurrency tests.
  - Non-breaking guard: tests only.

- [x] Optional: use Go 1.26 experimental `goroutineleak` profile in non-prod diagnostics.
  - Build/test with `GOEXPERIMENT=goroutineleakprofile` only outside prod.
  - Non-breaking guard: no prod experiment enabled by default.

## Things to avoid in this modernization pass

- Do not change route names or add `/v2` unless separate migration planned.
- Do not change JSON field names, casing, status codes, or date format.
- Do not replace text error bodies with JSON yet.
- Do not change exported package function return types yet, especially `*Holidays`.
- Do not remove `Access-Control-Allow-Origin: *`.
- Do not change holiday date list without tests and release notes.
- Do not switch production API JSON to experimental `encoding/json/v2` yet; public response contract matters more.
- Do not replace current stdout/stderr split logger with `slog.NewMultiHandler` unless duplicate logs are desired; `MultiHandler` fans out to all handlers, it does not split by level.
- Do not enable Go experiments in production without separate rollout/rollback plan.
