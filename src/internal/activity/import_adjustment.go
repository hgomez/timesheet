/*
 * import_adjustment.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "strconv"
    "strings"
)

// ---------------------------------------------------------------------------
// ImportAdjustment declaration
// ---------------------------------------------------------------------------

type ImportAdjustment struct {
    Date    Date     `json:"date"`
    Counter string   `json:"counter"`
    Amount  Quantity `json:"amount"`
    Reason  string   `json:"reason"`
}

// ---------------------------------------------------------------------------
// ImportAdjustment implementation
// ---------------------------------------------------------------------------

func (a ImportAdjustment) Problems(config Configuration, calendar Calendar, horizon Month) []ImportProblem {
    m := a.Date.Month()
    if a.Date.IsZero() || !calendar.Tracks(a.Date) || calendar.IsValidated(m) || m.After(horizon) {
        return []ImportProblem{{Date: a.Date, Kind: importAdjustmentOutside}}
    }
    rule, known := config.Counter(a.Counter)
    if !known || rule.Archived {
        return []ImportProblem{{Date: a.Date, Kind: importAdjustmentCounter, Detail: strconv.QuoteToASCII(a.Counter)}}
    }
    problems := []ImportProblem{}
    if a.Amount.IsZero() || (a.Amount.Abs().Cmp(NewQuantity(maxImportedDays, 1)) > 0) {
        problems = append(problems, ImportProblem{Date: a.Date, Kind: importAdjustmentQuantity})
    }
    if (a.Reason == "") || (a.Reason != strings.TrimSpace(a.Reason)) || !NewTextPolicy().Accepts(a.Reason, MaxReasonLength) {
        problems = append(problems, ImportProblem{Date: a.Date, Kind: importAdjustmentReason})
    }
    return problems
}

func (a ImportAdjustment) Exists(calendar Calendar, rule CounterRule) bool {
    for _, existing := range calendar.Adjustments {
        if (existing.Date == a.Date) && (existing.Counter == rule.ID) && existing.Amount.Equal(a.Amount) && (existing.Reason == a.Reason) {
            return true
        }
    }
    return false
}

func (a ImportAdjustment) Adjustment(id string, rule CounterRule) Adjustment {
    return Adjustment{ID: id, Date: a.Date, Counter: rule.ID, Amount: a.Amount, Reason: a.Reason}
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
