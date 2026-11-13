//go:build unix && !aix && (!solaris || illumos)

/*
 * session_file_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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

package store

import (
    "bytes"
    "errors"
    "log/slog"
    "os"
    "path/filepath"
    "strings"
    "testing"
    "time"

    "timesheet/internal/account"
    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// SessionFile tests
// ---------------------------------------------------------------------------

func TestSessionFileRoundTrip(t *testing.T) {
    directory := t.TempDir()
    os.Chmod(directory, 0o700)
    if name := filepath.Base(NewSessionFile(filepath.Join(directory, "timesheet-database.json"), nil).Path()); name != ".timesheet-database.json.sessions" {
        t.Fatalf("default name: %s", name)
    }
    file := NewSessionFile(filepath.Join(directory, "ts.json"), slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
    if filepath.Base(file.Path()) != ".ts.json.sessions" {
        t.Fatalf("path: %s", file.Path())
    }
    if records, err := file.LoadSessions(); (err != nil) || (len(records) != 0) {
        t.Fatalf("missing file: %v %v", records, err)
    }
    now := time.Date(2026, time.November, 11, 21, 0, 0, 0, time.UTC)
    saved := []account.SessionRecord{{Key: strings.Repeat("ab", 32), Login: "oponcet", CSRF: "token", Created: now, Seen: now.Add(time.Minute)}}
    if err := file.SaveSessions(saved); err != nil {
        t.Fatal(err)
    }
    if info, _ := os.Stat(file.Path()); info.Mode().Perm() != 0o600 {
        t.Fatalf("mode: %v", info.Mode().Perm())
    }
    loaded, err := file.LoadSessions()
    if (err != nil) || (len(loaded) != 1) || (loaded[0] != saved[0]) {
        t.Fatalf("loaded: %v %v", loaded, err)
    }
    for name, content := range map[string]string{
        "identifier in clear": `[{"key": "abc", "login": "oponcet", "csrf": "t", "created": "2026-11-11T21:00:00Z", "seen": "2026-11-11T21:00:00Z"}]`,
        "unknown field": `[{"key": "` + strings.Repeat("ab", 32) + `", "id": "x", "login": "oponcet", "csrf": "t", "created": "2026-11-11T21:00:00Z", "seen": "2026-11-11T21:00:00Z"}]`,
        "seen before created": `[{"key": "` + strings.Repeat("ab", 32) + `", "login": "oponcet", "csrf": "t", "created": "2026-11-11T21:00:00Z", "seen": "2026-11-11T20:00:00Z"}]`,
    } {
        os.WriteFile(file.Path(), []byte(content), 0o600)
        if _, err := file.LoadSessions(); !errors.Is(err, ErrMalformedSessions) {
            t.Errorf("%s: %v", name, err)
        }
    }
}

func TestSessionFileHoldsEveryRestorableSession(t *testing.T) {
    directory := t.TempDir()
    os.Chmod(directory, 0o700)
    file := NewSessionFile(filepath.Join(directory, "ts.json"), slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
    now := time.Date(2026, time.November, 11, 21, 0, 0, 123456789, time.FixedZone("CET", 3600))
    records := make([]account.SessionRecord, activity.MaxRestorableSessions)
    for i := range records {
        records[i] = account.SessionRecord{Key: strings.Repeat("ab", 32), Login: account.Login(strings.Repeat("l", 32)), CSRF: strings.Repeat("c", 43), Created: now, Seen: now}
    }
    if err := file.SaveSessions(records); err != nil {
        t.Fatalf("the largest set of sessions must fit in the session file: %v", err)
    }
    if loaded, err := file.LoadSessions(); (err != nil) || (len(loaded) != len(records)) {
        t.Fatalf("got %d %v", len(loaded), err)
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
