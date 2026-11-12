/*
 * init_command.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "log/slog"
    "os"

    "timesheet/internal/activity"
    "timesheet/internal/store"
)

// ---------------------------------------------------------------------------
// InitCommand declaration
// ---------------------------------------------------------------------------

type InitCommand struct {
    settings *Settings
    logger   *slog.Logger
}

const maxTemplateSize = 1 << 20

// ---------------------------------------------------------------------------
// InitCommand implementation
// ---------------------------------------------------------------------------

func NewInitCommand(settings *Settings, logger *slog.Logger) *InitCommand {
    return &InitCommand{settings: settings, logger: logger}
}

func (c *InitCommand) Run(out io.Writer) error {
    config, err := c.template()
    if err != nil {
        return err
    }
    database := store.NewDatabase(c.settings.Database(), c.logger)
    if err := database.Lock(); err != nil {
        return err
    }
    defer database.Unlock()
    if err := database.Create(config); err != nil {
        return err
    }
    fmt.Fprintf(out, "Base créée : %s\nCréez maintenant un administrateur : %s user add <login> --profile=ADMIN --database=%s\n", database.Path(), ProgramName, c.settings.Database())
    return nil
}

func (c *InitCommand) template() (activity.Configuration, error) {
    file, err := os.Open(c.settings.Template())
    if err != nil {
        return activity.Configuration{}, err
    }
    defer file.Close()
    data, err := io.ReadAll(io.LimitReader(file, maxTemplateSize + 1))
    if err != nil {
        return activity.Configuration{}, err
    }
    if len(data) > maxTemplateSize {
        return activity.Configuration{}, fmt.Errorf("%s: larger than %d MiB", c.settings.Template(), maxTemplateSize >> 20)
    }
    config := activity.Configuration{}
    decoder := json.NewDecoder(bytes.NewReader(data))
    decoder.DisallowUnknownFields()
    if err := decoder.Decode(&config); err != nil {
        return activity.Configuration{}, fmt.Errorf("%s: %w", c.settings.Template(), err)
    }
    if _, err := decoder.Token(); err != io.EOF {
        return activity.Configuration{}, fmt.Errorf("%s: data after the configuration", c.settings.Template())
    }
    if err := config.Validate(); err != nil {
        return activity.Configuration{}, fmt.Errorf("%s: %w", c.settings.Template(), err)
    }
    return config, nil
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
