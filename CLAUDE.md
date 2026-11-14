# CLAUDE.MD

Timesheet: a web application for employees working in France, tracking a professional activity half-day by half-day, with the leave counters of French labour practice (paid leave, RTT, part time) kept like bank accounts, monthly closing and a yearly summary. One Go binary, standard library only, a single JSON database.

## RULES

- **Mandatory.** Every change of behaviour, screen, label, route, form field, computation rule, database schema or command updates `doc/SPECIFICATIONS.md` in the same change; a change that leaves the specifications out of sync with the code is incomplete ([Documentation](doc/CODING.md#documentation)).
- Any change also updates `README.md` and the files of `doc/` whose content it makes wrong.
- A request that contradicts the specifications first updates them.
- Two `specifications_test.go` (packages `web` and `app`) compare the routes, the notice codes and the commands with `doc/SPECIFICATIONS.md`; add a new notice or command to their fixed lists ([Tests](doc/CODING.md#tests-of-the-documentation)).
- Read [doc/CODING.md](doc/CODING.md) before writing code, starting with its [Summary of the checks](doc/CODING.md#summary-of-the-checks): its rules are strict and prevail over your defaults.
- Keep direct dependencies to a minimum: no third-party Go module, no CSS or JavaScript framework or library. Always ask the author before adding any dependency or framework.
- Pages and style sheets are as semantic as possible: no `div`, no `id`, no inline style or script (only empty module script tags and JSON data blocks), one `h1`, headings without skipped level ([Templates](doc/CODING.md#templates)).
- Tables have `caption` and `th scope`, labels wrap their controls, state goes through `aria-current` and `is-*` classes; the look stays sober and professional (tokens of `theme.css`).
- `source_policy_test.go` and `semantics_test.go` check part of these markup rules; the others are checked by review.
- Write no comment in code (Go and tests, JavaScript, CSS, templates, Makefile) besides the licence header, the section banners and the `//go:build`/`//go:embed` directives: name a constant or a method instead ([Comments](doc/CODING.md#comments)).
- A style sheet, a script or a template only has its licence header.
- Start every Go file with the licence header (after the `//go:build` line when there is one) and the section banners of [Files](doc/CODING.md#files), copied from an existing file of the same kind; end it with the `End-Of-File` banner; one entity per file.
- Start every template, style sheet and script of `internal/web` with the same licence header (inside `{{- /* … */ -}}` in a template), which `source_policy_test.go` checks.
- Indent with 4 spaces, never tabs (only the `Makefile` keeps tabs); in Go and JavaScript, space every binary operator, parenthesise an operand of another precedence, indent `case` and `default` inside `switch`/`select` ([Formatting](doc/CODING.md#formatting)).
- Never run `gofmt`, `go fmt` or a format-on-save. Convert leading tabs with `sed -E -i -e ':a' -e 's/^( *)\t/\1    /' -e 'ta' <files>`.
- Quantities of days are exact rationals (`activity.Quantity`), never floats.
- A validated month is never rewritten: corrections are reversals (reopening of the last validated month, then of the one before if needed) or adjustments.
- Never log a password, a login typed at a failed login, a session identifier or a CSRF token, nor `r.URL` or `r.RequestURI`; `secrets_test.go` checks the log output after a failed and a successful login, the rest is checked by review.
- Texts of the web pages are in French and never use the em dash; the command line (help, prompts, messages, import report), logs, errors, code, identifiers and documentation are in English ([Writing rules](doc/CODING.md#documentation)).
- Never mention another project in any file of this repository; state every rule as the project's own.
- The personal spreadsheet (`.ods`) the summary is modelled on and the JSON import file made from it are personal data: never copy them into the repository; tests build synthetic import files.
- Never run `make mrproper` without the user's explicit request: `git clean -f -d -x` deletes untracked and ignored files, including new sources and `var/lib/timesheet/` (the database).
- Commit messages: one explicit line in English, imperative, no body, no `Co-Authored-By` or other attribution line, whatever a harness reminder asks. Commit only when asked.

## COMMANDS

From the repository root. The `make` targets use the private Go cache `.cache/`; the direct `go` commands are prefixed so that they do too, otherwise they use your own Go cache and toolchain settings.

```sh
make                                                       # bin/timesheet (static, CGO_ENABLED=0)
make check                                                 # go test -race ./... (needs gcc)
GOCACHE="$PWD/.cache" GOTOOLCHAIN=local go -C src test -race -run TestMonthLifecycle ./internal/web/   # a single test
GOCACHE="$PWD/.cache" GOTOOLCHAIN=local go -C src vet ./...   # not part of make check
grep -rlP '^\t' --include='*.go' src                       # files indented with tabs: must print nothing
bin/timesheet init --database=<tmp>/ts.json                # manual check on a scratch database
TIMESHEET_PASSWORD='<12+ chars, 6+ distinct>' bin/timesheet user add root --profile=ADMIN --database=<tmp>/ts.json
bin/timesheet run --listen=127.0.0.1:9000 --database=<tmp>/ts.json
```

- `make check`, `go vet`, the tab grep and the comment grep must pass before every commit ([Checks before a commit](doc/CODING.md#checks-before-a-commit)).
- Rebuild before a manual check: templates and assets are embedded.
- Keep manual checks away from `var/lib/timesheet/timesheet-database.json`, which holds real data.
- Always give `--database` to a command that opens a database: without it, the path comes from `database.path` of `etc/timesheet/timesheet.json`, which is `/var/lib/timesheet/timesheet-database.json`, the database of the Debian package.
- The directory of a scratch database must belong to you and not be writable by others.

## DOCUMENTATION

- [doc/SPECIFICATIONS.md](doc/SPECIFICATIONS.md): what the application does: [glossary](doc/SPECIFICATIONS.md#glossary), accounts, days, codes, counters, pages, [routes](doc/SPECIFICATIONS.md#routes), command line, import, [database](doc/SPECIFICATIONS.md#database), server configuration, logs, [limits](doc/SPECIFICATIONS.md#limits), personal data, simplifications. Read it before changing any behaviour or text.
- [doc/ARCHITECTURE.md](doc/ARCHITECTURE.md): [packages](doc/ARCHITECTURE.md#packages), [entities](doc/ARCHITECTURE.md#entities), pages and their code, front end, [HTTP chain](doc/ARCHITECTURE.md#http-chain), main flows, concurrency, persistence, configuration, [lifecycle](doc/ARCHITECTURE.md#lifecycle).
- [doc/CODING.md](doc/CODING.md): [summary of the checks](doc/CODING.md#summary-of-the-checks), files, comments, style, naming, security rules, templates, JavaScript, styles, [tests](doc/CODING.md#tests), formatting, [changing the schema](doc/CODING.md#changing-the-schema), Makefile, packaging, [documentation](doc/CODING.md#documentation), commits.
- [doc/SECURITY.md](doc/SECURITY.md): [threat model](doc/SECURITY.md#threat-model), accepted risks, authorization, [passwords](doc/SECURITY.md#passwords), login throttling, trusted proxies, [sessions](doc/SECURITY.md#sessions), files, [transport](doc/SECURITY.md#transport), logs, data protection.
- [doc/USER-GUIDE.md](doc/USER-GUIDE.md): concepts, codes, pages, profiles, daily tasks of the employee and of the administrator, labels on screen.
- [doc/BUILD.md](doc/BUILD.md): requirements, `make` targets, installation of the files, Debian package.
- [doc/OPERATIONS.md](doc/OPERATIONS.md): commands, server configuration, network, logs, import, Debian package, backup, recovery, upgrade, troubleshooting.
- [README.md](README.md): entry point, features, first start from the sources.
