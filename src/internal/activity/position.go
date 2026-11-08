/*
 * position.go - Copyright (c) 2023-2026 - Olivier Poncet
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
// Position declaration
// ---------------------------------------------------------------------------

type Position struct {
    Rule         CounterRule
    Today        Quantity
    Validated    Quantity
    ForecastDate Date
    Forecast     Quantity
}

// ---------------------------------------------------------------------------
// Position implementation
// ---------------------------------------------------------------------------

func NewPosition(journal *Journal, rule CounterRule, today Date) Position {
    p := Position{Rule: rule, Today: journal.Balance(rule.ID, today), Validated: journal.Validated(rule.ID), ForecastDate: rule.PeriodStart.NextAfter(today).AddDays(-1)}
    if last := NewMonth(MaxYear, time.December).Last(); p.ForecastDate.After(last) {
        p.ForecastDate = last
    }
    p.Forecast = journal.Balance(rule.ID, p.ForecastDate)
    return p
}

func (p Position) Horizon() Month {
    return p.ForecastDate.Month()
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
