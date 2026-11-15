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
    if a.Date.IsZero() {
        return []ImportProblem{{Kind: importAdjustmentDate}}
    }
    m := a.Date.Month()
    if !calendar.Tracks(a.Date) || calendar.IsValidated(m) || m.After(horizon) {
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

func (a ImportAdjustment) Adjustment(id string, rule CounterRule) Adjustment {
    return Adjustment{ID: id, Date: a.Date, Counter: rule.ID, Amount: a.Amount, Reason: a.Reason}
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
