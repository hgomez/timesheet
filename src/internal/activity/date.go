/*
 * date.go - Copyright (c) 2023-2026 - Olivier Poncet
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
// Date declaration
// ---------------------------------------------------------------------------

type Date struct {
    year  int
    month time.Month
    day   int
}

const (
    dateLayout = "2006-01-02"
    MinYear    = 2000
    MaxYear    = 2099
)

var ErrInvalidDate = errors.New("activity: invalid date")

// ---------------------------------------------------------------------------
// Date implementation
// ---------------------------------------------------------------------------

func NewDate(year int, month time.Month, day int) Date {
    return NewDateOf(time.Date(year, month, day, 0, 0, 0, 0, time.UTC))
}

func NewDateOf(t time.Time) Date {
    return Date{year: t.Year(), month: t.Month(), day: t.Day()}
}

func NewDateFromText(text string) (Date, error) {
    d := Date{}
    if err := d.UnmarshalText([]byte(text)); err != nil {
        return Date{}, err
    }
    return d, nil
}

func (d Date) Year() int {
    return d.year
}

func (d Date) Month() Month {
    return Month{year: d.year, month: d.month}
}

func (d Date) Day() int {
    return d.day
}

func (d Date) IsZero() bool {
    return d.year == 0
}

func (d Date) Time() time.Time {
    return time.Date(d.year, d.month, d.day, 0, 0, 0, 0, time.UTC)
}

func (d Date) Weekday() time.Weekday {
    return d.Time().Weekday()
}

func (d Date) IsWeekend() bool {
    weekday := d.Weekday()
    return (weekday == time.Saturday) || (weekday == time.Sunday)
}

func (d Date) AddDays(days int) Date {
    return NewDateOf(d.Time().AddDate(0, 0, days))
}

func (d Date) Compare(other Date) int {
    if d.year != other.year {
        return d.year - other.year
    }
    if d.month != other.month {
        return int(d.month) - int(other.month)
    }
    return d.day - other.day
}

func (d Date) Before(other Date) bool {
    return d.Compare(other) < 0
}

func (d Date) After(other Date) bool {
    return d.Compare(other) > 0
}

func (d Date) String() string {
    return d.Time().Format(dateLayout)
}

func (d Date) WeekdayName() string {
    switch d.Weekday() {
        case time.Monday:
            return "lundi"
        case time.Tuesday:
            return "mardi"
        case time.Wednesday:
            return "mercredi"
        case time.Thursday:
            return "jeudi"
        case time.Friday:
            return "vendredi"
        case time.Saturday:
            return "samedi"
    }
    return "dimanche"
}

func (d Date) Label() string {
    return d.WeekdayName() + " " + d.Numeric()
}

func (d Date) Short() string {
    return strconv.Itoa(d.day) + " " + d.Month().Name()
}

func (d Date) Numeric() string {
    return d.Time().Format("02/01/2006")
}

func (d Date) MarshalText() ([]byte, error) {
    return []byte(d.String()), nil
}

func (d *Date) UnmarshalText(data []byte) error {
    parsed, err := time.Parse(dateLayout, string(data))
    if (err != nil) || (parsed.Year() < MinYear) || (parsed.Year() > MaxYear) {
        return ErrInvalidDate
    }
    *d = NewDateOf(parsed)
    return nil
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
