/*
 * server_settings_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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

package activity

import (
    "bytes"
    "encoding/json"
    "errors"
    "os"
    "path/filepath"
    "reflect"
    "strings"
    "testing"
)

// ---------------------------------------------------------------------------
// ServerSettings tests
// ---------------------------------------------------------------------------

func TestServerFileMirrorsTheDefaultServerSettings(t *testing.T) {
    if err := NewDefaultServerSettings().Validate(); err != nil {
        t.Fatal(err)
    }
    data, err := os.ReadFile(filepath.Join("..", "..", "..", "etc", "timesheet", "timesheet.json"))
    if err != nil {
        t.Fatal(err)
    }
    file := struct {
        Server ServerSettings `json:"server"`
    }{}
    decoder := json.NewDecoder(bytes.NewReader(data))
    decoder.DisallowUnknownFields()
    if err := decoder.Decode(&file); err != nil {
        t.Fatal(err)
    }
    if !reflect.DeepEqual(file.Server, NewDefaultServerSettings()) {
        t.Fatalf("etc/timesheet/timesheet.json differs from NewDefaultServerSettings:\n%+v\n%+v", file.Server, NewDefaultServerSettings())
    }
}

func TestServerSettingsValidationNamesTheField(t *testing.T) {
    cases := map[string]func(*ServerSettings){
        "server.timezone":        func(s *ServerSettings) { s.Timezone = "Mars/Olympus" },
        "server.addr":            func(s *ServerSettings) { s.Addr = "nonsense" },
        "server.origins":         func(s *ServerSettings) { s.Origins = nil },
        "server.origins[1]":      func(s *ServerSettings) { s.Origins = []string{"http://a.test:1", "https://b.test"} },
        "server.trusted_proxies": func(s *ServerSettings) { s.TrustedProxies = []string{"10.0.0.0/8", "proxy.example"} },
    }
    for path, mutate := range cases {
        server := NewDefaultServerSettings().Clone()
        mutate(&server)
        err := server.Validate()
        if !errors.Is(err, ErrInvalidConfiguration) || !strings.Contains(err.Error(), path + ":") {
            t.Errorf("%s: got %v", path, err)
        }
    }
    emptied := NewDefaultServerSettings().Clone()
    emptied.TrustedProxies = nil
    if data, err := json.Marshal(emptied); (err != nil) || !strings.Contains(string(data), `"trusted_proxies":[]`) {
        t.Fatalf("an empty list is written as a list: %s %v", data, err)
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
