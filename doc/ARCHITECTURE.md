# ARCHITECTURE

This document describes how the code of Timesheet is organised and why: packages, entities, the HTTP chain, the main flows, concurrency, persistence and the lifecycle of the process.
It is written for the maintainer, for an AI agent and for a developer who joins the project.
It does not describe behaviour, which is specified in [SPECIFICATIONS.md](SPECIFICATIONS.md), nor the rules of writing code, which are in [CODING.md](CODING.md), nor the threats, which are in [SECURITY.md](SECURITY.md).

## CONTENTS

- [OVERVIEW](#overview)
- [REPOSITORY LAYOUT](#repository-layout)
- [PACKAGES](#packages)
- [ENTITIES](#entities)
- [PAGES AND THE CODE THAT SERVES THEM](#pages-and-the-code-that-serves-them)
- [FRONT END](#front-end)
- [HTTP CHAIN](#http-chain)
- [MAIN FLOWS](#main-flows)
- [CONCURRENCY](#concurrency)
- [PERSISTENCE](#persistence)
- [CONFIGURATION AND SERVER CONFIGURATION](#configuration-and-server-configuration)
- [LIFECYCLE](#lifecycle)
- [GLOSSARY](#glossary)

## OVERVIEW

Purpose and French specifics:

- Timesheet tracks the activity and the leave counters of employees working in France.
- The domain package `internal/activity` carries the French specifics:
  - the legal holidays of metropolitan France;
  - the rules of paid leave, RTT (récupération temps de travail) and part-time counters;
  - the French labels.

Delivery:

- The application is a single statically linked Go binary, `bin/timesheet`.
- It is built with the Go standard library only: there is no third-party code, Go or JavaScript.
- The HTML templates and the static assets are embedded in the binary with `go:embed`.
- There is no external service.
- The server speaks HTTP, or HTTPS (HTTP over TLS, Transport Layer Security) with a provided certificate pair.

Data:

- The activity codes, the counters, the accounts and the data of every account live in one JSON file, the database.
- The server configuration lives in a second file.
- The open sessions are saved in a third file, next to the database.

How the parts call each other:

```text
browser ──> HTTP chain (guards) ──> Router ──> AuthGuard ──> controllers ──> web.Repository ──> store.Database ──> JSON file
                                                                 │
                                                                 └──> view builders ──> HTML templates

command line ──> app.Application ──> commands ──> store.Database ──> JSON file
```

## REPOSITORY LAYOUT

```text
src/
  go.mod                          module timesheet, go 1.24
  cmd/timesheet/main.go           signal context, Application, os.Exit
  internal/activity/              domain (no HTTP, no I/O)
  internal/account/               logins, profiles, states, passwords, account record
  internal/store/                 SecureFile, Database (single JSON file), SessionFile
  internal/web/                   HTTP adaptation
    templates/                    HTML templates: views and partials (embedded)
    static/                       css, js, img (embedded)
  internal/app/                   command line, commands, application lifecycle
etc/timesheet/
  timesheet.json                  server configuration: address, origins, database path, timeouts, sessions, login, limits
share/timesheet/
  timesheet-database-template.json   configuration template read by init
  timesheet-import-template.json     fictitious example of an import file
var/lib/timesheet/                database of the repository, given with --database, its lock and the session file
debian/                           Debian packaging
.ci/                              GitLab CI configuration and the build scripts
```

| Directory | Content | Detailed in |
| --- | --- | --- |
| `debian/` | Native Debian package: `rules`, `control`, `changelog`, the systemd unit `timesheet.service`, and `postinst`, which creates the service account and the database | [CODING.md, Packaging](CODING.md#packaging) |
| `.ci/` | `gitlab-ci.yml`; the scripts run by the Makefile targets (`make-build.sh`, `make-check.sh`, `make-install.sh`, `make-clean.sh`); the CI scripts (`ci-setup-debian.sh`, `ci-build-package.sh`, `ci-cleanup.sh`) | [CODING.md, Makefile](CODING.md#makefile), [CODING.md, Packaging](CODING.md#packaging) |
| `_dist/` | At the repository root, not under `.ci/`: the packages built by `ci-build-package.sh` (ignored by git) | [CODING.md, Packaging](CODING.md#packaging) |
| `.cache/` | The Go build cache of the scripts (ignored by git) | [CODING.md, Makefile](CODING.md#makefile) |
| `var/lib/timesheet/` | The default database, its lock file and its session file (ignored by git, except `.gitkeep`) | [SPECIFICATIONS.md, Database](SPECIFICATIONS.md#database) |

## PACKAGES

### Import direction

```text
cmd → app → {web, store, activity, account}
store → {activity, account}
web → {activity, account}
```

- `activity` and `account` import nothing from the project.
- `web` never imports `app` or `store`: it works on the `web.Repository` interface, implemented by `store.Database` and injected through `SiteDependencies`.
- The tests of `web` use `store` to work on a real database.

### Roles

| Package | Role | May import (project packages) |
| --- | --- | --- |
| `internal/activity` | Domain: quantities, dates, configuration, calendar, holidays, counter bookkeeping, month closing, indicators, summary, colours, clocks | nothing |
| `internal/account` | Logins, profiles, states, password policy and hashes, account record, session record, account errors | nothing |
| `internal/store` | `SecureFile` (hardened private file, lock), `Database` (the single JSON database) and `SessionFile` (the saved sessions) | `activity`, `account` |
| `internal/web` | HTTP adaptation: HTTP chain of guards, sessions, controllers, view builders, HTML templates, assets | `activity`, `account` (`store` in tests) |
| `internal/app` | Command line, options, commands, application lifecycle | all |
| `cmd/timesheet` | `main` | `app` |

## ENTITIES

Each entity lives in its own file, named after the entity in `snake_case` (`MonthClosing` in `month_closing.go`).
The tables give one entity per row.
The column Kind says what the entity is:

| Kind | Meaning |
| --- | --- |
| value | Type with behaviour and unexported fields, or a typed string or map |
| record | Data with exported fields, persisted or rendered |
| service | Struct that carries behaviour and its dependencies |
| interface | Port declared by its consumer or by the domain |

### internal/activity

#### Values

| Entity | Kind | Role |
| --- | --- | --- |
| `Quantity` | value | Exact rational number of days (details below) |
| `TextPolicy` | service | Accepted free text: valid UTF-8, bounded length, no control, format, bidirectional, line or paragraph separator character, the zero-width joiners excepted |
| `Date` | value | Civil date, comparable, text-marshalled, with French labels |
| `Month` | value | Month of a year, comparable, text-marshalled, with French labels |
| `PeriodStart` | value | Yearly period start written `MM-DD`, comparable, text-marshalled |
| `NullableText` | value | Text written `null` when empty |
| `NullableTime` | value | Time written `null` when zero |
| `HalfDay` | value | Morning or afternoon |
| `DayKind` | value | Working day, weekend or holiday |
| `DayType` | value | Display type of a half-day: entered code, missing, empty, out of the tracking |
| `Color` | value | `#rrggbb` colour with its relative luminance and its readable ink, black or white |
| `ColorSet` | value | The colours a user gives to codes, over the colour of each code |

`Quantity`:

- numerator and denominator are `int64`, always normalised;
- every denominator divides `QuantityBase`, 63 000 000;
- sums are made on the least common multiple of the denominators, with overflow detection;
- it is read from `"2.08"` or `"7/12"` and written in a canonical form;
- `Format` displays `0,58` and `FormatDays` displays `1,5`.

`NullableText` and `NullableTime` exist so that every field of a record is always written.
The records of the database marshal through them: `DayEntry`, `ActivityCode`, `LedgerEntry`, `MonthEvent`, `MonthRecord` and `store.AccountRecord`.

#### Configuration

| Entity | Kind | Role |
| --- | --- | --- |
| `ActivityCode` | record | Code: `id`, label, consumed or fed counter, mandatory colour, archived flag; its effect is `Consuming` or `Feeding`. A code has no group: its counter says what it represents |
| `AccrualPeriod` | value | Accrual period of a counter: month or year |
| `Accrual` | record | Amount credited per month or per year |
| `CounterRule` | record | Counter rule: accrual, period start, carry over (with rounding up) or zeroing of its single balance |
| `Configuration` | record | Configuration section of the database (codes, counters, revision), with `Validate` (which returns the path of the faulty field) and lookups. `NewDefaultConfiguration` mirrors `share/timesheet/timesheet-database-template.json` |
| `ServerConfiguration` | record | The whole server configuration file: its sections `ServerSettings`, `DatabaseSettings`, `SessionSettings`, `LoginSettings`, `PasswordSettings` and `LimitSettings`, their defaults (`NewDefaultServerConfiguration`, whose encoding is `etc/timesheet/timesheet.json`), the bounds checked by `Validate`, including the links between keys, and the access by key `section.key` (`Keys`, `Value`, `Set`) used by `config show` and `config set` |
| `ServerSettings` | record | Section `server`: address, origins, trusted proxies, time zone, log level, TLS pair (`NullableText`, written `null` without a pair) and version, HSTS, timeouts and request sizes, with `Validate`, `Location` and `Level` |
| `DatabaseSettings`, `SessionSettings`, `LoginSettings`, `PasswordSettings`, `LimitSettings` | record | Sections `database`, `sessions`, `login`, `passwords` and `limits`, each with its defaults and `Validate` |
| `Duration` | value | Duration of the server configuration: read like `time.ParseDuration`, a whole number of milliseconds, written in its shortest form (`8h`, `1m`), shown in French (`8 h`, `1 min`) |
| `SettingCheck` | service | Bound checks of the server configuration, with the path `section.key` in the error |
| `TrustedProxies` | value | Addresses and prefixes of the proxies whose `X-Forwarded-For` header is believed, built from the list `server.trusted_proxies` |
| `HolidayCalendar` | service | Easter (Meeus/Butcher algorithm), the eleven legal holidays of metropolitan France, and the kind of a day |

#### Records of a user

| Entity | Kind | Role |
| --- | --- | --- |
| `DayEntry` | record | Codes of the two half-days of a day and its comment |
| `CounterSettings` | record | Settings of one counter for one user |
| `UserSettings` | record | Tracking start, counter settings, colours |
| `Adjustment` | record | Adjustment of a counter |
| `EntryKind` | value | Kind of a ledger entry |
| `LedgerEntry` | record | Frozen entry of the ledger |
| `MonthStatus` | value | Status of a month |
| `MonthEvent` | record | Event of the history of a month |
| `MonthRecord` | record | Status and history of a month |
| `MonthlyIndicators` | record | Snapshot of the indicators of a month |
| `Calendar` | record | All the data of a user: revision, settings, days, adjustments, months, ledger. Lookups: `IsValidated`, `LastValidated`, `FirstOpen`, `UsesActivity`, `UsesCounter`. Its file holds `ErrStaleRevision` |

#### Bookkeeping

| Entity | Kind | Role |
| --- | --- | --- |
| `Bookkeeper` | service | Rule engine of a user (methods below) |
| `JournalLine` | record | A ledger entry marked provisional or reversed |
| `Journal` | value | Lines indexed by counter; sums and balances by counter, date, month and kind |
| `MonthClosing` | service | Validation conditions and freezing of a month; reopening of the last validated month by reversal entries |
| `IdentifierGenerator` | service | Random identifiers of ledger entries |
| `CounterStatement` | record | Statement of a counter for a month: opening balance, active lines with running balance, cancelled lines, closing balance |
| `StatementLine` | record | Line of a statement with its running balance |
| `Position` | record | The three balances of a counter: balance today, validated balance, forecast at the end of the period |
| `RowState` | value | State of a row of the summary |
| `SummaryRow` | record | One row of the summary: carry-forward row, month row or totals |
| `Summary` | record | Summary of a range of months (the year or one month): carry-forward row, one row per month, totals |
| `Anomaly` | record | Invalid entry of a month, blocking or not |

| Method of `Bookkeeper` | Role |
| --- | --- |
| `Enabled`, `Amount` | Effective enabled flag and effective amount of a counter for the user |
| `Horizon` | Computation limit of the journal; it is not the planning horizon |
| `Journal(through)` | The frozen ledger, then provisional entries month by month, in this order: initial balances, zeroings, debits, yearly credits, credits of feeding codes, monthly credits, roundings, adjustments |
| `DayType` | Display type of a half-day |
| `Indicators` | Monthly indicators |
| `Anomalies` | Invalid entries of a month |
| `Choices` | Codes offered in a select, given the code already entered |
| `CheckEntry` | Check of one day entry |

#### Time

| Entity | Kind | Role |
| --- | --- | --- |
| `Clock` | interface | Time source |
| `SystemClock` | service | Clock of the system |
| `ManualClock` | service | Clock set by hand; it drives the tests |

### internal/account

| Entity | Kind | Role |
| --- | --- | --- |
| `Login` | value | Identifier of an account, `[a-z0-9][a-z0-9._-]{1,31}` |
| `Profile` | value | `GUEST < USER < ADMIN`; only `USER` and `ADMIN` are persisted |
| `State` | value | `enabled` or `disabled`; a disabled account cannot log in and keeps its data |
| `PasswordPolicy` | service | At least 12 characters, 6 distinct, no surrounding spaces, no control or format character, 256 bytes at most, not the login whatever the case; the rules are owned by [SECURITY.md, Passwords](SECURITY.md#passwords) |
| `PasswordHash` | value | `pbkdf2-sha256$<iterations>$<salt>$<key>` (PBKDF2, Password-Based Key Derivation Function 2), verified in constant time |
| `Account` | record | Login, optional name of 60 characters at most, profile, state, hash, `must_change_password`, creation, last login. Its file holds the account errors `ErrUnknownAccount`, `ErrDuplicateLogin`, `ErrLastAdmin`, `ErrTooManyAccounts` and `ErrInvalidName` |
| `SessionRecord` | record | Saved form of a session: SHA-256 of the identifier, login, CSRF (cross-site request forgery) token, creation, last activity |

### internal/store

| Entity | Kind | Role |
| --- | --- | --- |
| `SecureFile` | service | Private file with hardened reading, atomic writes and an exclusive lock |
| `DatabaseRecord` | record | JSON schema of the file: version, revision, configuration, accounts |
| `AccountRecord` | record | One account of the file: `account.Account` and `activity.Calendar` flattened |
| `SessionFile` | service | Saved sessions; implements `web.SessionArchive` |
| `Database` | service | Owner of the database; implements `web.Repository` |

#### SecureFile

- Reading is hardened: no symbolic link, a regular file, owner and mode checks.
- The directory is checked when the file is locked or written.
- A write is atomic: `CreateTemp`, `Sync`, `Rename`, then `Sync` of the directory.
- Once the file is renamed, the write has succeeded: a failed `Sync` of the directory is only logged.
- It removes its own orphan temporary files, named `.<name>.<digits>.tmp`.
- The lock is an exclusive `flock` on `.<name>.lock`.
- It works on Unix only (`secure_file_unix.go`); other platforms return `ErrUnsupportedPlatform`.

#### SessionFile

- The file is named after the database: `.<name>.sessions` next to `<name>`, so `.timesheet-database.json.sessions` next to `timesheet-database.json`.
- The lock follows the same pattern, `.<name>.lock`.
- A database name never starts with a dot, so neither file collides with another database.
- It is read strictly and written atomically through a `SecureFile`.
- The content of the file is described in [SECURITY.md, Sessions](SECURITY.md#sessions).

#### Database

Operations:

- `Create` writes a new database.
- `Lock` takes the lock, then removes the leftover temporary files.
- `Load` decodes strictly, checks the version, checks that the record encodes and decodes again, and validates everything.
- Reads return structural copies (`Calendar.Clone`, `Configuration.Clone`) that share nothing with the state.
- The size of each account is capped (value in [SPECIFICATIONS.md, Limits](SPECIFICATIONS.md#limits)).

A write runs these steps under one `sync.RWMutex`:

1. Copy the committed record.
2. Replace in the copy only the changed account or the configuration; a committed record is never modified in place.
3. Encode the copy and check that it decodes again.
4. Write the file atomically through the `SecureFile`.
5. Replace the state in memory.

### internal/web

#### Assembly

| Entity | Kind | Role |
| --- | --- | --- |
| `Repository` | interface | Port to the database, implemented by `store.Database` |
| `SiteDependencies` | record | Dependencies given to the site: repository, session archive, clock, random source, logger, origins, server configuration, trusted proxies, local prefixes, hash iterations, version |
| `Site` | service | Assembly of the routes and of the HTTP chain |

#### HTTP chain

| Entity | Kind | Role |
| --- | --- | --- |
| `AccessLog` | service | Guard that writes the access log |
| `AccessEntry` | value | Entry that carries the login of a request to the access log |
| `TrackingWriter` | service | Response writer that tracks the status and the size of an answer |
| `SecurityHeaders` | service | Guard that sets the security headers |
| `RecoveryHandler` | service | Guard that turns a panic into a 500 page |
| `HostGuard` | service | Guard that checks the `Host` header |
| `RequestLimits` | service | Guard that caps the URL and the body |
| `OriginGuard` | service | Guard that checks the origin of an unsafe request |
| `Origins` | value | Accepted origins and hosts |
| `ClientAddress` | service | Address of the client behind the trusted proxies |
| `LoginThrottle` | service | Login throttling per address |

#### Routing

| Entity | Kind | Role |
| --- | --- | --- |
| `Router` | service | Path patterns of the `ServeMux` with a method dispatcher: 405 with `Allow`, HEAD served as GET, 404 fallback rendered with the viewer of the session |
| `Route` | record | Method and pattern of a route; the list of routes is compared with the specifications by a test |

#### Sessions and access

| Entity | Kind | Role |
| --- | --- | --- |
| `Session` | value | Session in memory: login, CSRF token, creation, last activity |
| `Sessions` | service | Sessions keyed by the SHA-256 of their identifier, with a CSRF token per session, kept in memory and saved through a `SessionArchive` so that they survive a restart |
| `SessionArchive` | interface | Port to the session file, implemented by `store.SessionFile` |
| `Viewer` | record | Connected account of a request, attached to the request context |
| `AuthGuard` | service | Guard by minimal profile, redirection while the password must change, CSRF check of every POST |

#### Rendering

| Entity | Kind | Role |
| --- | --- | --- |
| `Renderer` | service | Parses each view with `layout.html` and the partial of its page family, and renders a page |
| `PageModel` | record | What a view receives: chrome, section, notice, alerts, CSRF token, data |
| `Chrome` | record | Chrome of the page: viewer, initials, CSRF token of the menu forms, current path, version, year |
| `ErrorPages` | service | Error pages (404, 403, other statuses) |
| `Notices` | service | Fixed table from a notice code to its message |
| `EmbeddedAssets` | service | Embedded templates and static files |
| `AssetServer` | service | Serves the static files |

#### Request data

| Entity | Kind | Role |
| --- | --- | --- |
| `Today` | service | Today in the configured time zone |
| `Workspace` | record | Data of one account loaded for a request, own or read-only, and the paths of its pages |
| `WorkspaceLoader` | service | Loads a workspace, a month or a year from a request |

#### View builders

| View builder | Page | Records it builds |
| --- | --- | --- |
| `DashboardView` | Dashboard | Balances (through `PositionView`) |
| `MonthView` | Activity page (« Activité ») and month page (« Mois ») | Activity page: grid, validation, adjustments. Month page: totals, calendar, summary, statements |
| `YearView` | Year page (« Année ») | Calendar, summary, counters and totals |
| `CountersView` | Year page | Recap and ledger of each counter (through `StatementView`) |
| `StatementView` | Month and year pages | Statement of a counter |
| `PositionView` | Dashboard | The three balances of a counter |
| `SummaryView` | Month and year pages | Summary table |
| `LegendView` | Activity, month and year pages | Legend of the codes |
| `OptionRecord` | Activity page | Option of a select: value, label, selected flag |

#### Controllers

| Controller | Routes |
| --- | --- |
| `HomeController` | `/` |
| `LoginController` | `/login`, `/logout` |
| `DashboardController` | `/dashboard/`, `/dashboard/activity…`, `/dashboard/month…`, `/dashboard/year/…`, `/dashboard/adjustments`, `/dashboard/colors.css`, and the read-only pages `/admin/accounts/{login}/dashboard`, `…/month/…`, `…/year/…` |
| `SettingsController` | `/dashboard/settings` |
| `ProfileController` | `/dashboard/profile` (name, password, sessions) |
| `AdminServerController` | `/admin/` |
| `AdminController` | `/admin/accounts`, `/admin/accounts/{login}` |
| `AdminSettingsController` | `/admin/settings` |

### internal/app

| Entity | Kind | Role |
| --- | --- | --- |
| `Settings` | value | Parsed command line: command and subcommand, positional arguments, `--name=value` options checked per command |
| `OptionList` | value | Option given several times, as a `flag.Value` |
| `Usage` | service | Text of `timesheet help` |
| `UsageError` | value | Error of usage of the command line |
| `Application` | service | Dispatch of the commands and lifecycle of `run` (see [Lifecycle](#lifecycle)) |
| `ServerFile` | service | Server configuration file: strict bounded reading with defaults when it is missing, validation, atomic rewriting by `config set` |
| `InitCommand` | service | `init` |
| `ConfigCommand` | service | `config show`, `config set` |
| `UserCommand` | service | `user add`, `user profile`, `user passwd`, `user del`, `user list` |
| `SessionCommand` | service | `session clear` |
| `ImportCommand` | service | `import` |
| `PasswordReader` | service | Reads a password from `TIMESHEET_PASSWORD`, from the terminal without echo, or from a line of the standard input |
| `TLSProvider` | service | Provided TLS certificate pair |
| `BuildInfo` | value | Version of the release (`ReleaseVersion`) with the commit, its date and its state, from the build information |
| `ServerErrorLog` | service | Sends the errors of `http.Server` to the logger |

`PasswordReader` reads the terminal through termios and `syscall` in `password_reader_linux.go`; `password_reader_other.go` covers the other platforms.

#### Import

| Entity | Kind | Role |
| --- | --- | --- |
| `ImportFile` | record | The JSON import file |
| `ImportSettings` | record | Settings of the import file |
| `ImportCounter` | record | Settings of one counter in the import file |
| `ImportAdjustment` | record | Adjustment in the import file |

`ImportCommand`:

1. decodes the file strictly, within a size bound;
2. applies the settings like the settings page (« Réglages »);
3. checks every day with `Bookkeeper.CheckEntry`;
4. checks the adjustments like the activity page;
5. prints a report;
6. validates the imported months when asked.

The format is specified in [SPECIFICATIONS.md, Import](SPECIFICATIONS.md#import).

## PAGES AND THE CODE THAT SERVES THEM

Every view template is rendered inside `layout.html` and the header partial of its page family (`partials/public.html`, `dashboard.html` or `admin.html`).
Every view of `templates/*.html` other than `layout.html` is parsed at start and rendered by its file name. `Renderer.familiesOf` gives its families: `admin_*` the admin family only; `home`, `login`, `logout` and `logged_out` the public family (a fixed list that a new public page must join); the error pages all three; any other view the admin family on an `/admin/` path for an administrator, the dashboard family otherwise.
The routes and their answers are specified in [SPECIFICATIONS.md, Routes](SPECIFICATIONS.md#routes).

| Page | Route | Controller | View builder | Template | Script |
| --- | --- | --- | --- | --- | --- |
| Home (« Accueil ») | `/` | `HomeController` | none | `home.html` | `chrome_page.js` |
| Login (« Connexion ») | `/login` | `LoginController` | none | `login.html` | `chrome_page.js` |
| Logout (« Déconnexion », « Déconnecté ») | `/logout` | `LoginController` | none | `logout.html`, `logged_out.html` | `chrome_page.js` |
| Dashboard (« Tableau de bord ») | `/dashboard/` | `DashboardController` | `DashboardView` | `dashboard.html` | `chrome_page.js` |
| Activity page (« Activité ») | `/dashboard/activity/{year}/{month}` and its POST routes `…/validate`, `…/reopen`, `/dashboard/adjustments` | `DashboardController` | `MonthView` | `activity.html` | `chrome_page.js`, `month_sheet_page.js` |
| Month page (« Mois ») | `/dashboard/month/{year}/{month}` | `DashboardController` | `MonthView` | `month.html` | `chrome_page.js` |
| Year page (« Année ») | `/dashboard/year/{year}` | `DashboardController` | `YearView` | `year.html` | `chrome_page.js` |
| Settings page (« Réglages ») | `/dashboard/settings` | `SettingsController` | none (the controller builds its record) | `settings.html` | `chrome_page.js` |
| Profile page (« Mon profil ») | `/dashboard/profile` | `ProfileController` | none (the controller builds its record) | `profile.html` | `chrome_page.js` |
| Colours of the codes | `/dashboard/colors.css` | `DashboardController` | none | none (generated style sheet) | none |
| Server page (« Serveur ») | `/admin/` | `AdminServerController` | none (the controller builds its record from the server configuration read at start) | `admin_server.html` | `chrome_page.js` |
| Accounts (« Comptes ») | `/admin/accounts` | `AdminController` | none (the controller builds its records) | `admin_accounts.html` | `chrome_page.js` |
| Account (« Compte … ») | `/admin/accounts/{login}` | `AdminController` | none (the controller builds its record) | `admin_account.html` | `chrome_page.js` |
| Read-only pages of an account | `/admin/accounts/{login}/dashboard`, `…/month/{year}/{month}`, `…/year/{year}` | `DashboardController` | `DashboardView`, `MonthView`, `YearView` | `dashboard.html`, `month.html`, `year.html` | `chrome_page.js` |
| Configuration page (« Configuration ») | `/admin/settings` | `AdminSettingsController` | none (the controller builds its record) | `admin_settings.html` | `chrome_page.js` |
| Errors (« Page introuvable », « Accès refusé », other statuses) | any | `ErrorPages` | none | `not_found.html`, `forbidden.html`, `error.html` | `chrome_page.js` |
| Static files | `/static/{path...}` | `AssetServer` | none | none | none |

`/dashboard/activity` and `/dashboard/month` redirect to the page of the current month.

## FRONT END

The visible behaviour is specified in [SPECIFICATIONS.md, Page structure](SPECIFICATIONS.md#page-structure).
The rules for writing templates, scripts and style sheets are in [CODING.md, Templates](CODING.md#templates), [JavaScript](CODING.md#javascript) and [Styles](CODING.md#styles).

### Pages

- Pages are rendered on the server.
- Each page has a dark header bar, a burger menu on narrow windows, a user menu in `<details>`, and a footer.
- No form depends on JavaScript.

### Scripts

Two ES modules are loaded with `<script type="module">`:

| Script | Class | Loaded on | Behaviour |
| --- | --- | --- | --- |
| `chrome_page.js` | `ChromePage` | Every page | Shows the theme button; applies and switches the light or dark theme, kept in `localStorage` only; closes the burger menu, the user menu and the deletion box of `/admin/accounts` on a click outside or on `Échap`; gives the focus to the confirmation field of that box |
| `month_sheet_page.js` | `MonthSheetPage` | Activity page (editable grid) | Reads the code-to-class table of the `<script type="application/json">` block; copies the « Journée » choice into the selects of the half-days; empties « Journée » when a half-day is changed by hand; repaints the cells |

### Style sheets

| File | Content |
| --- | --- |
| `static/css/theme.css` | Design tokens of the light and dark themes, base elements, page layout |
| `static/css/components.css` | Components |
| `/dashboard/colors.css` | Generated per user: the colours of the activity codes |

### Theme

1. The layout writes no theme.
2. `chrome_page.js` sets `data-theme` on `<html>` from the choice kept in `localStorage`.
3. For the automatic theme it sets nothing, and the browser follows `prefers-color-scheme`.

## HTTP CHAIN

```text
AccessLog → SecurityHeaders → RecoveryHandler → HostGuard → RequestLimits → OriginGuard → Router → AuthGuard → controllers
```

Each guard wraps the next one and may answer instead of it.

| Order | Entity | Role | Refusal |
| --- | --- | --- | --- |
| 1 | `AccessLog` | Writes one line per request once it is answered | none |
| 2 | `SecurityHeaders` | Sets the security headers on every response | none |
| 3 | `RecoveryHandler` | Turns a panic into a 500 page when nothing was written yet | 500 |
| 4 | `HostGuard` | Refuses a `Host` that is not an accepted origin | 421 |
| 5 | `RequestLimits` | Caps the URL and the body (`server.max_url_bytes` and `server.max_body_bytes`, 2048 bytes and 256 KiB by default) | 414 for the URL; a body over the cap makes the reading of the form fail later (400) |
| 6 | `OriginGuard` | Refuses an unsafe request (any method other than GET, HEAD, OPTIONS, TRACE) without an accepted `Origin`, or with a `Sec-Fetch-Site` other than absent, `same-origin` or `none` | 403 |
| 7 | `Router` | Dispatches by path and method | 404, 405 |
| 8 | `AuthGuard` | Resolves the session, reloads the account, enforces the minimal profile, redirects to « Mon profil » while the password must change, checks the CSRF token of every POST | redirection to `/login` without session, 403, 400 when the form cannot be read |
| 9 | controllers | Answer the request | see [SPECIFICATIONS.md, Routes](SPECIFICATIONS.md#routes) |

The statuses and the messages of every refusal are owned by [SPECIFICATIONS.md, Routes](SPECIFICATIONS.md#routes).
Requests refused by the HTTP parser of `net/http` itself (headers beyond the limit, 431; malformed request, 400) never reach the chain: they get a plain-text answer, without security headers and without a line in the access log.

### Access log

- `AccessLog` writes one line per request that reaches the handlers, once it is answered, whatever the guard or the controller that answered.
- Its fields are the client and peer addresses, the login of the session, the method, the host, the path, the protocol, the status, the size, the duration and the user agent.
- It hands the request an `AccessEntry`; `Viewer.Attach` and a successful login fill it with the login.
- A request that never reaches `Viewer.Attach` is logged with the user `-`.
- The fields and the events of the logs are owned by [SPECIFICATIONS.md, Logs](SPECIFICATIONS.md#logs).

### Security headers

`SecurityHeaders` sets, on every response that reaches the chain:

- HSTS (HTTP Strict Transport Security), with the `max-age` of `server.hsts_max_age`, when the origins are `https`;
- the CSP (Content Security Policy):
  - `default-src 'none'`;
  - scripts, styles, images, connections and fonts from `'self'`;
  - `form-action 'self'`, `base-uri 'none'`, `frame-ancestors 'none'`;
  - Trusted Types;
- `X-Content-Type-Options: nosniff` and `X-Frame-Options: DENY`;
- the referrer, cross-origin and permissions policies;
- headers that forbid caching.

The reasons are given in [SECURITY.md, Threat model](SECURITY.md#threat-model).

## MAIN FLOWS

### Login

1. GET `/login` sets a short-lived `ts_login` cookie that holds the token of the form.
2. POST `/login` checks that token against the form.
3. It throttles the attempts per client address (`LoginThrottle`).
4. It verifies the PBKDF2 hash of the password.
5. When the login is unknown, it verifies a decoy hash instead, so that the timing does not reveal whether the login exists.
6. It opens a session, which replaces the previous session of the browser.
7. It redirects to `/dashboard/`.

### Month save

1. The controller loads the calendar of the account, with its revision.
2. It rebuilds the day entries of the month from the form.
3. It checks each entry with `Bookkeeper.CheckEntry`.
4. It saves the whole calendar, giving the store the revision it loaded.
5. The revision read from the form must equal that revision, so that a save never replaces a newer calendar with a stale copy.

The configuration page and the settings page follow the same revision check.

### Validation and reopening

1. `MonthClosing` produces a new `Calendar`: with frozen entries for a validation, with reversals for a reopening.
2. The controller saves it like a month save, with the same revision check.

### Page display

1. `WorkspaceLoader` loads the account and the configuration.
2. `Bookkeeper.Journal` computes the journal up to the computation limit (`Horizon`): the shown month, the end of its year and the end of the running periods.
3. The view builders turn the journal into records.
4. The HTML templates render the records.

## CONCURRENCY

| Lock | Protects | Held during |
| --- | --- | --- |
| `sync.RWMutex` of `Database` | The record in memory | A read copy, or a whole write including the file |
| file lock `.<name>.lock` | The database file against a second process | The life of `run`, or of a writing command |
| leaf lock of `Sessions` | The sessions in memory | One operation on the sessions |
| leaf lock of `LoginThrottle` | The counters of failed logins | One operation on the throttle |
| leaf lock of `ManualClock` | The time of the test clock | One read or change |
| leaf lock of `Application` | The listening address read by `Listen` | One read or change |

- `Database` reads copy the data.
- `Database` writes build a new record that replaces only what changes, write the file, then replace the state.
- Nothing is logged while the lock of `Database` is held, except the warning of a directory that could not be synchronised after a write.
- Concurrency between browsers is optimistic: the configuration and each account carry a revision, and a stale form is refused, never merged.
- One process writes the database at a time: the file lock is taken by `run` and by every command that writes, `import --dry-run` included.

## PERSISTENCE

The file format, its mode, its directory rules, the atomic write, the lock and the version are owned by [SPECIFICATIONS.md, Database](SPECIFICATIONS.md#database).

- `store.SecureFile` checks the file and its directory and writes atomically.
- `store.Database` validates the whole file at load and refuses a newer schema version.
- `store.SessionFile` writes the sessions through its own `SecureFile`.

## CONFIGURATION AND SERVER CONFIGURATION

The word "configuration" names the configuration section of the database; the server configuration is a separate file.

### Configuration and its template

1. `init` initialises the database from the configuration template `share/timesheet/timesheet-database-template.json`.
2. Afterwards, the activity codes, the counters and their colours live in the database.
3. An administrator edits them on the configuration page, `/admin/settings`.

The content of the template is specified in [SPECIFICATIONS.md, Configuration template](SPECIFICATIONS.md#configuration-template).

### Server configuration

1. The server configuration lives in `etc/timesheet/timesheet.json` from the sources, `/etc/timesheet/timesheet.json` in the package; `--config` gives its path, `DefaultConfig` of `app` (relative to the current directory) by default.
2. `Application` reads it with `ServerFile` before any command but `help`, `version` and the `config` commands, which read it themselves.
3. `Settings.ResolveDatabase` takes the database path from `database.path` unless `--database` is given, and makes it absolute.
4. `Application` hands the configuration to the commands, which open the database with `store.NewDatabase(path, logger, database.max_bytes, limits)` and hash with `passwords.iterations`.
5. `run` builds the `http.Server` from the section `server`, and gives the whole configuration to the web layer through `SiteDependencies.Server`: `NewSessions` takes the section `sessions`, `NewLoginThrottle` the section `login`, `NewRequestLimits` and `NewSecurityHeaders` the sizes and HSTS, the login controller the life of its form, the configuration page every value.
6. The database never holds it.

Its settings are specified in [SPECIFICATIONS.md, Server configuration](SPECIFICATIONS.md#server-configuration).

## LIFECYCLE

### main

1. `main` creates the signal context.
2. It builds the `Application` with `os.Args`, the environment and the standard streams.
3. It also passes the local time zone, which is not used: today follows `server.timezone`.
4. The rules of `main` are in [CODING.md, Object-oriented style](CODING.md#object-oriented-style).

### run

1. Parse the options (`Settings`).
2. Create the logger.
3. Read the server configuration file, resolve the database path, and create the logger again at `server.log_level` when `--log-level` is not given.
4. Lock, then load the database.
5. Check that an administrator account exists.
6. Open the listener.
7. Compute the accepted origins.
8. Load the TLS pair of the options or of the file, when one is given, after the check that the origins are https.
9. Open the session file: check its directory, remove the leftover temporary files.
10. Build the site with `web.NewSite`, which restores the saved sessions.
11. Build the `http.Server` with the timeouts and the header size of the section `server`.
12. Serve until the signal context ends.
13. Shut down gracefully, within `server.shutdown_timeout`.

### Other commands

1. Restore the default signal handling (`WithSignals`), so that one interrupt stops the command.
2. Read the server configuration and resolve the database path, except for `config show` and `config set`, which only read or rewrite the file.
3. Open the database.
4. Lock it when the command writes: `init`, `user add`, `user profile`, `user passwd`, `user del`, `session clear` and `import`, `--dry-run` included. `user list` reads without the lock; `config show` and `config set` do not open the database.
5. Act, then exit.

## GLOSSARY

The terms of the domain (counter, ledger, journal, accrual, tracking start and others) are defined in [SPECIFICATIONS.md, Glossary](SPECIFICATIONS.md#glossary).
The terms below belong to the code.

| Term | Meaning |
| --- | --- |
| entity | A Go type with its own file, named after it |
| record | Data with exported, tagged fields, persisted or rendered |
| guard | A handler of the HTTP chain that wraps the next one and may answer instead of it |
| HTTP chain | The ordered guards in front of the router and the controllers |
| view builder | An entity of `web` that turns a workspace into the records of a template |
| viewer | The connected account of a request: login, name, profile, must-change flag, session key, CSRF token |
| workspace | The data of one account loaded for a request: owner, calendar, configuration, bookkeeper, today, base path, read-only flag |
| chrome | The parts of a page around its content: header bar, menus, footer |
| page family | The header partial a view is rendered with: public, dashboard or admin |
| computation limit | The last month the journal is computed for (`Bookkeeper.Horizon`); not the planning horizon |
| leaf lock | A lock under which no other lock is taken |
| unsafe request | A request whose method is not GET, HEAD, OPTIONS or TRACE |
