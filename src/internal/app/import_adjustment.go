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

package app

import (
    "strconv"
    "strings"

    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// ImportAdjustment declaration
// ---------------------------------------------------------------------------

type ImportAdjustment struct {
    Date    activity.Date     `json:"date"`
    Counter string            `json:"counter"`
    Amount  activity.Quantity `json:"amount"`
    Reason  string            `json:"reason"`
}

// ---------------------------------------------------------------------------
// ImportAdjustment implementation
// ---------------------------------------------------------------------------

func (a ImportAdjustment) Problems(config activity.Configuration, calendar activity.Calendar, horizon activity.Month) []string {
    prefix := "adjustment of " + a.Date.String() + ": "
    m := a.Date.Month()
    if a.Date.IsZero() || !calendar.Tracks(a.Date) || calendar.IsValidated(m) || m.After(horizon) {
        return []string{prefix + "the date must fall in an open month, from the tracking start"}
    }
    rule, known := config.Counter(a.Counter)
    if !known || rule.Archived {
        return []string{prefix + "counter " + strconv.QuoteToASCII(a.Counter) + " unknown or archived"}
    }
    problems := []string{}
    if a.Amount.IsZero() || (a.Amount.Abs().Cmp(activity.NewQuantity(maxImportedDays, 1)) > 0) {
        problems = append(problems, prefix + "a non-zero quantity between -366 and 366 is expected")
    }
    if (a.Reason == "") || (a.Reason != strings.TrimSpace(a.Reason)) || !activity.NewTextPolicy().Accepts(a.Reason, activity.MaxReasonLength) {
        problems = append(problems, prefix + "a reason is required, one line of 200 characters at most")
    }
    return problems
}

func (a ImportAdjustment) Exists(calendar activity.Calendar, rule activity.CounterRule) bool {
    for _, existing := range calendar.Adjustments {
        if (existing.Date == a.Date) && (existing.Counter == rule.ID) && existing.Amount.Equal(a.Amount) && (existing.Reason == a.Reason) {
            return true
        }
    }
    return false
}

func (a ImportAdjustment) Adjustment(id string, rule activity.CounterRule) activity.Adjustment {
    return activity.Adjustment{ID: id, Date: a.Date, Counter: rule.ID, Amount: a.Amount, Reason: a.Reason}
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
