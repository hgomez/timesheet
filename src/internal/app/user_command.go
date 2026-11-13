/*
 * user_command.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "text/tabwriter"
    "time"

    "timesheet/internal/account"
    "timesheet/internal/activity"
    "timesheet/internal/store"
)

// ---------------------------------------------------------------------------
// UserCommand declaration
// ---------------------------------------------------------------------------

type UserCommand struct {
    settings   *Settings
    server     activity.ServerConfiguration
    passwords  *PasswordReader
    clock      activity.Clock
    random     io.Reader
    iterations int
    logger     *slog.Logger
}

// ---------------------------------------------------------------------------
// UserCommand implementation
// ---------------------------------------------------------------------------

func NewUserCommand(settings *Settings, server activity.ServerConfiguration, passwords *PasswordReader, clock activity.Clock, random io.Reader, iterations int, logger *slog.Logger) *UserCommand {
    return &UserCommand{settings: settings, server: server, passwords: passwords, clock: clock, random: random, iterations: iterations, logger: logger}
}

func (c *UserCommand) Run(out io.Writer) error {
    database := store.NewDatabase(c.settings.Database(), c.logger, c.server.Database.MaxBytes, c.server.Limits)
    if c.settings.Command() != "user list" {
        err := database.Lock()
        if errors.Is(err, store.ErrLocked) {
            return fmt.Errorf("%w (stop the server or use /admin/)", err)
        }
        if err != nil {
            return err
        }
        defer database.Unlock()
    }
    if err := database.Load(); err != nil {
        return err
    }
    if c.settings.Command() == "user list" {
        return c.list(out, database)
    }
    login, err := account.NewLogin(c.settings.Argument(0))
    if err != nil {
        return err
    }
    if c.settings.Command() == "user add" {
        return c.add(out, database, login)
    }
    if _, _, found := database.Account(login); !found {
        return account.ErrUnknownAccount
    }
    switch c.settings.Command() {
        case "user profile":
            profile, err := account.NewProfile(c.settings.Argument(1))
            if err != nil {
                return err
            }
            if err := c.closeSessions(login); err != nil {
                return err
            }
            if err := database.SetProfile(login, profile); err != nil {
                return err
            }
            fmt.Fprintf(out, "Profil de %s : %s\n", login, profile)
        case "user passwd":
            hash, err := c.password(login)
            if err != nil {
                return err
            }
            if err := c.closeSessions(login); err != nil {
                return err
            }
            if err := database.SetPassword(login, hash, false); err != nil {
                return err
            }
            fmt.Fprintf(out, "Mot de passe de %s changé.\n", login)
        case "user del":
            if err := c.closeSessions(login); err != nil {
                return err
            }
            if err := database.DeleteAccount(login); err != nil {
                return err
            }
            fmt.Fprintf(out, "Compte %s supprimé.\n", login)
    }
    return nil
}

func (c *UserCommand) add(out io.Writer, database *store.Database, login account.Login) error {
    profile, err := account.NewProfile(c.settings.Profile())
    if err != nil {
        return err
    }
    if _, _, found := database.Account(login); found {
        return account.ErrDuplicateLogin
    }
    hash, err := c.password(login)
    if err != nil {
        return err
    }
    now := c.clock.Now()
    today := activity.NewDateOf(now.In(c.server.Server.Location()))
    created := account.Account{Login: login, Profile: profile, State: account.StateEnabled, Password: hash, Created: now.UTC().Truncate(time.Second)}
    if err := database.CreateAccount(created, activity.NewCalendar(today.Month().First())); err != nil {
        return err
    }
    fmt.Fprintf(out, "Compte %s créé (%s).\n", login, profile)
    return nil
}

func (c *UserCommand) closeSessions(login account.Login) error {
    sessions := store.NewSessionFile(c.settings.Database(), c.logger)
    records, err := sessions.LoadSessions()
    if err != nil {
        c.logger.Warn("unreadable session file replaced by an empty one", "error", err.Error())
        records = nil
    }
    kept := []account.SessionRecord{}
    for _, record := range records {
        if record.Login != login {
            kept = append(kept, record)
        }
    }
    return sessions.SaveSessions(kept)
}

func (c *UserCommand) password(login account.Login) (account.PasswordHash, error) {
    password, err := c.passwords.Read()
    if err != nil {
        return "", err
    }
    if err := account.NewPasswordPolicy().Check(login, password); err != nil {
        return "", err
    }
    return account.NewPasswordHash(password, c.iterations, c.random)
}

func (c *UserCommand) list(out io.Writer, database *store.Database) error {
    writer := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
    fmt.Fprintln(writer, "LOGIN\tPROFIL\tÉTAT\tCRÉÉ LE\tDERNIÈRE CONNEXION\tNOM")
    for _, item := range database.Accounts() {
        last := "jamais"
        if !item.LastLogin.IsZero() {
            last = item.LastLogin.Format(time.RFC3339)
        }
        fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n", item.Login, item.Profile, item.State, item.Created.Format(time.RFC3339), last, item.Name)
    }
    return writer.Flush()
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
