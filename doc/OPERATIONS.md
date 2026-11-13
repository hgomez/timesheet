# OPERATIONS

This document describes how to install, configure, expose, monitor, back up, upgrade and repair a Timesheet server.

It is written for the operator, the person who runs the server and has a shell on its host. Building the program and the package is described in [BUILD.md](BUILD.md); the first start from the sources in the [README](../README.md#getting-started); the daily use in the [User guide](USER-GUIDE.md).

## CONTENTS

- [TWO INSTALLATIONS](#two-installations)
- [COMMANDS](#commands)
- [SERVER CONFIGURATION](#server-configuration)
- [NETWORK AND SECURITY](#network-and-security)
- [LOGS](#logs)
- [IMPORTING](#importing)
- [DEBIAN PACKAGE](#debian-package)
- [BACKUP AND RESTORE](#backup-and-restore)
- [RECOVERY](#recovery)
- [UPGRADE](#upgrade)
- [TROUBLESHOOTING](#troubleshooting)

## TWO INSTALLATIONS

Every procedure below gives two forms:

- the repository form: the binary `bin/timesheet`, run from the repository root by your own user;
- the package form: the binary `timesheet` of the [Debian package](#debian-package), run as root or as the account `timesheet`.

| Item | Repository | Package |
| --- | --- | --- |
| Database | `var/lib/timesheet/timesheet-database.json` | `/var/lib/timesheet/timesheet-database.json` |
| Server configuration | `etc/timesheet/timesheet.json` | `/etc/timesheet/timesheet.json` |
| Configuration template | `share/timesheet/timesheet-database-template.json` | `/usr/share/timesheet/timesheet-database-template.json` |
| Import template | `share/timesheet/timesheet-import-template.json` | `/usr/share/timesheet/timesheet-import-template.json` |
| Sessions | `.timesheet-database.json.sessions`, next to the database | the same, in `/var/lib/timesheet/` |
| Server | `bin/timesheet run`, in the foreground | the unit `timesheet.service` |
| Logs | the standard error of `run` | the journal: `journalctl -u timesheet` |
| Commands that use the database | `bin/timesheet <command>` | `runuser -u timesheet -- timesheet <command> --database=/var/lib/timesheet/timesheet-database.json --config=/etc/timesheet/timesheet.json` |
| `config set` | `bin/timesheet config set …` | `timesheet config set … --config=/etc/timesheet/timesheet.json`, as root |

The default paths are relative to the current directory, for the installed binary too. Outside the repository root, always give `--database`, `--config` and, for `init`, `--template`.

## COMMANDS

| Command | Purpose | Server stopped |
| --- | --- | --- |
| `init` | Creates the database from the configuration template. | Not applicable |
| `run` | Starts the web server. | Not applicable |
| `config show` | Prints the server configuration. | No |
| `config set KEY VALUE` | Changes one setting of the server configuration file. | No |
| `user add LOGIN` | Creates an account (`--profile=USER` or `ADMIN`, `USER` by default). | Yes |
| `user profile LOGIN PROFILE` | Changes the profile of an account. | Yes |
| `user passwd LOGIN` | Sets a permanent password and closes the sessions of the account. | Yes |
| `user del LOGIN` | Deletes an account and its data. | Yes |
| `user list` | Lists the accounts. | No |
| `session clear` | Closes every session. | Yes |
| `import LOGIN FILE` | Imports a JSON file, see [Importing](#importing). | Yes, even with `--dry-run` |
| `help`, `version` | Prints the commands and their options; prints the release number and the commit it was built from. | No |

- The exact syntax, the options, the exit codes and the messages are in [Command line](SPECIFICATIONS.md#command-line).
- Every command accepts `--database=FILE`, `--config=FILE` and `--log-level=LEVEL` (`debug`, `info`, `warn` or `error`).
- Options are written `--name=value`, before or after the arguments. An argument cannot start with a dash.
- A command that needs the server stopped refuses to run while the server holds the database. Accounts can also be managed in « Administration » while it runs.
- The sessions survive a restart of the server: they are saved, hashed, in `.<name>.sessions` next to a database `<name>`.

## SERVER CONFIGURATION

The server configuration file holds four settings, apart from the database:

| Setting | Meaning | Default |
| --- | --- | --- |
| `server.addr` | The listening address, `host:port`. `0.0.0.0:8080` listens on every interface. | `127.0.0.1:8080` |
| `server.origins` | The addresses under which the site is opened, all `http` or all `https`. Any other host name gets « Hôte inconnu » (421). | `http://127.0.0.1:8080`, `http://localhost:8080` |
| `server.trusted_proxies` | The IP addresses or prefixes of the reverse proxies whose `X-Forwarded-For` header is believed. An empty list trusts none. | `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16` |
| `server.timezone` | The IANA time zone (a name such as `Europe/Paris`) that decides what « today » is. | `Europe/Paris` |

The format, the defaults and the validation rules are in [Server configuration](SPECIFICATIONS.md#server-configuration).

To change a setting:

1. Run `config set`. A list is written with commas.

   Repository:

   ```sh
   bin/timesheet config set server.addr 0.0.0.0:8080
   ```

   Package, as root:

   ```sh
   timesheet config set server.addr 0.0.0.0:8080 --config=/etc/timesheet/timesheet.json
   ```

   It answers `server.addr mis à jour dans <file> ; relancez le serveur pour l'appliquer.` and refuses an invalid value. You can also edit the file by hand (as root on a package).

2. When you change `server.addr`, change `server.origins` too: a request whose host is not an accepted origin is refused.

3. Restart the server: a change applies at the next start. On a package: `systemctl restart timesheet`.

Rules to know:

- Without `--config`, `config set` and `config show` act on `etc/timesheet/timesheet.json` under the current directory. `config set` creates that file when it is missing. On a package, always give `--config=/etc/timesheet/timesheet.json`.
- `config show` prints the settings of the file, validated and re-encoded (an absent list is printed `[]`), or the defaults when the file is missing, without saying so. It fails with the faulty setting when the file is invalid. It never shows the options given to a running server.
- The options of `run` take precedence over the file: `--addr` over `server.addr`, `--origin` over `server.origins`, `--trusted-proxy` over `server.trusted_proxies`.
- With `--addr` and without `--origin`, the origins of the file are not used. The server then accepts, on that port, `localhost`, `127.0.0.1`, `::1`, the IPv4 addresses of the interfaces, the host name and `<host name>.local`. Give `--origin` to fix the list.
- The defaults apply only when the whole file is missing. In an existing file, a missing `trusted_proxies` trusts no proxy.

## NETWORK AND SECURITY

The server does not refuse plain HTTP on a network address. With a non-loopback address (`--addr` or `server.addr`) and no TLS, passwords and session cookies travel in clear. A warning is logged whenever the origins are not `https`.

Plain HTTP on the loopback is meant for a host without other users or other local web services.

On a network, use one of:

- a certificate pair: `--tls-cert` and `--tls-key`, with `https` origins;
- a TLS reverse proxy, with `https` origins and the address of the proxy in the trusted proxies.

With a reverse proxy:

- The trusted proxy is an IP address or a prefix, such as `10.0.0.1` or `10.0.0.0/8`. The option `--trusted-proxy` can be repeated.
- The proxy must pass the public host unchanged and append the client address to `X-Forwarded-For`.
- The default configuration trusts the private ranges of RFC 1918 (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`): any machine of these ranges can then choose the client address it claims. Narrow the list to the address of your proxy.

The risks and the protections are in [Transport](SECURITY.md#transport), [Client address and trusted proxies](SECURITY.md#client-address-and-trusted-proxies) and [Accepted risks](SECURITY.md#accepted-risks).

### Behind a reverse proxy on another machine

In this example, the server listens on `192.0.2.10`, the proxy has the address `192.0.2.20` and the public address is `https://timesheet.example.org`.

1. Set the listening address to an address the proxy can reach, not the loopback:

   ```sh
   bin/timesheet config set server.addr 192.0.2.10:8080
   ```

2. Set the origin to the public address, not to the address of the server:

   ```sh
   bin/timesheet config set server.origins https://timesheet.example.org
   ```

3. Trust only the proxy:

   ```sh
   bin/timesheet config set server.trusted_proxies 192.0.2.20
   ```

4. Restart the server.
5. Configure the proxy: TLS towards the clients, the `Host`, `Origin` and `Sec-Fetch-Site` headers passed unchanged, the client address appended to `X-Forwarded-For`, requests sent to `http://192.0.2.10:8080`.
6. Check the log: a request through the proxy shows `client=<address of the client> peer=192.0.2.20`, and a request sent directly to `192.0.2.10:8080` with that address as host gets 421. This does not block a direct request that sends the public host name: only a firewall does (below).

On a package, write the three commands `timesheet config set … --config=/etc/timesheet/timesheet.json` as root, then `systemctl restart timesheet`.

With `https` origins, the session cookies are marked `Secure` and the server sends HSTS (Strict-Transport-Security). Limits of this setup:

- The link between the proxy and the server is plain HTTP: passwords and cookies cross it in clear, and the server logs no warning.
- Any machine that reaches the listening port can talk to the server without the proxy, by sending the public host name.
- Restrict the listening port to the proxy with a firewall, or serve TLS on that link too with `--tls-cert` and `--tls-key`.

## LOGS

- The server logs one line per request (`msg=request`) and one line per event, such as a login or a saved month. The fields are listed in [Logs](SPECIFICATIONS.md#logs).
- The default level is `info`. `--log-level=warn` keeps only the warnings and the errors.
- The commands of the command line log no request and no event: they only print their warnings on the standard error, in the same format.

| | Repository | Package |
| --- | --- | --- |
| Read the logs | the standard error of `bin/timesheet run` | `journalctl -u timesheet` |
| Change the level | `bin/timesheet run --log-level=warn` | a drop-in of the unit, see [Debian package](#debian-package) |
| Retention | what you keep of the standard error | the settings of the journal (`/etc/systemd/journald.conf`) |

What the logs hold about people is listed in [Personal data](SPECIFICATIONS.md#personal-data).

## IMPORTING

Days, counter settings and adjustments can be imported from a JSON file of the format `timesheet-import`, described in [Import](SPECIFICATIONS.md#import).

- The file holds the tracking start, counter settings, days with morning, afternoon and comment, and adjustments.
- The application reads no spreadsheet. A spreadsheet is first converted to this format outside the application, its categories becoming the activity codes.
- `share/timesheet/timesheet-import-template.json` is a small, fictitious file of this format: one month (January 2026) with a holiday, paid leave, a half-day of RTT, an absence, comments, the settings of three counters and one adjustment. Copy it as a starting point.

To import:

1. Create the account: from « Comptes » (provisional password, changed at the first login) or with `user add` (the operator, who runs the command, knows the password until the user changes it).
2. Stop the server.
3. Run a dry run until it reports no error:

   ```sh
   bin/timesheet import <login> ~/private/activity.json --dry-run
   ```

   Try it first on the shipped template, `bin/timesheet import <login> share/timesheet/timesheet-import-template.json --dry-run`, which answers:

   ```text
   Réglages lus : début du suivi 2026-01-01, 3 compteur(s).
   31 journée(s) lue(s), de 2026-01 à 2026-01.
   1 ajustement(s) lu(s).
   Simulation : rien n'a été écrit.
   ```

   - Every day is checked like an entry of the grid, every adjustment like one of « Activité ». Any problem is listed with its date and nothing is written.
   - It lists, month by month, the days before today left without entry, weekends and holidays included.
   - With `--validate`, it also lists the months that would be validated.

4. Import for real, without `--dry-run`. Use `--validate` only if no missing day is listed: each month is then recorded as validated on its last day. Otherwise complete the listed days in the grid and validate month by month.
5. Start the server.

- Other options: `--from=YYYY-MM` and `--to=YYYY-MM` limit the months; `--replace` replaces the days already entered at the dates of the file.
- Gaps with the payslip or hand corrections of a balance are adjustments, in the file or in « Activité ».
- The spreadsheet and the file made from it hold personal data. Keep them out of the repository, and give the file by a path outside it: a file inside the repository may be committed by mistake.
- On a package, the command runs as the account `timesheet` (`runuser -u timesheet -- timesheet import … --config=/etc/timesheet/timesheet.json --database=/var/lib/timesheet/timesheet-database.json`): the file must be readable by that account.

## DEBIAN PACKAGE

### What the package installs

The package is built as described in [BUILD.md](BUILD.md#build-the-debian-package), into `_dist/`.

It installs:

- `/usr/bin/timesheet`;
- the configuration template and the import template in `/usr/share/timesheet/`;
- the server configuration `/etc/timesheet/timesheet.json`, a configuration file of the package: an upgrade keeps your changes;
- the unit `timesheet.service`, which runs `timesheet run --config=/etc/timesheet/timesheet.json --database=/var/lib/timesheet/timesheet-database.json` as the account `timesheet`.

Its installation script creates, each when missing:

- the system account `timesheet`;
- the directory `/var/lib/timesheet`, mode 0700, owned by that account;
- the database `/var/lib/timesheet/timesheet-database.json`.

The installation script leaves an existing directory as it is. The unit (`StateDirectory=`) gives `/var/lib/timesheet` to the account `timesheet` at each start of the service; an existing directory keeps its mode (systemd only warns when it differs from 0700).

### Install

1. Install the package, as root:

   ```sh
   apt install ./_dist/timesheet_<version>_amd64.deb
   ```

2. The service starts, and fails: the new database has no administrator. It restarts on failure until systemd gives up.
3. Create the first administrator, with the service stopped:

   ```sh
   systemctl stop timesheet
   runuser -u timesheet -- timesheet user add <login> --profile=ADMIN --database=/var/lib/timesheet/timesheet-database.json --config=/etc/timesheet/timesheet.json
   systemctl start timesheet
   ```

   If `systemctl start` refuses because the service failed too often, run `systemctl reset-failed timesheet` first.

4. Check the journal: `journalctl -u timesheet`. The service listens on `http://127.0.0.1:8080`.
5. To change the address, the origins, the trusted proxies or the time zone, see [Server configuration](#server-configuration).

### Upgrade the package

1. Back up the database (see [Backup and restore](#backup-and-restore)).
2. If the database was created before the server settings moved to `/etc/timesheet/timesheet.json`, stop the service (`systemctl stop timesheet`) and migrate it first, as described at the end of [Upgrade](#upgrade).
3. Install the new package, as root: `apt install ./_dist/timesheet_<version>_amd64.deb`. The package restarts the service.
4. When both you and the new package changed `/etc/timesheet/timesheet.json`, `apt` asks which version to keep.
5. Check the journal: `journalctl -u timesheet`.

### Options of the service

The TLS options and the log level are options of `run`. Give them in a drop-in of the unit:

1. Run `systemctl edit timesheet` and write:

   ```ini
   [Service]
   ExecStart=
   ExecStart=/usr/bin/timesheet run --config=/etc/timesheet/timesheet.json --database=/var/lib/timesheet/timesheet-database.json --tls-cert=/etc/timesheet/cert.pem --tls-key=/etc/timesheet/key.pem --log-level=warn
   ```

   The empty `ExecStart=` clears the command of the unit. Keep only the options you need.

2. Make the certificate pair readable by the account `timesheet`.
3. Run `systemctl restart timesheet`.

### Remove

Removing or purging the package keeps `/var/lib/timesheet` and the account `timesheet`.

## BACKUP AND RESTORE

To back up, copy the database file. Writes are atomic, so a copy is consistent even while the server runs.

| | Repository | Package, as root |
| --- | --- | --- |
| Back up | `cp var/lib/timesheet/timesheet-database.json <backup>` | `cp /var/lib/timesheet/timesheet-database.json <backup>` |

Protect the copy like the original: it contains the password hashes and every user's absences and comments. See [Data protection](SECURITY.md#data-protection).

To restore:

1. Stop the server: it keeps the database in memory and would overwrite the file. On a package: `systemctl stop timesheet`.
2. Copy the file back, owned by the user of the server, without any right for the group or others, in a directory owned by that user and not writable by others (unless it has the sticky bit).

   Repository, as your user:

   ```sh
   install -m 600 <backup> var/lib/timesheet/timesheet-database.json
   ```

   Package, as root:

   ```sh
   install -m 600 -o timesheet -g timesheet <backup> /var/lib/timesheet/timesheet-database.json
   ```

   Adapt the path when you use `--database`.

3. If the sessions saved next to the database must not stay open, close them: `bin/timesheet session clear`, or on a package `runuser -u timesheet -- timesheet session clear --database=/var/lib/timesheet/timesheet-database.json`.
4. Start the server again. On a package: `systemctl start timesheet`.

A restored backup brings back the accounts, passwords, profiles and states of its date. Change again the passwords changed since, and disable or delete again the accounts disabled or deleted since.

## RECOVERY

When the password of an administrator is lost:

1. Stop the server. On a package: `systemctl stop timesheet`.
2. Set a new password:

   Repository:

   ```sh
   bin/timesheet user passwd <login>
   ```

   Package, as root:

   ```sh
   runuser -u timesheet -- timesheet user passwd <login> --database=/var/lib/timesheet/timesheet-database.json --config=/etc/timesheet/timesheet.json
   ```

3. Start the server. On a package: `systemctl start timesheet`.

- The password is permanent and the sessions of the account are closed.
- `user passwd` does not enable a disabled account: enable it from « Comptes » with another administrator.

## UPGRADE

This procedure upgrades the repository form. For the package, see [Upgrade the package](#upgrade-the-package).

1. Back up the database.
2. Stop the server.
3. Get the new sources and rebuild with `make` ([BUILD.md](BUILD.md#build-the-program)): the templates and the assets are embedded in the binary.
4. Apply the migration notes that concern your database: see [Migration notes](SPECIFICATIONS.md#migration-notes). The one known today is below.
5. Start the server again.

A database created before the server settings moved to the server configuration file holds a `server` object in its `configuration`. The new binary refuses it with `store: malformed database: json: unknown field "server"`. To migrate it, with the server stopped and a backup taken:

1. Read the former settings:

   ```sh
   jq '.configuration.server' var/lib/timesheet/timesheet-database.json
   ```

2. Copy `addr`, `origins` and `timezone` into the server configuration with `config set`.
3. Write the former `trusted_proxy` as the list `server.trusted_proxies`: its address, or `config set server.trusted_proxies ''` (written `[]`) when it was `null` (no trusted proxy).
4. Remove the `server` object from the database:

   ```sh
   jq 'del(.configuration.server)' var/lib/timesheet/timesheet-database.json > <new file>
   install -m 600 <new file> var/lib/timesheet/timesheet-database.json
   ```

On a package, use `/var/lib/timesheet/timesheet-database.json`, `--config=/etc/timesheet/timesheet.json` and `install -m 600 -o timesheet -g timesheet`, as root.

Other points:

- The database is checked entirely at load. An older binary refuses a database written by a newer one: keep the backup taken before the upgrade to go back.
- A server configuration file using the former key `trusted_proxy` is refused (`json: unknown field "trusted_proxy"`): rename it `trusted_proxies` and write it as a list.
- Sessions are saved in `.<database file name>.sessions`. A `timesheet-sessions.json` left by an older version is no longer read, so its users log in again. Delete it: it holds the CSRF tokens of its sessions.
- The certificate pair, when used, is read at start: restart the server after renewing it.

## TROUBLESHOOTING

The errors of the commands are prefixed with `timesheet:`. Those of `run` are log lines `level=ERROR msg=…`, except `no administrator account`.

| Message | Cause | What to do |
| --- | --- | --- |
| `open share/timesheet/timesheet-database-template.json: no such file or directory` | `init` is run outside the repository root, or by the installed binary. | Go to the root, or give `--template` and `--database`. |
| `cannot lock …` with `the directory does not exist` | `run` is started outside the repository root, or with a wrong `--database`. | Go to the root, run `bin/timesheet.sh`, or give `--database`. |
| `the database does not exist, run init first` | The database file is missing, at the default path or at `--database` (for `user list`, also when its directory is missing). | Run `init`, or go to the repository root. |
| `the database already exists` | `init` is run a second time. | Nothing: the database is ready. For another one, give `--database`. |
| `insecure directory: … is writable by other` | The directory of the database is open to other users. | `chmod 700` on that directory. |
| `insecure file: … is accessible to group or other (expected 0600)` | The database was copied with a wrong mode, often by a restore. | `chmod 600` on the file. |
| `… belongs to another user` | The command runs as another user than the owner of the file, such as root on a package. | Run it as the owner: `runuser -u timesheet -- …` on a package. |
| `malformed database: json: unknown field "server"` | The database was created before the server settings moved to their own file. | Follow [Upgrade](#upgrade). |
| `the password needs at least 12 characters…` | The password is refused by the policy. | Choose a longer or more varied one. |
| `no administrator account: run 'timesheet user add …' first` | `run` finds no enabled administrator. | Create one with `user add <login> --profile=ADMIN`. |
| `cannot lock …` with `the file is used by another process` | Another instance already serves this database. | Use that instance, or stop it. |
| `cannot listen on 127.0.0.1:8080` with `address already in use` | Another program, or an instance serving another database, uses the port. | Stop it, or start with `--addr=127.0.0.1:8081`: the accepted origins then follow that address. |
| `the file is used by another process … (stop the server or use /admin/)`, or `(stop the server first)` | A command that needs the server stopped is run while the server holds the database. | Stop the server first, or act from « Administration ». |
| « Hôte inconnu » (421) | The page is opened under a host name that is not an accepted origin. | Use `127.0.0.1` or `localhost`, or declare the name in `server.origins` (`config set server.origins`) or with `--origin`. |
| « Identifiant ou mot de passe incorrect. » | Wrong login or password, or disabled account. | Try again. Each failure makes the address wait, from 1 s up to 5 min (« Trop de tentatives », 429). |
| « Le formulaire a expiré ou n'est pas valide. Rechargez la page et recommencez. » (403) | The page was open too long, or the session was replaced. | Reload the page and submit again. |
| « Les données ont changé entre-temps… » (409) | The same account was changed in another tab. | Reload the page and enter the change again. |

The limits behind other refusals (size of a request, of an account, number of accounts) are in [Limits](SPECIFICATIONS.md#limits).
