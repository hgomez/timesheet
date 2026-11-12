/*
 * session_command.go - Copyright (c) 2023-2026 - Olivier Poncet
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

    "timesheet/internal/store"
)

// ---------------------------------------------------------------------------
// SessionCommand declaration
// ---------------------------------------------------------------------------

type SessionCommand struct {
    settings *Settings
    logger   *slog.Logger
}

// ---------------------------------------------------------------------------
// SessionCommand implementation
// ---------------------------------------------------------------------------

func NewSessionCommand(settings *Settings, logger *slog.Logger) *SessionCommand {
    return &SessionCommand{settings: settings, logger: logger}
}

func (c *SessionCommand) Run(out io.Writer) error {
    database := store.NewDatabase(c.settings.Database(), c.logger)
    err := database.Lock()
    if errors.Is(err, store.ErrLocked) {
        return fmt.Errorf("%w (stop the server first)", err)
    }
    if err != nil {
        return err
    }
    defer database.Unlock()
    if err := database.Load(); err != nil {
        return err
    }
    sessions := store.NewSessionFile(c.settings.Database(), c.logger)
    if err := sessions.SaveSessions(nil); err != nil {
        return err
    }
    fmt.Fprintln(out, "Toutes les sessions sont fermées.")
    return nil
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
