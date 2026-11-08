/*
 * holiday_calendar_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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

import "testing"

// ---------------------------------------------------------------------------
// HolidayCalendar tests
// ---------------------------------------------------------------------------

func TestEasterDates(t *testing.T) {
    holidays := NewHolidayCalendar()
    cases := map[int]string{2000: "2000-04-23", 2019: "2019-04-21", 2024: "2024-03-31", 2025: "2025-04-20", 2026: "2026-04-05", 2038: "2038-04-25"}
    for year, want := range cases {
        if got := holidays.Easter(year).String(); got != want {
            t.Errorf("%d: got %s, want %s", year, got, want)
        }
    }
}

func TestFrenchHolidays(t *testing.T) {
    holidays := NewHolidayCalendar()
    count := 0
    for d := day(t, "2025-01-01"); d.Year() == 2025; d = d.AddDays(1) {
        if holidays.IsHoliday(d) {
            count++
        }
    }
    if count != 11 {
        t.Fatalf("2025 has %d holidays", count)
    }
    for text, name := range map[string]string{"2025-04-21": "Lundi de Pâques", "2025-05-29": "Ascension", "2025-06-09": "Lundi de Pentecôte", "2026-11-11": "Armistice 1918"} {
        if got, _ := holidays.Name(day(t, text)); got != name {
            t.Errorf("%s: got %q, want %q", text, got, name)
        }
    }
    if holidays.Kind(day(t, "2025-11-01")) != HolidayDay {
        t.Fatal("a holiday on a Saturday is still a holiday")
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
