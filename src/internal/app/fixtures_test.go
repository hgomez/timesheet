/*
 * fixtures_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "context"
    "encoding/json"
    "io"
    "log/slog"
    "os"
    "path/filepath"
    "strings"
    "testing"
    "time"

    "timesheet/internal/account"
    "timesheet/internal/activity"
    "timesheet/internal/store"
)

// ---------------------------------------------------------------------------
// fixtures
// ---------------------------------------------------------------------------

type testRun struct {
    code   int
    stdout string
    stderr string
}

func testEnv(password string) func(string) (string, bool) {
    return func(name string) (string, bool) {
        if (name == PasswordVariable) && (password != "") {
            return password, true
        }
        return "", false
    }
}

func runCommand(t *testing.T, password string, args ...string) testRun {
    t.Helper()
    stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
    application := NewApplication(args, testEnv(password), strings.NewReader(""), stdout, stderr, time.UTC).WithIterations(account.MinIterations).WithClock(activity.NewManualClock(time.Date(2026, time.November, 8, 10, 0, 0, 0, time.UTC)))
    code := application.Run(context.Background())
    return testRun{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

func testConfigPath(directory string) string {
    return filepath.Join(directory, "timesheet.json")
}

func newTestDirectory(t *testing.T) (string, string) {
    t.Helper()
    directory := t.TempDir()
    os.Chmod(directory, 0o700)
    database := filepath.Join(directory, "ts.json")
    configuration := activity.NewDefaultServerConfiguration()
    configuration.Database.Path = activity.NullableText(database)
    if err := NewServerFile(testConfigPath(directory)).Save(configuration); err != nil {
        t.Fatal(err)
    }
    return directory, "--database=" + database
}

func writeTestImport(t *testing.T, path string) {
    t.Helper()
    days := map[string]activity.DayEntry{}
    holidays := activity.NewPublicHolidayCalendar(activity.ZoneMetropole)
    for d := activity.NewDate(2025, time.January, 1); d.Month().Number() == 1; d = d.AddDays(1) {
        switch {
            case d.IsWeekend():
                days[d.String()] = activity.DayEntry{AM: "WE", PM: "WE"}
            case holidays.IsPublicHoliday(d):
                days[d.String()] = activity.DayEntry{AM: "JF", PM: "JF", Comment: "Jour de l’an"}
            case d.Weekday() == time.Monday:
                days[d.String()] = activity.DayEntry{AM: "TP", PM: "TP"}
            case d.Day() == 7:
                days[d.String()] = activity.DayEntry{AM: "CP", PM: "CP", Comment: "Vacances"}
            default:
                days[d.String()] = activity.DayEntry{AM: "TT", PM: "TT"}
        }
    }
    data, err := json.Marshal(map[string]any{"format": activity.ImportFormat, "version": activity.ImportVersion, "settings": map[string]any{"start": "2025-01-01", "counters": map[string]any{"CP": map[string]any{"initial": "2.29"}, "TPC": map[string]any{"enabled": true, "amount": "7/12", "initial": "0"}}}, "days": days, "adjustments": []map[string]string{{"date": "2025-01-31", "counter": "CP", "amount": "1", "reason": "Récupération d'un jour férié"}}})
    if err != nil {
        t.Fatal(err)
    }
    writeTestFile(t, path, string(data))
}

func openTestDatabase(t *testing.T, option string) *store.Database {
    t.Helper()
    database := store.NewDatabase(strings.TrimPrefix(option, "--database="), slog.New(slog.NewTextHandler(io.Discard, nil)), activity.NewDefaultDatabaseSettings().MaxBytes, activity.NewDefaultLimitSettings())
    if err := database.Load(); err != nil {
        t.Fatal(err)
    }
    return database
}

func writeTestFile(t *testing.T, path string, content string) {
    t.Helper()
    if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
        t.Fatal(err)
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
