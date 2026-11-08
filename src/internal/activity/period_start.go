/*
 * period_start.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "fmt"
    "time"
)

// ---------------------------------------------------------------------------
// PeriodStart declaration
// ---------------------------------------------------------------------------

type PeriodStart struct {
    month time.Month
    day   int
}

var ErrInvalidPeriodStart = errors.New("activity: invalid period start (MM-DD, February 29 excluded)")

// ---------------------------------------------------------------------------
// PeriodStart implementation
// ---------------------------------------------------------------------------

func NewPeriodStart(month time.Month, day int) PeriodStart {
    return PeriodStart{month: month, day: day}
}

func (p PeriodStart) Valid() bool {
    if (p.month < time.January) || (p.month > time.December) || (p.day < 1) || ((p.month == time.February) && (p.day > 28)) {
        return false
    }
    return p.day <= NewMonth(2001, p.month).Days()
}

func (p PeriodStart) In(year int) Date {
    return NewDate(year, p.month, p.day)
}

func (p PeriodStart) Matches(d Date) bool {
    return (d.Month().Number() == int(p.month)) && (d.Day() == p.day)
}

func (p PeriodStart) NextAfter(d Date) Date {
    candidate := p.In(d.Year())
    if !candidate.After(d) {
        candidate = p.In(d.Year() + 1)
    }
    return candidate
}

func (p PeriodStart) String() string {
    return fmt.Sprintf("%02d-%02d", int(p.month), p.day)
}

func (p PeriodStart) Label() string {
    return fmt.Sprintf("%02d/%02d", p.day, int(p.month))
}

func (p PeriodStart) MarshalText() ([]byte, error) {
    return []byte(p.String()), nil
}

func (p *PeriodStart) UnmarshalText(data []byte) error {
    parsed, err := time.Parse("01-02", string(data))
    if err != nil {
        return ErrInvalidPeriodStart
    }
    candidate := PeriodStart{month: parsed.Month(), day: parsed.Day()}
    if !candidate.Valid() {
        return ErrInvalidPeriodStart
    }
    *p = candidate
    return nil
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
