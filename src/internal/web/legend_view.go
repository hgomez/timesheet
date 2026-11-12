/*
 * legend_view.go - Copyright (c) 2023-2026 - Olivier Poncet
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

import "timesheet/internal/activity"

// ---------------------------------------------------------------------------
// LegendView declaration
// ---------------------------------------------------------------------------

type LegendView struct{}

type legendRecord struct {
    Class string
    Label string
}

// ---------------------------------------------------------------------------
// LegendView implementation
// ---------------------------------------------------------------------------

func NewLegendView() *LegendView {
    return &LegendView{}
}

func (v *LegendView) Build(config activity.Configuration) []legendRecord {
    legend := []legendRecord{}
    for _, code := range config.Activities {
        if !code.Archived {
            legend = append(legend, legendRecord{Class: "code-" + code.ID, Label: code.ID + " : " + code.Label})
        }
    }
    return append(legend, legendRecord{Class: "day-" + activity.DayMissing.Slug(), Label: activity.DayMissing.Label()})
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
