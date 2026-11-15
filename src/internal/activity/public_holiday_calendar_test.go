/*
 * public_holiday_calendar_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "slices"
    "testing"
    "time"
)

// ---------------------------------------------------------------------------
// PublicHolidayCalendar tests
// ---------------------------------------------------------------------------

func TestEasterDates(t *testing.T) {
    holidays := NewPublicHolidayCalendar(ZoneMetropole)
    cases := map[int]string{2000: "2000-04-23", 2019: "2019-04-21", 2024: "2024-03-31", 2025: "2025-04-20", 2026: "2026-04-05", 2038: "2038-04-25"}
    for year, want := range cases {
        if got := holidays.Easter(year).String(); got != want {
            t.Errorf("%d: got %s, want %s", year, got, want)
        }
    }
}

func TestNationalPublicHolidays(t *testing.T) {
    holidays := NewPublicHolidayCalendar(ZoneMetropole)
    count := 0
    for d := day(t, "2025-01-01"); d.Year() == 2025; d = d.AddDays(1) {
        if holidays.IsPublicHoliday(d) {
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

func TestPublicHolidaysOfEveryZone(t *testing.T) {
    national := []string{"2026-01-01", "2026-04-06", "2026-05-01", "2026-05-08", "2026-05-14", "2026-05-25", "2026-07-14", "2026-08-15", "2026-11-01", "2026-11-11", "2026-12-25"}
    local := map[PublicHolidayZone][]string{
        ZoneAlsaceMoselle:   {"2026-04-03", "2026-12-26"},
        ZoneGuadeloupe:      {"2026-05-27"},
        ZoneGuyane:          {"2026-06-10"},
        ZoneLaReunion:       {"2026-12-20"},
        ZoneMartinique:      {"2026-05-22"},
        ZoneMayotte:         {"2026-04-27"},
        ZoneSaintBarthelemy: {"2026-10-09"},
        ZoneSaintMartin:     {"2026-05-28"},
    }
    for _, zone := range PublicHolidayZones() {
        want := slices.Sorted(slices.Values(append(slices.Clone(national), local[zone]...)))
        if got := publicHolidaysOf(NewPublicHolidayCalendar(zone), 2026); !slices.Equal(got, want) {
            t.Errorf("%s: got %v, want %v", zone, got, want)
        }
    }
    if got := publicHolidaysOf(NewPublicHolidayCalendar(ZoneSaintMartin), 2017); !slices.Contains(got, "2017-05-27") || slices.Contains(got, "2017-05-28") {
        t.Errorf("Saint-Martin 2017: got %v", got)
    }
    alsace := NewPublicHolidayCalendar(ZoneAlsaceMoselle)
    for text, name := range map[string]string{"2026-04-03": "Vendredi saint", "2026-12-26": "Saint-Étienne"} {
        if got, _ := alsace.Name(day(t, text)); got != name {
            t.Errorf("%s: got %q, want %q", text, got, name)
        }
    }
    if got, _ := NewPublicHolidayCalendar(ZoneLaReunion).Name(day(t, "2026-12-20")); got != "Abolition de l'esclavage" {
        t.Errorf("La Réunion: got %q", got)
    }
}

func TestPublicHolidayZonesAreKnown(t *testing.T) {
    if zones := PublicHolidayZones(); (len(zones) != 13) || (zones[0] != ZoneMetropole) {
        t.Fatalf("zones: %v", zones)
    }
    for _, zone := range PublicHolidayZones() {
        if !zone.Valid() {
            t.Errorf("%s is not valid", zone)
        }
    }
    for _, zone := range []PublicHolidayZone{"", "metropole", "FR-metropole", "fr-corse"} {
        if zone.Valid() {
            t.Errorf("%q is valid", zone)
        }
    }
}

func publicHolidaysOf(holidays *PublicHolidayCalendar, year int) []string {
    dates := []string{}
    for d := NewDate(year, time.January, 1); d.Year() == year; d = d.AddDays(1) {
        if holidays.IsPublicHoliday(d) {
            dates = append(dates, d.String())
        }
    }
    return dates
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
