/*
 * today.go - Copyright (c) 2023-2026 - Olivier Poncet
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

package web

import (
    "time"

    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// Today declaration
// ---------------------------------------------------------------------------

type Today struct {
    clock    activity.Clock
    location *time.Location
}

// ---------------------------------------------------------------------------
// Today implementation
// ---------------------------------------------------------------------------

func NewToday(clock activity.Clock, location *time.Location) *Today {
    return &Today{clock: clock, location: location}
}

func (t *Today) Location() *time.Location {
    return t.location
}

func (t *Today) Date() activity.Date {
    return activity.NewDateOf(t.clock.Now().In(t.location))
}

func (t *Today) Year() int {
    return t.Date().Year()
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
