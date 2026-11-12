/*
 * dashboard_view.go - Copyright (c) 2023-2026 - Olivier Poncet
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
// DashboardView declaration
// ---------------------------------------------------------------------------

type DashboardView struct {
    positions *PositionView
}

type dashboardRecord struct {
    Owner      string
    Today      string
    TodayLabel string
    Positions  []positionRecord
    Total      dashboardTotalRecord
}

type dashboardTotalRecord struct {
    Since    string
    SinceDay string
    Entered  string
    Unit     string
    Upcoming string
    Missing  string
}

// ---------------------------------------------------------------------------
// DashboardView implementation
// ---------------------------------------------------------------------------

func NewDashboardView() *DashboardView {
    return &DashboardView{positions: NewPositionView()}
}

func (v *DashboardView) Build(ws *Workspace) dashboardRecord {
    book := ws.Book
    record := dashboardRecord{Today: ws.Today.String(), TodayLabel: ws.Today.Numeric(), Positions: v.positions.Build(book, book.Journal(book.Horizon(ws.Today)), ws.Today)}
    if ws.ReadOnly {
        record.Owner = string(ws.Owner.Login)
    }
    record.Total = v.total(ws)
    return record
}

func (v *DashboardView) total(ws *Workspace) dashboardTotalRecord {
    since := activity.NewDate(ws.Today.Year(), time.January, 1)
    if start := ws.Calendar.Settings.Start; start.After(since) {
        since = start
    }
    half := activity.NewQuantity(1, 2)
    entered, upcoming, missing := activity.NewQuantity(0, 1), activity.NewQuantity(0, 1), activity.NewQuantity(0, 1)
    for d := since; d.Year() == ws.Today.Year(); d = d.AddDays(1) {
        for _, part := range (activity.DayEntry{}).Halves() {
            switch ws.Book.DayType(d, part, ws.Today) {
                case activity.DayCode:
                    entered = entered.Add(half)
                    if d.After(ws.Today) {
                        upcoming = upcoming.Add(half)
                    }
                case activity.DayMissing:
                    missing = missing.Add(half)
            }
        }
    }
    unit := "jours saisis"
    if entered.Cmp(activity.NewQuantity(2, 1)) < 0 {
        unit = "jour saisi"
    }
    return dashboardTotalRecord{Since: since.Numeric(), SinceDay: since.String(), Entered: entered.FormatDays(), Unit: unit, Upcoming: upcoming.FormatDays(), Missing: missing.FormatDays()}
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
