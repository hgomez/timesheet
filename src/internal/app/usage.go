/*
 * usage.go - Copyright (c) 2023-2026 - Olivier Poncet
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package app

import (
    "fmt"
    "io"
    "strings"
)

// ---------------------------------------------------------------------------
// Usage declaration
// ---------------------------------------------------------------------------

type Usage struct {
    program string
    groups  []usageGroup
}

type usageGroup struct {
    title   string
    options [][2]string
}

// ---------------------------------------------------------------------------
// Usage implementation
// ---------------------------------------------------------------------------

func NewUsage(program string) *Usage {
    return &Usage{
        program: program,
        groups: []usageGroup{
            {title: "Commands", options: [][2]string{
                {"init", "create the database from the configuration template"},
                {"run", "start the web server"},
                {"config show [KEY]", "show the server configuration, or the value of a single key"},
                {"config set KEY VALUE...", "change one or more settings of the configuration file (the keys of config show)"},
                {"user add LOGIN", "create an account (password asked)"},
                {"user profile LOGIN PROFILE", "change the profile of an account (USER or ADMIN)"},
                {"user passwd LOGIN", "change the password of an account"},
                {"user del LOGIN", "delete an account and its data"},
                {"user list", "list the accounts"},
                {"session clear", "close all the sessions (server stopped)"},
                {"import LOGIN FILE", "import a JSON file (format timesheet-import)"},
                {"help", "show this help"},
                {"version", "show the version"},
            }},
            {title: "Common options", options: [][2]string{
                {"--database=<file>", "single database (database.path of --config when absent; no default)"},
                {"--config=<file>", "server configuration (the defaults without it; required by config show and config set)"},
                {"--log-level=<level>", "debug, info, warn or error (default info, server.log_level of the configuration for run)"},
                {"-h, --help", "show this help"},
            }},
            {title: "Options of init", options: [][2]string{
                {"--template=<file>", "configuration template (required)"},
            }},
            {title: "Options of run", options: [][2]string{
                {"--listen=<address>", "listen address (default the one of the configuration)"},
                {"--origin=<origin>", "public origin, repeatable (default the ones of the configuration)"},
                {"--trusted-proxy=<address>", "proxy whose X-Forwarded-For is trusted, address or prefix, repeatable (default the ones of the configuration)"},
                {"--tls-cert=<file>", "certificate of a provided pair (HTTPS, default the one of the configuration)"},
                {"--tls-key=<file>", "private key of a provided pair (HTTPS, default the one of the configuration)"},
            }},
            {title: "Options of user add", options: [][2]string{
                {"--profile=<profile>", "USER or ADMIN (default USER)"},
                {"--provisional", "the password must be changed at the first login"},
            }},
            {title: "Options of import", options: [][2]string{
                {"--from=<YYYY-MM>", "first imported month"},
                {"--to=<YYYY-MM>", "last imported month"},
                {"--dry-run", "check without writing anything"},
                {"--replace", "replace the existing entries at the dates present in the file"},
                {"--validate", "then validate the imported months, in order"},
            }},
        },
    }
}

func (u *Usage) Write(w io.Writer) {
    fmt.Fprintf(w, "Usage: %s COMMAND [ARGUMENT...] [OPTION...]\n\n", u.program)
    width := 0
    for _, group := range u.groups {
        for _, option := range group.options {
            width = max(width, len(option[0]))
        }
    }
    for _, group := range u.groups {
        fmt.Fprintf(w, "%s:\n", group.title)
        for _, option := range group.options {
            fmt.Fprintf(w, "  %s%s  %s\n", option[0], strings.Repeat(" ", width - len(option[0])), option[1])
        }
        fmt.Fprintln(w)
    }
    fmt.Fprintf(w, "The password of user add and user passwd is read from the terminal without echo, otherwise from a line of the standard input, or from the variable TIMESHEET_PASSWORD.\nOptions are written --name=value, before or after the arguments; an argument cannot start with a dash.\n")
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
