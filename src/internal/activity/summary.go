/*
 * summary.go - Copyright (c) 2023-2026 - Olivier Poncet
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

// ---------------------------------------------------------------------------
// Summary declaration
// ---------------------------------------------------------------------------

type Summary struct {
    Year     int
    Counters []CounterRule
    Opening  map[string]Quantity
    Rows     []SummaryRow
    Totals   SummaryRow
}

// ---------------------------------------------------------------------------
// Summary implementation
// ---------------------------------------------------------------------------

func NewSummary(bookkeeper *Bookkeeper, journal *Journal, first Month, last Month, today Date) Summary {
    calendar := bookkeeper.Calendar()
    year := first.Year()
    summary := Summary{Year: year, Opening: map[string]Quantity{}}
    for _, rule := range bookkeeper.Configuration().Counters {
        if bookkeeper.Enabled(rule) || journal.HasMovement(rule.ID, year) {
            summary.Counters = append(summary.Counters, rule)
        }
    }
    openingDate := first.First()
    if start := calendar.Settings.Start; start.After(openingDate) {
        openingDate = start
    }
    for _, rule := range summary.Counters {
        summary.Opening[rule.ID] = journal.Before(rule.ID, openingDate)
    }
    totals := SummaryRow{Debits: map[string]Quantity{}, Credits: map[string]Quantity{}, Adjustments: map[string]Quantity{}, Balances: map[string]Quantity{}, Indicators: MonthlyIndicators{ByCounter: map[string]Quantity{}, ByCode: map[string]Quantity{}}}
    current := today.Month()
    for m := first; !m.After(last); m = m.Next() {
        row := SummaryRow{Month: m, Debits: map[string]Quantity{}, Credits: map[string]Quantity{}, Adjustments: map[string]Quantity{}, Balances: map[string]Quantity{}}
        switch {
            case m.Before(calendar.Settings.Start.Month()):
                row.State = RowUntracked
            case calendar.IsValidated(m):
                row.State = RowValidated
            case m.After(current):
                row.State = RowFuture
            default:
                row.State = RowProvisional
        }
        if row.State == RowUntracked {
            summary.Rows = append(summary.Rows, row)
            continue
        }
        row.Indicators = bookkeeper.Indicators(m, today)
        if record := calendar.Months[m]; (row.State == RowValidated) && (record.Indicators != nil) {
            row.Indicators = *record.Indicators
        }
        for _, rule := range summary.Counters {
            row.Debits[rule.ID] = journal.Sum(rule.ID, m, KindDebit).Neg()
            row.Credits[rule.ID] = journal.Sum(rule.ID, m, KindCredit)
            row.Adjustments[rule.ID] = journal.Sum(rule.ID, m, KindAdjustment)
            row.Balances[rule.ID] = journal.Balance(rule.ID, m.Last())
            totals.Debits[rule.ID] = totals.Debits[rule.ID].Add(row.Debits[rule.ID])
            totals.Credits[rule.ID] = totals.Credits[rule.ID].Add(row.Credits[rule.ID])
            totals.Adjustments[rule.ID] = totals.Adjustments[rule.ID].Add(row.Adjustments[rule.ID])
            totals.Balances[rule.ID] = row.Balances[rule.ID]
        }
        summary.accumulate(&totals.Indicators, row.Indicators)
        summary.Rows = append(summary.Rows, row)
    }
    summary.Totals = totals
    return summary
}

func (s Summary) accumulate(total *MonthlyIndicators, row MonthlyIndicators) {
    total.Holidays = total.Holidays.Add(row.Holidays)
    total.Total = total.Total.Add(row.Total)
    total.WorkingDays = total.WorkingDays.Add(row.WorkingDays)
    total.Missing = total.Missing.Add(row.Missing)
    total.Upcoming = total.Upcoming.Add(row.Upcoming)
    for id, value := range row.ByCounter {
        total.ByCounter[id] = total.ByCounter[id].Add(value)
    }
    for id, value := range row.ByCode {
        total.ByCode[id] = total.ByCode[id].Add(value)
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
