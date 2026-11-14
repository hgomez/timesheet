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
    "errors"
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
    configuration, err := file.Load()
    if err != nil {
        return err
    }
    arguments := c.settings.Arguments()
    if c.settings.Command() == "config show" {
        return c.show(out, file, configuration, arguments)
    }
    keys := []string{}
    for i := 0; i < len(arguments); i += 2 {
        key, value := arguments[i], arguments[i + 1]
        if err := c.check(key, value); err != nil {
            return err
        }
        if err := configuration.Set(key, value); err != nil {
            return c.unknown(err, key)
        }
        keys = append(keys, key)
    }
    if err := file.Save(configuration); err != nil {
        return err
    }
    fmt.Fprintf(out, "%s updated in %s; restart the server to apply it.\n", strings.Join(keys, ", "), file.Path())
    return nil
}

func (c *ConfigCommand) show(out io.Writer, file *ServerFile, configuration activity.ServerConfiguration, arguments []string) error {
    if len(arguments) == 1 {
        value, err := configuration.Value(arguments[0])
        if err != nil {
            return c.unknown(err, arguments[0])
        }
        _, err = fmt.Fprintln(out, value)
        return err
    }
    data, err := file.Encode(configuration)
    if err != nil {
        return err
    }
    _, err = out.Write(data)
    return err
}

func (c *ConfigCommand) check(key string, value string) error {
    if key != "server.listen" {
        return nil
    }
    if _, port, err := net.SplitHostPort(value); err == nil {
        if _, err := net.LookupPort("tcp", port); err != nil {
            return fmt.Errorf("%w: server.listen: a port from 0 to 65535 or a service name is expected", activity.ErrInvalidConfiguration)
        }
    }
    return nil
}

func (c *ConfigCommand) unknown(err error, key string) error {
    if errors.Is(err, activity.ErrUnknownSetting) {
        return NewUsageError("unknown key " + key + " (the keys are those printed by config show)")
    }
    return err
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
