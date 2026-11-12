/*
 * config_command.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "log/slog"
    "net"
    "strings"

    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// ConfigCommand declaration
// ---------------------------------------------------------------------------

type ConfigCommand struct {
    settings *Settings
    logger   *slog.Logger
}

// ---------------------------------------------------------------------------
// ConfigCommand implementation
// ---------------------------------------------------------------------------

func NewConfigCommand(settings *Settings, logger *slog.Logger) *ConfigCommand {
    return &ConfigCommand{settings: settings, logger: logger}
}

func (c *ConfigCommand) Run(out io.Writer) error {
    file := NewServerFile(c.settings.Config())
    server, err := file.Load()
    if err != nil {
        return err
    }
    if c.settings.Command() == "config show" {
        data, err := file.Encode(server)
        if err != nil {
            return err
        }
        _, err = out.Write(data)
        return err
    }
    value := c.settings.Argument(1)
    switch c.settings.Argument(0) {
        case "server.addr":
            if _, port, err := net.SplitHostPort(value); err == nil {
                if _, err := net.LookupPort("tcp", port); err != nil {
                    return fmt.Errorf("%w: server.addr: a port from 0 to 65535 or a service name is expected", activity.ErrInvalidConfiguration)
                }
            }
            server.Addr = value
        case "server.origins":
            server.Origins = c.list(value)
        case "server.trusted_proxies":
            server.TrustedProxies = c.list(value)
        case "server.timezone":
            server.Timezone = value
        default:
            return NewUsageError("unknown key " + c.settings.Argument(0) + " (server.addr, server.origins, server.trusted_proxies, server.timezone)")
    }
    if err := file.Save(server); err != nil {
        return err
    }
    fmt.Fprintf(out, "%s mis à jour dans %s ; relancez le serveur pour l'appliquer.\n", c.settings.Argument(0), file.Path())
    return nil
}

func (c *ConfigCommand) list(value string) []string {
    values := []string{}
    for _, element := range strings.Split(value, ",") {
        if element = strings.TrimSpace(element); element != "" {
            values = append(values, element)
        }
    }
    return values
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
