# BUILD

This document describes how to build Timesheet from its sources: the program, its tests, the installation of its files and the Debian package.

It is written for whoever builds the software. How the build scripts are written is in [CODING.md](CODING.md#makefile); how to run the result is in [OPERATIONS.md](OPERATIONS.md).

## CONTENTS

- [REQUIREMENTS](#requirements)
- [BUILD THE PROGRAM](#build-the-program)
- [RUN THE TESTS](#run-the-tests)
- [INSTALL THE FILES](#install-the-files)
- [CLEAN](#clean)
- [BUILD THE DEBIAN PACKAGE](#build-the-debian-package)
- [VERSION](#version)

## REQUIREMENTS

| To | You need |
| --- | --- |
| Build the program | Go 1.24 or later, and `make`. |
| Run the tests | `gcc` as well: the race detector needs cgo. |
| Build the Debian package | A Debian system with `build-essential`, `debhelper`, `devscripts`, `golang` and `git`. |

The program has no dependency besides the Go standard library: nothing is downloaded during the build.

Run every command of this document from the root of the repository.

## BUILD THE PROGRAM

```sh
make
```

- It produces `bin/timesheet`, a static binary that runs without any other file: the pages, the styles and the scripts are embedded in it.
- It ends with `=== build ok ===`.
- The build cache is kept in `.cache/`, at the root of the repository. The first build, and the first build after `make clean`, take a few seconds more.
- Rebuild after every change of the sources, templates and styles included.

## RUN THE TESTS

```sh
make check
```

- It runs every test with the race detector, and ends with `=== check ok ===`.
- It needs `gcc`.

## INSTALL THE FILES

`make install` copies the built files into a directory tree. It does not build: run `make` first.

```sh
make
make install DESTDIR=/tmp/stage
```

| File | Installed as | Mode |
| --- | --- | --- |
| `bin/timesheet` | `$(DESTDIR)/usr/bin/timesheet` | 755 |
| `share/timesheet/timesheet-database-template.json` (configuration template) | `$(DESTDIR)/usr/share/timesheet/` | 644 |
| `share/timesheet/timesheet-import-template.json` (import template) | `$(DESTDIR)/usr/share/timesheet/` | 644 |
| `etc/timesheet/timesheet.json` (server configuration) | `$(DESTDIR)/etc/timesheet/` | 644 |

- Without `DESTDIR`, the files go under `/usr` and `/etc`: run `sudo make install`.
- It fails when `bin/timesheet` is missing.
- It replaces an `/etc/timesheet/timesheet.json` already there: save your own first.
- It creates neither the service account, nor the database directory, nor the service. The [Debian package](#build-the-debian-package) does.
- The installed binary still takes its default paths from the current directory: give it `--config=/etc/timesheet/timesheet.json`, `--database=…` and, for `init`, `--template=/usr/share/timesheet/timesheet-database-template.json`.

## CLEAN

| Command | Effect |
| --- | --- |
| `make clean` | Removes `bin/timesheet` and empties the build cache `.cache/`. The database is kept. |
| `make mrproper` | Runs `git clean -f -d -x`: deletes every file that git does not track, the database under `var/lib/timesheet/` and any new source file included. |

Never run `make mrproper` without a backup of the database and a commit of your new files.

## BUILD THE DEBIAN PACKAGE

1. On a Debian system with the [requirements](#requirements), from the root of the repository:

   ```sh
   .ci/ci-build-package.sh
   ```

2. The build runs `make check` and stops if a test fails.
3. The packages are written into `_dist/`, at the root of the repository: `timesheet_<version>_amd64.deb` (the architecture is the one of the build machine), its debug symbols `timesheet-dbgsym_<version>_amd64.deb`, and the `.buildinfo` and `.changes` files.
4. To remove the build leftovers afterwards:

   ```sh
   .ci/ci-cleanup.sh
   ```

- The package is built from the working tree, uncommitted changes included.
- `<version>` is the first entry of `debian/changelog`.
- Installing and running the package is described in [OPERATIONS.md](OPERATIONS.md#debian-package).

## VERSION

```sh
bin/timesheet version
```

It prints the release number, the commit the binary was built from and its date, and `modified` when the tree had uncommitted changes. Releasing a new version is described in [CODING.md](CODING.md#release).
