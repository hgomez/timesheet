/*
 * import_command.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "os"
    "strings"

    "timesheet/internal/account"
    "timesheet/internal/activity"
    "timesheet/internal/store"
    "timesheet/internal/web"
)

// ---------------------------------------------------------------------------
// ImportCommand declaration
// ---------------------------------------------------------------------------

type ImportCommand struct {
    settings *Settings
    server   activity.ServerConfiguration
    clock    activity.Clock
    random   io.Reader
    logger   *slog.Logger
}

var ErrImportRefused = errors.New("app: import refused")

// ---------------------------------------------------------------------------
// ImportCommand implementation
// ---------------------------------------------------------------------------

func NewImportCommand(settings *Settings, server activity.ServerConfiguration, clock activity.Clock, random io.Reader, logger *slog.Logger) *ImportCommand {
    return &ImportCommand{settings: settings, server: server, clock: clock, random: random, logger: logger}
}

func (c *ImportCommand) Run(out io.Writer) error {
    options, err := c.options()
    if err != nil {
        return err
    }
    database := store.NewDatabase(c.settings.Database(), c.logger, c.server.Database.MaxBytes, c.server.Limits)
    err = database.Lock()
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
    login, err := account.NewLogin(c.settings.Argument(0))
    if err != nil {
        return err
    }
    _, calendar, found := database.Account(login)
    if !found {
        return account.ErrUnknownAccount
    }
    imported, err := c.read(c.settings.Argument(1))
    if err != nil {
        return err
    }
    location := c.server.Server.Location()
    today := activity.NewDateOf(c.clock.Now().In(location))
    horizon := today.Month()
    for range web.PlanningMonths {
        horizon = horizon.Next()
    }
    importer := activity.NewImporter(database.Configuration(), c.clock, activity.NewIdentifierGenerator(c.random), location)
    report, err := importer.Run(calendar, imported, options, today, horizon)
    if err != nil {
        return err
    }
    for _, warning := range report.Warnings {
        fmt.Fprintln(out, "warning: " + warning.Text())
    }
    if len(report.Problems) > 0 {
        for _, problem := range report.Problems {
            fmt.Fprintln(out, "error: " + problem.Text())
        }
        return fmt.Errorf("%w: %d error(s)", ErrImportRefused, len(report.Problems))
    }
    c.print(out, report)
    if c.settings.DryRun() {
        if err := database.CheckCalendar(login, calendar.Revision, report.Calendar); err != nil {
            return err
        }
        fmt.Fprintln(out, "Dry run: nothing was written.")
        return nil
    }
    if err := database.SaveCalendar(login, calendar.Revision, report.Calendar); err != nil {
        return err
    }
    fmt.Fprintln(out, "Import saved.")
    return nil
}

func (c *ImportCommand) print(out io.Writer, report activity.ImportReport) {
    if report.Settings {
        fmt.Fprintf(out, "Settings read: tracking start %s, %d counter(s).\n", report.Calendar.Settings.Start, report.Counters)
    }
    if len(report.Dates) > 0 {
        fmt.Fprintf(out, "%d day(s) read, from %s to %s.\n", len(report.Dates), report.First(), report.Last())
        for _, missing := range report.Missing {
            dates := []string{}
            for _, d := range missing.Dates {
                dates = append(dates, d.String())
            }
            fmt.Fprintf(out, "missing entries in %s: %s\n", missing.Month, strings.Join(dates, ", "))
        }
    }
    if report.Adjustments > 0 {
        fmt.Fprintf(out, "%d adjustment(s) read.\n", report.Adjustments)
    }
    for _, m := range report.Validated {
        fmt.Fprintf(out, "%s validated.\n", m)
    }
    if report.StopReason != nil {
        fmt.Fprintf(out, "validation stopped at %s: %v\n", report.Stopped, report.StopReason)
    }
}

func (c *ImportCommand) read(path string) (activity.ImportFile, error) {
    file, err := os.Open(path)
    if err != nil {
        return activity.ImportFile{}, err
    }
    defer file.Close()
    data, err := io.ReadAll(io.LimitReader(file, activity.MaxImportSize + 1))
    if err != nil {
        return activity.ImportFile{}, err
    }
    return activity.NewImportFile(data)
}

func (c *ImportCommand) options() (activity.ImportOptions, error) {
    options := activity.NewImportOptions()
    options.Replace, options.Validate = c.settings.Replace(), c.settings.Validate()
    fromText, toText := c.settings.Period()
    var err error
    if fromText != "" {
        if options.From, err = activity.NewMonthFromText(fromText); err != nil {
            return options, NewUsageError("invalid --from")
        }
    }
    if toText != "" {
        if options.To, err = activity.NewMonthFromText(toText); err != nil {
            return options, NewUsageError("invalid --to")
        }
    }
    if options.From.After(options.To) {
        return options, NewUsageError("--from after --to")
    }
    return options, nil
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
