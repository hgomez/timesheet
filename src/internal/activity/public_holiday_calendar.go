/*
 * public_holiday_calendar.go - Copyright (c) 2023-2026 - Olivier Poncet
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

import "time"

// ---------------------------------------------------------------------------
// PublicHolidayCalendar declaration
// ---------------------------------------------------------------------------

type PublicHolidayCalendar struct {
    zone PublicHolidayZone
}

// ---------------------------------------------------------------------------
// PublicHolidayCalendar implementation
// ---------------------------------------------------------------------------

func NewPublicHolidayCalendar(zone PublicHolidayZone) *PublicHolidayCalendar {
    return &PublicHolidayCalendar{zone: zone}
}

func (h *PublicHolidayCalendar) Zone() PublicHolidayZone {
    return h.zone
}

func (h *PublicHolidayCalendar) Easter(year int) Date {
    a := year % 19
    b := year / 100
    c := year % 100
    d := b / 4
    e := b % 4
    f := (b + 8) / 25
    g := (b - f + 1) / 3
    k := ((19 * a) + b - d - g + 15) % 30
    i := c / 4
    l := c % 4
    m := (32 + (2 * e) + (2 * i) - k - l) % 7
    n := (a + (11 * k) + (22 * m)) / 451
    month := (k + m - (7 * n) + 114) / 31
    day := ((k + m - (7 * n) + 114) % 31) + 1
    return NewDate(year, time.Month(month), day)
}

func (h *PublicHolidayCalendar) Name(d Date) (string, bool) {
    if name, holiday := h.national(d); holiday {
        return name, true
    }
    return h.local(d)
}

func (h *PublicHolidayCalendar) national(d Date) (string, bool) {
    easter := h.Easter(d.Year())
    switch {
        case (d.Month().Number() == 1) && (d.Day() == 1):
            return "Jour de l'an", true
        case d == easter.AddDays(1):
            return "Lundi de Pâques", true
        case (d.Month().Number() == 5) && (d.Day() == 1):
            return "Fête du travail", true
        case (d.Month().Number() == 5) && (d.Day() == 8):
            return "Victoire 1945", true
        case d == easter.AddDays(39):
            return "Ascension", true
        case d == easter.AddDays(50):
            return "Lundi de Pentecôte", true
        case (d.Month().Number() == 7) && (d.Day() == 14):
            return "Fête nationale", true
        case (d.Month().Number() == 8) && (d.Day() == 15):
            return "Assomption", true
        case (d.Month().Number() == 11) && (d.Day() == 1):
            return "Toussaint", true
        case (d.Month().Number() == 11) && (d.Day() == 11):
            return "Armistice 1918", true
        case (d.Month().Number() == 12) && (d.Day() == 25):
            return "Noël", true
    }
    return "", false
}

func (h *PublicHolidayCalendar) local(d Date) (string, bool) {
    if h.zone == ZoneAlsaceMoselle {
        switch {
            case d == h.Easter(d.Year()).AddDays(-2):
                return "Vendredi saint", true
            case (d.Month().Number() == 12) && (d.Day() == 26):
                return "Saint-Étienne", true
        }
        return "", false
    }
    if abolition, found := h.abolitionOfSlavery(d.Year()); found && (d == abolition) {
        return "Abolition de l'esclavage", true
    }
    return "", false
}

func (h *PublicHolidayCalendar) abolitionOfSlavery(year int) (Date, bool) {
    switch h.zone {
        case ZoneMayotte:
            return NewDate(year, time.April, 27), true
        case ZoneMartinique:
            return NewDate(year, time.May, 22), true
        case ZoneGuadeloupe:
            return NewDate(year, time.May, 27), true
        case ZoneSaintMartin:
            if year < 2018 {
                return NewDate(year, time.May, 27), true
            }
            return NewDate(year, time.May, 28), true
        case ZoneGuyane:
            return NewDate(year, time.June, 10), true
        case ZoneSaintBarthelemy:
            return NewDate(year, time.October, 9), true
        case ZoneLaReunion:
            return NewDate(year, time.December, 20), true
    }
    return Date{}, false
}

func (h *PublicHolidayCalendar) IsPublicHoliday(d Date) bool {
    _, holiday := h.Name(d)
    return holiday
}

func (h *PublicHolidayCalendar) Kind(d Date) DayKind {
    switch {
        case h.IsPublicHoliday(d):
            return HolidayDay
        case d.IsWeekend():
            return WeekendDay
    }
    return WorkingDay
}

func (h *PublicHolidayCalendar) IsWorkingDay(d Date) bool {
    return h.Kind(d) == WorkingDay
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
