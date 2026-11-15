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
- [EXAMPLE DATABASE](#example-database)
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
| Database | `var/lib/timesheet/timesheet-database.json`, given with `--database` | `/var/lib/timesheet/timesheet-database.json`, the `database.path` of the server configuration |
| Server configuration | `etc/timesheet/timesheet.json` | `/etc/timesheet/timesheet.json` |
| Configuration template | `share/timesheet/database-template.json` | `/usr/share/timesheet/database-template.json` |
| Example database | `share/timesheet/database-example.json` | `/usr/share/timesheet/database-example.json` |
| Import template | `share/timesheet/import-template.json` | `/usr/share/timesheet/import-template.json` |
| Sessions | `.timesheet-database.json.sessions`, next to the database | the same, in `/var/lib/timesheet/` |
| Server | `bin/timesheet.sh`, or `bin/timesheet run --database=var/lib/timesheet/timesheet-database.json`, in the foreground | the unit `timesheet.service` |
| Logs | the standard error of `run` | the journal: `journalctl -u timesheet` |
| Commands that use the database | `bin/timesheet <command> --database=var/lib/timesheet/timesheet-database.json` | `runuser -u timesheet -- timesheet <command> --config=/etc/timesheet/timesheet.json` |
| `config set` | `bin/timesheet config set …` | `timesheet config set … --config=/etc/timesheet/timesheet.json`, as root |

- The database comes from `database.path` of the server configuration, `/var/lib/timesheet/timesheet-database.json` by default: the database of the package.
- Every binary reads `etc/timesheet/timesheet.json` under the current directory by default, the binary of the package included: on a package, always give `--config=/etc/timesheet/timesheet.json`. The database then comes from that file, so the commands of the package need no `--database`.
- From the sources, run the binary from the repository root, and give `--database=var/lib/timesheet/timesheet-database.json` to every command that uses the database of the repository: without it, the command opens `/var/lib/timesheet/timesheet-database.json`. `bin/timesheet.sh` passes both paths to `run`.

## COMMANDS

| Command | Purpose | Server stopped |
| --- | --- | --- |
| `init` | Creates the database from the configuration template. | Not applicable |
| `run` | Starts the web server. | Not applicable |
| `config show [KEY]` | Prints the server configuration, or the raw value of one key. | No |
| `config set KEY VALUE...` | Changes one or several settings of the server configuration file. | No |
| `user add LOGIN` | Creates an account (`--profile=USER` or `ADMIN`, `USER` by default). | Yes |
| `user profile LOGIN PROFILE` | Changes the profile of an account. | Yes |
| `user passwd LOGIN` | Sets a permanent password and closes the sessions of the account. | Yes |
| `user del LOGIN` | Deletes an account and its data. | Yes |
| `user list` | Lists the accounts. | No |
| `session clear` | Closes every session. | Yes |
| `import LOGIN FILE` | Imports a JSON file, see [Importing](#importing). | Yes, even with `--dry-run` |
| `help`, `version` | Prints the commands and their options; prints the release number and the commit it was built from. | No |

- The exact syntax, the options, the exit codes and the messages are in [Command line](SPECIFICATIONS.md#command-line).
- Every command accepts `--database=FILE`, `--config=FILE` and `--log-level=LEVEL` (`debug`, `info`, `warn` or `error`). `--database` and `--log-level` take precedence over `database.path` and `server.log_level`.
- Options are written `--name=value`, before or after the arguments. An argument cannot start with a dash.
- A command that needs the server stopped refuses to run while the server holds the database. Accounts can also be managed in « Administration » while it runs.
- The sessions survive a restart of the server: they are saved, hashed, in `.<name>.sessions` next to a database `<name>`.

## SERVER CONFIGURATION

The server configuration file holds the settings of the server and of its operation, apart from the database. Every key, its bounds and its default are in [Server configuration](SPECIFICATIONS.md#server-configuration); the main ones:

| Setting | Meaning | Default |
| --- | --- | --- |
| `server.listen` | The listening address, `host:port`. `0.0.0.0:8080` listens on every interface. | `127.0.0.1:8080` |
| `server.origins` | The addresses under which the site is opened, all `http` or all `https`. Any other host name gets « Hôte inconnu » (421). | `http://127.0.0.1:8080`, `http://localhost:8080` |
| `server.trusted_proxies` | The IP addresses or prefixes of the reverse proxies whose `X-Forwarded-For` header is believed. An empty list trusts none. | `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16` |
| `server.timezone` | The IANA time zone (a name such as `Europe/Paris`) that decides what « today » is. | `Europe/Paris` |
| `server.log_level` | The lowest level of the logs of the server. | `info` |
| `server.tls_cert`, `server.tls_key` | A certificate pair to serve HTTPS, or `null`. | `null` |
| `server.*_timeout`, `server.max_*_bytes` | The timeouts and the request sizes of the server. | 5 s, 15 s, 30 s, 1 min, 10 s; 8 KiB, 256 KiB, 2048 bytes |
| `database.path` | The database file. | `/var/lib/timesheet/timesheet-database.json` |
| `database.max_bytes` | The largest database file. | 64 MiB |
| `sessions.*` | The expiry of the sessions and their number per account. | 8 h without activity, 24 h at most, 64 |
| `login.*` | The life of the login form and the login throttle. | 1 h; 1 s doubling up to 5 min |
| `passwords.iterations` | The cost of a new password hash. | 600 000 |
| `limits.*` | The accounts, and the size and adjustments of one account. | 200; 1 MiB, 1000 |

To read a setting:

```sh
bin/timesheet config show                   # the whole file, every key included
bin/timesheet config show database.path     # one raw value, for a script
```

To change settings:

1. Run `config set` with one or several pairs `KEY VALUE`. A list is written with commas; a duration as `500ms`, `30s`, `5m` or `8h`; a size in bytes.

   Repository:

   ```sh
   bin/timesheet config set server.listen 0.0.0.0:8080 server.origins http://timesheet.example:8080
   ```

   Package, as root:

   ```sh
   timesheet config set server.listen 0.0.0.0:8080 server.origins http://timesheet.example:8080 --config=/etc/timesheet/timesheet.json
   ```

   It answers `server.listen, server.origins updated in <file>; restart the server to apply it.`, and refuses an invalid or out-of-bounds value without writing anything. You can also edit the file by hand (as root on a package); the next `config set` rewrites it in its layout, one entry per line.

2. Change linked settings in the same command: `server.listen` with `server.origins`, `server.tls_cert` with `server.tls_key`, `sessions.idle_timeout` with `sessions.max_lifetime`, `login.initial_delay` or `login.forget_after` with `login.max_delay`, `limits.accounts` with `sessions.per_account`. One after the other, the first may be refused.

3. Restart the server: a change applies at the next start. On a package: `systemctl restart timesheet`. The page « Serveur » of the administration shows the values read at the start.

Common changes:

On a package, add `--config=/etc/timesheet/timesheet.json` to each of these commands, run as root.

| Need | Command |
| --- | --- |
| Listen on every interface | `config set server.listen 0.0.0.0:8080 server.origins http://<host name>:8080` |
| Keep the sessions open a working week | `config set sessions.idle_timeout 12h sessions.max_lifetime 120h` |
| Harden the login throttle | `config set login.initial_delay 2s login.max_delay 15m login.forget_after 2h` |
| Serve HTTPS with a certificate pair | `config set server.tls_cert /etc/timesheet/cert.pem server.tls_key /etc/timesheet/key.pem server.origins https://<host name>:8080` |
| Allow more accounts | `config set limits.accounts 400 sessions.per_account 32` |
| Write more logs | `config set server.log_level debug` |

Rules to know:

- Without `--config`, `config set` and `config show` act on `etc/timesheet/timesheet.json` under the current directory, and `config set` creates that file when it is missing. On a package, always give `--config=/etc/timesheet/timesheet.json`.
- `config show` prints the settings of the file, validated and re-encoded (an absent list is printed `[]`, every absent key with its default), or the defaults when the file is missing, without saying so. It fails with the faulty setting when the file is invalid. It never shows the options given to a running server.
- The options take precedence over the file: `--listen` over `server.listen`, `--origin` over `server.origins`, `--trusted-proxy` over `server.trusted_proxies`, `--tls-cert` and `--tls-key` over the pair of the file, `--database` over `database.path`, `--log-level` over `server.log_level`.
- With `--listen` and without `--origin`, the origins of the file are not used. The server then accepts, on that port, `localhost`, `127.0.0.1`, `::1`, the IPv4 addresses of the interfaces, the host name and `<host name>.local`. Give `--origin` to fix the list.
- In an existing file, `server.listen`, `server.origins` and `server.timezone` are mandatory, and a missing `trusted_proxies` trusts no proxy. Every other missing key takes its default.
- The bounds keep a floor for security (600 000 iterations, 256 KiB of request body for the largest form) and keep the sessions restorable (at most 16 384 sessions in all).

## NETWORK AND SECURITY

The server does not refuse plain HTTP on a network address. With a non-loopback address (`--listen` or `server.listen`) and no TLS, passwords and session cookies travel in clear. A warning is logged whenever the origins are not `https`.

Plain HTTP on the loopback is meant for a host without other users or other local web services.

On a network, use one of:

- a certificate pair: `server.tls_cert` and `server.tls_key` in the server configuration (or `--tls-cert` and `--tls-key`), with `https` origins; the lowest TLS version is `server.tls_min_version`, 1.2 by default;
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
   bin/timesheet config set server.listen 192.0.2.10:8080
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

On a package, run the three commands `timesheet config set … --config=/etc/timesheet/timesheet.json` as root, then `systemctl restart timesheet`. The three settings can also go in one command: `config set server.listen … server.origins … server.trusted_proxies …`.

With `https` origins, the session cookies are marked `Secure` and the server sends HSTS (Strict-Transport-Security) with the `max-age` of `server.hsts_max_age`, one year by default. Limits of this setup:

- The link between the proxy and the server is plain HTTP: passwords and cookies cross it in clear, and the server logs no warning.
- Any machine that reaches the listening port can talk to the server without the proxy, by sending the public host name.
- Restrict the listening port to the proxy with a firewall, or serve TLS on that link too with `server.tls_cert` and `server.tls_key`.

## LOGS

- The server logs one line per request (`msg=request`) and one line per event, such as a login or a saved month. The fields are listed in [Logs](SPECIFICATIONS.md#logs).
- The level of the server is `server.log_level`, `info` by default. `warn` keeps only the warnings and the errors. `--log-level` replaces it for one run, and sets the level of the other commands.
- The commands of the command line log no request and no event: they only print their warnings on the standard error, in the same format.

| | Repository | Package |
| --- | --- | --- |
| Read the logs | the standard error of `bin/timesheet.sh` or `bin/timesheet run` | `journalctl -u timesheet` |
| Change the level | `bin/timesheet config set server.log_level warn`, or `bin/timesheet.sh --log-level=warn` for one run | `timesheet config set server.log_level warn --config=/etc/timesheet/timesheet.json` as root, then `systemctl restart timesheet` |
| Retention | what you keep of the standard error | the settings of the journal (`/etc/systemd/journald.conf`) |

What the logs hold about people is listed in [Personal data](SPECIFICATIONS.md#personal-data).

## IMPORTING

Days, counter settings and adjustments can be imported from a JSON file of the format `timesheet-import`, described in [Import](SPECIFICATIONS.md#import).

- The file holds the tracking start, counter settings, days with morning, afternoon and comment, and adjustments.
- The application reads no spreadsheet. A spreadsheet is first converted to this format outside the application, its categories becoming the activity codes.
- `share/timesheet/import-template.json` is a small, fictitious file of this format: one month (January 2026) with a holiday, paid leave, a half-day of RTT, an absence, comments, the settings of three counters and one adjustment. Copy it as a starting point.

To import:

1. Create the account: from « Comptes » (provisional password, changed at the first login) or with `user add` (the operator, who runs the command, knows the password until the user changes it).
2. Stop the server.
3. Run a dry run until it reports no error:

   ```sh
   bin/timesheet import <login> ~/private/activity.json --dry-run --database=var/lib/timesheet/timesheet-database.json
   ```

   Try it first on the shipped template, `bin/timesheet import <login> share/timesheet/import-template.json --dry-run --database=var/lib/timesheet/timesheet-database.json`, which answers:

   ```text
   Settings read: tracking start 2026-01-01, 3 counter(s).
   31 day(s) read, from 2026-01 to 2026-01.
   1 adjustment(s) read.
   Dry run: nothing was written.
   ```

   - Every day is checked like an entry of the grid, every adjustment like one of « Activité ». Any problem is listed with its date and nothing is written.
   - It lists, month by month, the days before today left without entry, weekends and holidays included.
   - With `--validate`, it also lists the months that would be validated.

4. Import for real, without `--dry-run`. Use `--validate` only if no missing day is listed: each month is then recorded as validated on its last day. Otherwise complete the listed days in the grid and validate month by month.
5. Start the server.

- Other options: `--from=YYYY-MM` and `--to=YYYY-MM` limit the months; `--replace` replaces the days already entered at the dates of the file.
- Gaps with the payslip or hand corrections of a balance are adjustments, in the file or in « Activité ».
- The spreadsheet and the file made from it hold personal data. Keep them out of the repository, and give the file by a path outside it: a file inside the repository may be committed by mistake.
- On a package, the command runs as the account `timesheet` (`runuser -u timesheet -- timesheet import … --config=/etc/timesheet/timesheet.json`): the file must be readable by that account.

## EXAMPLE DATABASE

`share/timesheet/database-example.json` is an anonymized database with one account and nearly three years of entries, from November 2023 to December 2026, validated up to October 2026 ([Example database](SPECIFICATIONS.md#example-database)). To try the application with it, without touching your database:

1. Copy it, mode 0600, into a directory of your own:

   ```sh
   mkdir -m 700 ~/timesheet-example
   install -m 600 share/timesheet/database-example.json ~/timesheet-example/timesheet-database.json
   ```

2. Start a server on it, on another port if a server already runs on 8080:

   ```sh
   bin/timesheet run --listen=127.0.0.1:8081 --database="$HOME/timesheet-example/timesheet-database.json"
   ```

3. Open `http://127.0.0.1:8081` and log in with the login `demo` and the password `demonstration`.

- The password is public: never serve this database on a network, and never use it as a real database.
- Every change is written into the copy; copy the file again to start over.
- On a package, the file is `/usr/share/timesheet/database-example.json`.

## DEBIAN PACKAGE

### What the package installs

The package is built as described in [BUILD.md](BUILD.md#build-the-debian-package), into `_dist/`.

It installs:

- `/usr/bin/timesheet`;
- the configuration template, the example database and the import template in `/usr/share/timesheet/`;
- the server configuration `/etc/timesheet/timesheet.json`, the file of the repository installed as is, whose `database.path` is `/var/lib/timesheet/timesheet-database.json`; it is a configuration file of the package: an upgrade keeps your changes;
- the unit `timesheet.service`, which runs `timesheet run --config=/etc/timesheet/timesheet.json` as the account `timesheet`; the database comes from that file.

Its installation script creates, each when missing:

- the system account `timesheet`;
- the directory `/var/lib/timesheet`, mode 0700, owned by that account;
- the database at the `database.path` of `/etc/timesheet/timesheet.json`, `/var/lib/timesheet/timesheet-database.json` by default. A path outside `/var/lib/timesheet` is not created: the script warns, and you create it with `timesheet init`. When the server configuration cannot be read, the script warns and creates no database.

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
   runuser -u timesheet -- timesheet user add <login> --profile=ADMIN --config=/etc/timesheet/timesheet.json
   systemctl start timesheet
   ```

   If `systemctl start` refuses because the service failed too often, run `systemctl reset-failed timesheet` first.

4. Check the journal: `journalctl -u timesheet`. The service listens on `http://127.0.0.1:8080`.
5. To change the address, the origins, the trusted proxies, the time zone, TLS, the logs, the sessions or the limits, see [Server configuration](#server-configuration).

### Upgrade the package

1. Back up the database (see [Backup and restore](#backup-and-restore)).
2. If the database was created before the server settings moved to `/etc/timesheet/timesheet.json`, stop the service (`systemctl stop timesheet`) and migrate it first, as described at the end of [Upgrade](#upgrade).
3. If you changed `/etc/timesheet/timesheet.json` by hand and it still holds the key `addr`, rename it `listen` before or right after the installation ([Listening address renamed](SPECIFICATIONS.md#listening-address-renamed)): otherwise the service refuses to start and the installation warns that the file cannot be read. Replace `--addr` by `--listen` in your drop-ins too.
4. Install the new package, as root: `apt install ./_dist/timesheet_<version>_amd64.deb`. The package restarts the service.
5. When both you and the new package changed `/etc/timesheet/timesheet.json`, `apt` asks which version to keep. Keeping yours keeps your values: the keys the new version added take their default.
6. Check the journal: `journalctl -u timesheet`.

### Options of the service

The unit runs `timesheet run --config=/etc/timesheet/timesheet.json`: every setting, TLS and the log level included, comes from `/etc/timesheet/timesheet.json` ([Server configuration](#server-configuration)). For HTTPS:

1. Put the certificate pair in `/etc/timesheet/`, readable by the account `timesheet`.
2. Set it, with https origins, as root:

   ```sh
   timesheet config set server.tls_cert /etc/timesheet/cert.pem server.tls_key /etc/timesheet/key.pem server.origins https://timesheet.example.org --config=/etc/timesheet/timesheet.json
   ```

3. Run `systemctl restart timesheet`.

To move the database:

- Inside `/var/lib/timesheet`, the directory of the service, stop the service, change `database.path` with `config set`, move the file and its session file `.<name>.sessions` to the new name, then start the service.
- Elsewhere, the sandbox of the unit (`ProtectSystem=strict`) allows writes only in `/var/lib/timesheet`. Create the directory with mode 0700 for the account `timesheet`, allow it in a drop-in of the unit, then change `database.path`:

  ```ini
  [Service]
  ReadWritePaths=/srv/timesheet
  ```

  Write it with `systemctl edit timesheet`, then run `systemctl restart timesheet`. The installation script of the package does not create a database outside `/var/lib/timesheet`.

A drop-in written for a former version, with an `ExecStart=` that passes `--config`, `--database`, `--tls-cert` or `--log-level`, still works, since the options come first; it must not pass `--addr`, now `--listen`.

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

3. If the sessions saved next to the database must not stay open, close them: `bin/timesheet session clear --database=var/lib/timesheet/timesheet-database.json`, or on a package `runuser -u timesheet -- timesheet session clear --config=/etc/timesheet/timesheet.json`.
4. Start the server again. On a package: `systemctl start timesheet`.

A restored backup brings back the accounts, passwords, profiles and states of its date. Change again the passwords changed since, and disable or delete again the accounts disabled or deleted since.

## RECOVERY

When the password of an administrator is lost:

1. Stop the server. On a package: `systemctl stop timesheet`.
2. Set a new password:

   Repository:

   ```sh
   bin/timesheet user passwd <login> --database=var/lib/timesheet/timesheet-database.json
   ```

   Package, as root:

   ```sh
   runuser -u timesheet -- timesheet user passwd <login> --config=/etc/timesheet/timesheet.json
   ```

3. Start the server. On a package: `systemctl start timesheet`.

- The password is permanent and the sessions of the account are closed.
- `user passwd` does not enable a disabled account: enable it from « Comptes » with another administrator.

## UPGRADE

This procedure upgrades the repository form. For the package, see [Upgrade the package](#upgrade-the-package).

1. Back up the database.
2. Stop the server.
3. Get the new sources and rebuild with `make` ([BUILD.md](BUILD.md#build-the-program)): the templates and the assets are embedded in the binary.
4. Apply the migration notes that concern your installation: see [Migration notes](SPECIFICATIONS.md#migration-notes). The ones known today are below.
5. Start the server again.

A database created before the server settings moved to the server configuration file holds a `server` object in its `configuration`. The new binary refuses it with `store: malformed database: json: unknown field "server"`. To migrate it, with the server stopped and a backup taken:

1. Read the former settings:

   ```sh
   jq '.configuration.server' var/lib/timesheet/timesheet-database.json
   ```

2. Copy the former `addr` into `server.listen`, and `origins` and `timezone` into `server.origins` and `server.timezone`, with `config set`.
3. Write the former `trusted_proxy` as the list `server.trusted_proxies`: its address, or `config set server.trusted_proxies ''` (written `[]`) when it was `null` (no trusted proxy).
4. Remove the `server` object from the database:

   ```sh
   jq 'del(.configuration.server)' var/lib/timesheet/timesheet-database.json > <new file>
   install -m 600 <new file> var/lib/timesheet/timesheet-database.json
   ```

On a package, use `/var/lib/timesheet/timesheet-database.json`, `--config=/etc/timesheet/timesheet.json` and `install -m 600 -o timesheet -g timesheet`, as root.

A server configuration written before the key `addr` became `listen` is refused with `json: unknown field "addr"`, by `run` and by every command that reads it. Rename the key in the file by hand: `config show` and `config set` refuse the file until then. Replace `--addr` by `--listen` in your scripts.

The binary built from the sources no longer opens `var/lib/timesheet/timesheet-database.json` by default: the default database is now `database.path` of the server configuration, `/var/lib/timesheet/timesheet-database.json`. Start the server with `bin/timesheet.sh`, which passes the database of the repository, and give `--database=var/lib/timesheet/timesheet-database.json` to the other commands ([Database path in the server configuration](SPECIFICATIONS.md#database-path-in-the-server-configuration)).

Other points:

- The database is checked entirely at load. An older binary refuses a database written by a newer one: keep the backup taken before the upgrade to go back.
- A server configuration file using the former key `trusted_proxy` is refused (`json: unknown field "trusted_proxy"`): rename it `trusted_proxies` and write it as a list.
- Sessions are saved in `.<database file name>.sessions`. A `timesheet-sessions.json` left by an older version is no longer read, so its users log in again. Delete it: it holds the CSRF tokens of its sessions.
- The certificate pair, when used, is read at start: restart the server after renewing it.

## TROUBLESHOOTING

The errors of the commands are prefixed with `timesheet:`. Those of `run` are log lines `level=ERROR msg=…`, except `no administrator account`.

| Message | Cause | What to do |
| --- | --- | --- |
| `open share/timesheet/database-template.json: no such file or directory` | `init` is run outside the repository root, or by the installed binary. | Go to the root, or give `--template` and `--database`. |
| `cannot lock …` with `the directory does not exist` | `run` is started with a wrong `--database` or `database.path`, or a relative path from another directory. | Run `bin/timesheet.sh`, or give `--database`. |
| `the database does not exist, run init first`, or a permission error, naming `/var/lib/timesheet/timesheet-database.json` | A command of the sources is run without `--database`: the default database is the one of the package. | Give `--database=var/lib/timesheet/timesheet-database.json`. |
| `the database does not exist, run init first` | The database file is missing, at `database.path` or at `--database` (for `user list`, also when its directory is missing). | Run `init`, or fix the path. |
| `json: unknown field "addr"` | The server configuration was written before `addr` became `listen`. | Rename the key `listen` ([Upgrade](#upgrade)). |
| `<section>.<key>: a duration from … to … is expected`, or `an integer from … to … is expected` | A value of the server configuration is out of its bounds. | Fix it with `config set`; the bounds are in [Server configuration](SPECIFICATIONS.md#server-configuration). |
| `config set` refuses `sessions.idle_timeout`, `login.initial_delay` or `server.tls_cert` alone | A linked setting does not agree with the current value of the other. | Set both in the same command. |
| The service is killed by systemd while stopping | `server.shutdown_timeout` was raised above what systemd waits. | Keep it at 60 s at most (the bound); systemd waits 90 s by default. |
| `the database already exists` | `init` is run a second time. | Nothing: the database is ready. For another one, give `--database`. |
| `insecure directory: … is writable by other` | The directory of the database is open to other users. | `chmod 700` on that directory. |
| `insecure file: … is accessible to group or other (expected 0600)` | The database was copied with a wrong mode, often by a restore. | `chmod 600` on the file. |
| `… belongs to another user` | The command runs as another user than the owner of the file, such as root on a package. | Run it as the owner: `runuser -u timesheet -- …` on a package. |
| `malformed database: json: unknown field "server"` | The database was created before the server settings moved to their own file. | Follow [Upgrade](#upgrade). |
| `the password needs at least 12 characters…` | The password is refused by the policy. | Choose a longer or more varied one. |
| `no administrator account: run 'timesheet user add …' first` | `run` finds no enabled administrator. | Create one with `user add <login> --profile=ADMIN`. |
| `cannot lock …` with `the file is used by another process` | Another instance already serves this database. | Use that instance, or stop it. |
| `cannot listen on 127.0.0.1:8080` with `address already in use` | Another program, or an instance serving another database, uses the port. | Stop it, or start with `--listen=127.0.0.1:8081`: the accepted origins then follow that address. |
| `the file is used by another process … (stop the server or use /admin/accounts)`, or `(stop the server first)` | A command that needs the server stopped is run while the server holds the database. | Stop the server first, or act from « Administration ». |
| « Hôte inconnu » (421) | The page is opened under a host name that is not an accepted origin. | Use `127.0.0.1` or `localhost`, or declare the name in `server.origins` (`config set server.origins`) or with `--origin`. |
| « Identifiant ou mot de passe incorrect. » | Wrong login or password, or disabled account. | Try again. Each failure makes the address wait, from `login.initial_delay` up to `login.max_delay`, 1 s and 5 min by default (« Trop de tentatives », 429). |
| « Le formulaire a expiré ou n'est pas valide. Rechargez la page et recommencez. » (403) | The page was open too long (the login form lives `login.form_lifetime`, 1 h by default), or the session was replaced. | Reload the page and submit again. |
| « Les données ont changé entre-temps… » (409) | The same account was changed in another tab. | Reload the page and enter the change again. |

The limits behind other refusals (size of a request, of an account, number of accounts) are in [Limits](SPECIFICATIONS.md#limits).
