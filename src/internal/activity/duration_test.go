/*
 * duration_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "encoding/json"
    "errors"
    "testing"
)

// ---------------------------------------------------------------------------
// Duration tests
// ---------------------------------------------------------------------------

func TestDurationsUseTheirShortestForm(t *testing.T) {
    cases := map[string][2]string{
        "8h":     {"8h", "8 h"},
        "8h0m0s": {"8h", "8 h"},
        "60s":    {"1m", "1 min"},
        "90m":    {"90m", "90 min"},
        "1h30m":  {"90m", "90 min"},
        "1.5s":   {"1500ms", "1500 ms"},
        "8760h":  {"8760h", "8760 h"},
        "100ms":  {"100ms", "100 ms"},
    }
    for text, want := range cases {
        parsed, err := ParseDuration(text)
        if (err != nil) || (parsed.String() != want[0]) || (parsed.Label() != want[1]) {
            t.Errorf("%s: got %q %q %v", text, parsed.String(), parsed.Label(), err)
        }
    }
    for _, text := range []string{"", "0s", "-5s", "1500us", "soon", "5"} {
        if _, err := ParseDuration(text); !errors.Is(err, ErrInvalidDuration) {
            t.Errorf("%q: got %v", text, err)
        }
    }
}

func TestDurationsRoundTripThroughJSON(t *testing.T) {
    record := struct {
        Value Duration `json:"value"`
    }{Value: Duration(0)}
    if err := json.Unmarshal([]byte(`{"value": "1h30m"}`), &record); (err != nil) || (record.Value.String() != "90m") {
        t.Fatalf("got %v %v", record.Value, err)
    }
    if data, err := json.Marshal(record); (err != nil) || (string(data) != `{"value":"90m"}`) {
        t.Fatalf("got %s %v", data, err)
    }
    if err := json.Unmarshal([]byte(`{"value": 30}`), &record); !errors.Is(err, ErrInvalidDuration) {
        t.Fatalf("a number is not a duration: %v", err)
    }
    if err := json.Unmarshal([]byte(`{"value": null}`), &record); (err != nil) || (record.Value.String() != "90m") {
        t.Fatalf("null keeps the value: %v %v", record.Value, err)
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
