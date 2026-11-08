/*
 * month.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "errors"
    "strconv"
    "time"
)

// ---------------------------------------------------------------------------
// Month declaration
// ---------------------------------------------------------------------------

type Month struct {
    year  int
    month time.Month
}

const monthLayout = "2006-01"

var ErrInvalidMonth = errors.New("activity: invalid month")

// ---------------------------------------------------------------------------
// Month implementation
// ---------------------------------------------------------------------------

func NewMonth(year int, month time.Month) Month {
    t := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
    return Month{year: t.Year(), month: t.Month()}
}

func NewMonthFromText(text string) (Month, error) {
    m := Month{}
    if err := m.UnmarshalText([]byte(text)); err != nil {
        return Month{}, err
    }
    return m, nil
}

func (m Month) Year() int {
    return m.year
}

func (m Month) Number() int {
    return int(m.month)
}

func (m Month) IsZero() bool {
    return m.year == 0
}

func (m Month) Valid() bool {
    return (m.year >= MinYear) && (m.year <= MaxYear)
}

func (m Month) First() Date {
    return NewDate(m.year, m.month, 1)
}

func (m Month) Last() Date {
    return m.Next().First().AddDays(-1)
}

func (m Month) Days() int {
    return m.Last().Day()
}

func (m Month) Next() Month {
    return NewMonth(m.year, m.month + 1)
}

func (m Month) Previous() Month {
    return NewMonth(m.year, m.month - 1)
}

func (m Month) Compare(other Month) int {
    if m.year != other.year {
        return m.year - other.year
    }
    return int(m.month) - int(other.month)
}

func (m Month) Before(other Month) bool {
    return m.Compare(other) < 0
}

func (m Month) After(other Month) bool {
    return m.Compare(other) > 0
}

func (m Month) Contains(d Date) bool {
    return d.Month() == m
}

func (m Month) String() string {
    return m.First().Time().Format(monthLayout)
}

func (m Month) Name() string {
    names := [...]string{"janvier", "février", "mars", "avril", "mai", "juin", "juillet", "août", "septembre", "octobre", "novembre", "décembre"}
    return names[m.month - 1]
}

func (m Month) Title() string {
    name := []rune(m.Name())
    return string(name[0] - ('a' - 'A')) + string(name[1:])
}

func (m Month) Label() string {
    return m.Name() + " " + strconv.Itoa(m.year)
}

func (m Month) MarshalText() ([]byte, error) {
    return []byte(m.String()), nil
}

func (m *Month) UnmarshalText(data []byte) error {
    parsed, err := time.Parse(monthLayout, string(data))
    if (err != nil) || (parsed.Year() < MinYear) || (parsed.Year() > MaxYear) {
        return ErrInvalidMonth
    }
    *m = NewMonth(parsed.Year(), parsed.Month())
    return nil
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
