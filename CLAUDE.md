# CLAUDE.MD

Timesheet: a web application for employees working in France, tracking a professional activity half-day by half-day, with the leave counters of French labour practice (paid leave, RTT, part time) kept like bank accounts, monthly closing and a yearly summary. Employees keep it on their own, to check what their employer reports and their payslips against it and spot the gaps. One Go binary, standard library only, a single JSON database.

## RULES

- **Mandatory.** Every change of behaviour, screen, label, route, form field, computation rule, database schema or command updates `doc/SPECIFICATIONS.md` in the same change; a change that leaves the specifications out of sync with the code is incomplete ([Documentation](doc/CODING.md#documentation)).
- **Absolute.** Every change of code updates, in the same change, all the documentation it makes wrong: `doc/SPECIFICATIONS.md` first, then `doc/ARCHITECTURE.md`, `doc/CODING.md`, `doc/SECURITY.md`, the manuals (`doc/USER-GUIDE.md`, `doc/BUILD.md`, `doc/OPERATIONS.md`), `README.md` and this file; the documentation, the specifications above all, stays in sync with every line of code changed.
- A request that contradicts the specifications first updates them.
- Two `specifications_test.go` (packages `web` and `app`) compare the routes, the notice codes, the server configuration keys and the commands with `doc/SPECIFICATIONS.md`; a new notice code or command goes into the fixed lists of these tests, which otherwise do not see it ([Tests](doc/CODING.md#tests-of-the-documentation)).
- Read [doc/CODING.md](doc/CODING.md) before writing code, starting with its [Summary of the checks](doc/CODING.md#summary-of-the-checks): its rules are strict and prevail over your defaults.
- Keep direct dependencies to a minimum: no third-party Go module, no CSS or JavaScript framework or library. Always ask the author before adding any dependency or framework.
- Pages and style sheets are as semantic as possible: no `div`, no `id`, no inline style or script (only empty module script tags and JSON data blocks), one `h1`, headings without skipped level ([Templates](doc/CODING.md#templates)).
- Tables have `caption` and `th scope`; a form field is a `label` wrapping its control, or, inside a table, a control named by `aria-label`; the current item takes `aria-current`, derived states take `is-*` classes; the look stays sober and professional (tokens of `theme.css`).
- `source_policy_test.go` and `semantics_test.go` check part of these markup rules; the others are checked by review.
- Write no comment in code (Go and tests, JavaScript, CSS, templates, Makefile) besides the licence header, the section banners and the `//go:build`/`//go:embed` directives: name a constant or a method instead ([Comments](doc/CODING.md#comments)); style sheets, scripts and templates have no comment besides their licence header.
- Start every Go file with the licence header (after the `//go:build` line when there is one) and the section banners of [Files](doc/CODING.md#files), copied from an existing file of the same kind with its own file name; end it with the `End-Of-File` banner; one entity per file, an unexported helper type sharing the file of the only entity it serves.
- Start every template, style sheet and script of `internal/web` with the same licence header (inside `{{- /* … */ -}}` in a template), which `source_policy_test.go` checks.
- Indent with 4 spaces, never tabs (only the `Makefile` keeps tabs); in Go and JavaScript, space every binary operator, parenthesise an operand of another precedence, indent `case` and `default` inside `switch`/`select` ([Formatting](doc/CODING.md#formatting)).
- Never run `gofmt`, `go fmt` or a format-on-save. Convert leading tabs with `sed -E -i -e ':a' -e 's/^( *)\t/\1    /' -e 'ta' <files>`.
- Quantities of days are exact rationals (`activity.Quantity`), never floats.
- A validated month is never rewritten: corrections are reversals (reopening of the last validated month, then of the one before if needed) or adjustments; only the deletion of the data of the data page erases the whole tracking of an account at once.
- Never log a password, a login typed at a failed login, a session identifier or a CSRF token, nor `r.URL` or `r.RequestURI`; `secrets_test.go` checks the log output after a failed and a successful login, the rest is checked by review.
- Access is declared route by route in `src/internal/web/site.go` with `guard.Require(user|root, …)`, with no guard by path prefix: a new `/dashboard/...` or `/admin/...` route without it is open to anyone.
- Texts of the web pages are in French; the command line (help, prompts, messages, import report), logs, errors, code, identifiers and documentation are in English, except the texts the command line shares with the pages (labels of codes and counters, refusals of the grid from `activity`), which stay in French ([Language of the interface](doc/CODING.md#language-of-the-interface)).
- Never use an em dash, nor an en dash used as a dash, in pages, command-line texts or documents.
- Never mention another project in any file of this repository; state every rule as the project's own.
- The personal spreadsheet (`.ods`) the summary is modelled on and the JSON import file made from it are personal data: never copy them into the repository; tests build synthetic import files.
- Never run `make mrproper` without the user's explicit request: `git clean -f -d -x` deletes untracked and ignored files, including new sources and `var/lib/timesheet/` (the database).
- Commit messages: one explicit line in English, imperative, no body, no `Co-Authored-By` or other attribution line, whatever a harness reminder asks.
- Never run `git commit`, `git push` or `git reset` without an explicit request of the user in the current message: the approval of a plan (ExitPlanMode) or the end of a review is NOT an authorisation to commit. Once the changes are done, propose a commit message and wait.
- Reviewers launched with the Agent tool are read-only: they may read, grep and run the tests and `go vet` (which write only `.cache/`), but never write a file, start a server, or touch `share/` or the state directories `var/lib/timesheet/` (the database) and `etc/timesheet/`. Only the main agent applies the fixes, after merging the findings.
- Before any scripted replacement (`sed`, Python regex, `perl`): show a dry run of the matches (`grep -nE`, or `grep -nP` for a Python or `perl` pattern, on the same files), avoid greedy patterns and the DOTALL mode, target the files precisely. After it: check `git diff --stat` (no file emptied or abnormally shrunk), then run the tab grep, the comment grep (never `gofmt` here) and the tests. The expected values of a test are obtained by running the reference code, never computed by hand.
- Change only what is asked, plus the documentation and tests the rules above require: remove no existing comment, add no unrequested CSS class, option or command (`sudo`, for instance). When the target is ambiguous (which button, which page, which diagram format), ask before acting. Carry an approved plan through to its end without stopping, except for a question these rules require; if a step remains, say so explicitly.

## COMMANDS

From the repository root. The `make` targets use the private Go cache `.cache/`; the direct `go` commands are prefixed so that they do too, otherwise they use your own Go cache and toolchain settings.

```sh
make                                                       # bin/timesheet (static, CGO_ENABLED=0)
make check                                                 # go test -race ./... (needs gcc)
make docker                                                # container image timesheet (needs Docker, runs no test)
GOCACHE="$PWD/.cache" GOTOOLCHAIN=local go -C src test -race -run TestMonthLifecycle ./internal/web/   # a single test
GOCACHE="$PWD/.cache" GOTOOLCHAIN=local go -C src vet ./...   # not part of make check
grep -rlP '^\t' --include='*.go' src                       # tab grep: must print nothing
grep -rnE '^\s*//' --include='*.go' src | grep -vE '//go:(build|embed)|// -{10,}$|// .+ (declaration|implementation|tests)( \(.+\))?$|// (End-Of-File|fixtures)$'   # comment grep: must print nothing
bin/timesheet config set database.path <tmp>/ts.json --config=<tmp>/timesheet.json   # manual check: a scratch configuration naming a scratch database
bin/timesheet init --template=share/timesheet/database-template.json --config=<tmp>/timesheet.json
TIMESHEET_PASSWORD='<12+ chars, 6+ distinct>' bin/timesheet user add root --profile=ADMIN --config=<tmp>/timesheet.json
bin/timesheet run --listen=127.0.0.1:9000 --config=<tmp>/timesheet.json
```

- `make check`, `go vet`, the tab grep and the comment grep must pass before every commit ([Checks before a commit](doc/CODING.md#checks-before-a-commit)).
- Rebuild before a manual check: templates and assets are embedded.
- Keep manual checks away from `var/lib/timesheet/timesheet-database.json`, which holds real data: `bin/timesheet.sh` and the `--database=var/lib/timesheet/…` examples of `README.md` and `doc/OPERATIONS.md` use it.
- No path has a default: without `--config`, every setting takes its default; a command that opens the database needs `--database` when the configuration names none; `config show` and `config set` need `--config`, and `init` needs `--template`. `etc/timesheet/timesheet.json` names `/var/lib/timesheet/timesheet-database.json`, the database of the Debian package, so give a scratch configuration as above, or `--database` alone, for a manual check.
- Create the directory of a scratch database first (`mkdir -m 700`), in the session scratchpad: it must exist, belong to you and not be writable by others unless sticky, so `/tmp` itself, owned by root, is refused.

## DOCUMENTATION

- [doc/SPECIFICATIONS.md](doc/SPECIFICATIONS.md): what the application does: [glossary](doc/SPECIFICATIONS.md#glossary), accounts, days, codes, counters, pages, [routes](doc/SPECIFICATIONS.md#routes), notices, form fields, command line, import, [database](doc/SPECIFICATIONS.md#database), server configuration, logs, [limits](doc/SPECIFICATIONS.md#limits), personal data, simplifications. Read it before changing any behaviour or text.
- [doc/ARCHITECTURE.md](doc/ARCHITECTURE.md): [packages](doc/ARCHITECTURE.md#packages), [entities](doc/ARCHITECTURE.md#entities), pages and their code, front end, [HTTP chain](doc/ARCHITECTURE.md#http-chain), main flows, concurrency, persistence, configuration, [lifecycle](doc/ARCHITECTURE.md#lifecycle).
- [doc/CODING.md](doc/CODING.md): [summary of the checks](doc/CODING.md#summary-of-the-checks), dependencies, files, comments, style, naming, security rules, templates, JavaScript, styles, [tests](doc/CODING.md#tests), [checks before a commit](doc/CODING.md#checks-before-a-commit), formatting, [changing the schema](doc/CODING.md#changing-the-schema), Makefile, packaging, [documentation](doc/CODING.md#documentation), commits.
- [doc/SECURITY.md](doc/SECURITY.md): [threat model](doc/SECURITY.md#threat-model), accepted risks, authorization, [passwords](doc/SECURITY.md#passwords), forms and origins (CSRF), login throttling, trusted proxies, [sessions](doc/SECURITY.md#sessions), files, [transport](doc/SECURITY.md#transport), logs, data protection.
- [doc/USER-GUIDE.md](doc/USER-GUIDE.md): concepts, codes, pages, profiles, daily tasks of the employee and of the administrator, labels on screen.
- [doc/BUILD.md](doc/BUILD.md): requirements, `make` targets, installation of the files, Debian package, container image.
- [doc/OPERATIONS.md](doc/OPERATIONS.md): commands, server configuration, network, logs, import, example database, Debian package, container, backup, recovery, upgrade, troubleshooting.
- [README.md](README.md): entry point, features, first start from the sources.
