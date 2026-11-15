/*
 * duration.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "strconv"
    "time"
)

// ---------------------------------------------------------------------------
// Duration declaration
// ---------------------------------------------------------------------------

type Duration time.Duration

type durationUnit struct {
    size   time.Duration
    symbol string
    label  string
}

var ErrInvalidDuration = errors.New("activity: invalid duration (a whole number of milliseconds, such as 500ms, 30s, 5m or 8h)")

// ---------------------------------------------------------------------------
// Duration implementation
// ---------------------------------------------------------------------------

func ParseDuration(text string) (Duration, error) {
    parsed, err := time.ParseDuration(text)
    if (err != nil) || (parsed <= 0) || ((parsed % time.Millisecond) != 0) {
        return 0, ErrInvalidDuration
    }
    return Duration(parsed), nil
}

func (d Duration) Std() time.Duration {
    return time.Duration(d)
}

func (d Duration) String() string {
    count, unit := d.split()
    return strconv.FormatInt(count, 10) + unit.symbol
}

func (d Duration) Label() string {
    count, unit := d.split()
    return strconv.FormatInt(count, 10) + " " + unit.label
}

func (d Duration) units() []durationUnit {
    return []durationUnit{
        {size: time.Hour, symbol: "h", label: "h"},
        {size: time.Minute, symbol: "m", label: "min"},
        {size: time.Second, symbol: "s", label: "s"},
        {size: time.Millisecond, symbol: "ms", label: "ms"},
    }
}

func (d Duration) split() (int64, durationUnit) {
    units := d.units()
    for _, unit := range units {
        if (d.Std() % unit.size) == 0 {
            return int64(d.Std() / unit.size), unit
        }
    }
    last := units[len(units) - 1]
    return int64(d.Std() / last.size), last
}

func (d Duration) MarshalJSON() ([]byte, error) {
    return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(data []byte) error {
    if string(data) == "null" {
        return nil
    }
    text := ""
    if err := json.Unmarshal(data, &text); err != nil {
        return ErrInvalidDuration
    }
    parsed, err := ParseDuration(text)
    if err != nil {
        return err
    }
    *d = parsed
    return nil
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
