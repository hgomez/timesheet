/*
 * position_view.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// PositionView declaration
// ---------------------------------------------------------------------------

type PositionView struct{}

type positionRecord struct {
    Counter      string
    ShortLabel   string
    Today        string
    Validated    string
    Forecast     string
    ForecastDate string
    ForecastDay  string
    Negative     bool
    NegativeLast bool
    NegativeEnd  bool
}

// ---------------------------------------------------------------------------
// PositionView implementation
// ---------------------------------------------------------------------------

func NewPositionView() *PositionView {
    return &PositionView{}
}

func (v *PositionView) Build(book *activity.Bookkeeper, journal *activity.Journal, today activity.Date) []positionRecord {
    records := []positionRecord{}
    for _, rule := range book.Configuration().Counters {
        position := activity.NewPosition(journal, rule, today)
        if !book.Enabled(rule) && position.Today.IsZero() {
            continue
        }
        record := positionRecord{Counter: rule.Label, ShortLabel: rule.ShortLabel, Today: position.Today.Format(), Validated: position.Validated.Format(), Forecast: position.Forecast.Format(), ForecastDate: position.ForecastDate.Numeric(), ForecastDay: position.ForecastDate.String(), Negative: position.Today.Sign() < 0, NegativeLast: position.Validated.Sign() < 0, NegativeEnd: position.Forecast.Sign() < 0}
        records = append(records, record)
    }
    return records
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
