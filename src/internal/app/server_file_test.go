/*
 * server_file_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "os"
    "path/filepath"
    "testing"

    "timesheet/internal/account"
    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// ServerFile tests
// ---------------------------------------------------------------------------

func TestShippedServerFileIsTheDefaultConfiguration(t *testing.T) {
    path := filepath.Join("..", "..", "..", "etc", "timesheet", "timesheet.json")
    data, err := os.ReadFile(path)
    if err != nil {
        t.Fatal(err)
    }
    file := NewServerFile(path)
    defaults, err := file.Encode(activity.NewDefaultServerConfiguration())
    if err != nil {
        t.Fatal(err)
    }
    if string(data) != string(defaults) {
        t.Fatalf("etc/timesheet/timesheet.json differs from the output of config show without file:\n%s\n%s", data, defaults)
    }
    if loaded, err := file.Load(); (err != nil) || (loaded.Database.Path != activity.DefaultDatabasePath) {
        t.Fatalf("got %+v %v", loaded.Database, err)
    }
}

func TestServerFileKeepsTheDefaultsOfTheKeysItLeavesOut(t *testing.T) {
    path := filepath.Join(t.TempDir(), "timesheet.json")
    writeTestFile(t, path, `{"server": {"listen": "127.0.0.1:9000", "origins": ["http://127.0.0.1:9000"], "timezone": "Europe/Paris"}}`)
    loaded, err := NewServerFile(path).Load()
    if err != nil {
        t.Fatal(err)
    }
    if (len(loaded.Server.TrustedProxies) != 0) || (loaded.Database != activity.NewDefaultDatabaseSettings()) || (loaded.Sessions != activity.NewDefaultSessionSettings()) || (loaded.Server.KeepAliveTimeout != activity.NewDefaultServerSettings().KeepAliveTimeout) {
        t.Fatalf("an absent list trusts no proxy, the other absent keys keep their default: %+v", loaded)
    }
}

func TestPasswordIterationsStayWithinTheVerifiedRange(t *testing.T) {
    if (activity.MaxPasswordIterations != account.MaxIterations) || (activity.DefaultPasswordIterations < account.MinIterations) {
        t.Fatal("the iterations of the configuration must stay within the range that the verification accepts")
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
