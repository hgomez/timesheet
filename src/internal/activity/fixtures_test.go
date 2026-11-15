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

package activity

import (
    "crypto/rand"
    "testing"
    "time"
)

// ---------------------------------------------------------------------------
// fixtures
// ---------------------------------------------------------------------------

func q(t *testing.T, text string) Quantity {
    t.Helper()
    value, err := NewQuantityFromText(text)
    if err != nil {
        t.Fatalf("quantity %q: %v", text, err)
    }
    return value
}

func day(t *testing.T, text string) Date {
    t.Helper()
    d, err := NewDateFromText(text)
    if err != nil {
        t.Fatalf("date %q: %v", text, err)
    }
    return d
}

func month(t *testing.T, text string) Month {
    t.Helper()
    m, err := NewMonthFromText(text)
    if err != nil {
        t.Fatalf("month %q: %v", text, err)
    }
    return m
}

func authorConfiguration() Configuration {
    config := NewDefaultConfiguration()
    return config
}

func authorCalendar(t *testing.T, start string) Calendar {
    t.Helper()
    calendar := NewCalendar(month(t, start).First())
    enabled, disabled := true, false
    amount := q(t, "7/12")
    calendar.Settings.Counters["TPC"] = CounterSettings{Enabled: &enabled, Amount: &amount}
    calendar.Settings.Counters["RTT"] = CounterSettings{Enabled: &disabled}
    return calendar
}

func fillMonth(t *testing.T, calendar *Calendar, m Month, holidays *PublicHolidayCalendar) {
    t.Helper()
    for d := m.First(); !d.After(m.Last()); d = d.AddDays(1) {
        code := "TT"
        switch {
            case d.IsWeekend():
                code = "WE"
            case holidays.IsPublicHoliday(d):
                code = "JF"
            case d.Weekday() == time.Monday:
                code = "TP"
        }
        if _, set := calendar.Days[d]; !set {
            calendar.Days[d] = DayEntry{AM: code, PM: code}
        }
    }
}

func expectQuantity(t *testing.T, what string, got Quantity, want string) {
    t.Helper()
    if !got.Equal(q(t, want)) {
        t.Fatalf("%s: got %s, want %s", what, got, want)
    }
}

func newClosing(config Configuration, calendar Calendar) *MonthClosing {
    return NewMonthClosing(NewBookkeeper(config, calendar), NewIdentifierGenerator(rand.Reader))
}

func fixedNow() time.Time {
    return time.Date(2026, time.November, 8, 10, 0, 0, 0, time.UTC)
}

func testServerConfiguration() ServerConfiguration {
    configuration := NewDefaultServerConfiguration()
    configuration.Database.Path = "/srv/timesheet/timesheet-database.json"
    return configuration
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
