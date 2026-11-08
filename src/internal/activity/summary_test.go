/*
 * summary_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "testing"
)

// ---------------------------------------------------------------------------
// Summary tests
// ---------------------------------------------------------------------------

func TestSummaryFollowsTheWorkbookLayout(t *testing.T) {
    holidays := NewHolidayCalendar()
    calendar := authorCalendar(t, "2025-01")
    calendar.Settings.Counters["CP"] = CounterSettings{Initial: q(t, "14.32")}
    for m := month(t, "2025-01"); m.Year() == 2025; m = m.Next() {
        fillMonth(t, &calendar, m, holidays)
    }
    calendar.Days[day(t, "2025-02-11")] = DayEntry{AM: "TPC", PM: "TPC"}
    calendar.Days[day(t, "2025-01-07")] = DayEntry{AM: "CP", PM: "CP"}
    calendar.Days[day(t, "2026-02-10")] = DayEntry{AM: "CP", PM: "CP"}
    book := NewBookkeeper(authorConfiguration(), calendar, holidays)
    today := day(t, "2026-01-15")
    summary := NewSummary(book, book.Journal(month(t, "2026-12")), month(t, "2025-01"), month(t, "2025-12"), today)
    ids := []string{}
    for _, rule := range summary.Counters {
        ids = append(ids, rule.ID)
    }
    if strings.Join(ids, " ") != "TT WE JF CP CSS TP TPC ABS" {
        t.Fatalf("counters: %v", ids)
    }
    expectQuantity(t, "TT credited with the worked days", summary.Totals.Credits["TT"], summary.Totals.Indicators.ByCounter["TT"].String())
    expectQuantity(t, "TP credited with the part-time days", summary.Totals.Credits["TP"], summary.Totals.Indicators.ByCounter["TP"].String())
    expectQuantity(t, "TT before the tracking", summary.Opening["TT"], "0")
    expectQuantity(t, "report CP before the tracking", summary.Opening["CP"], "0")
    expectQuantity(t, "initial balance adjusted in January", summary.Rows[0].Adjustments["CP"], "14.32")
    for _, rule := range summary.Counters {
        balance := summary.Opening[rule.ID]
        for _, row := range summary.Rows {
            balance = balance.Add(row.Credits[rule.ID]).Sub(row.Debits[rule.ID]).Add(row.Adjustments[rule.ID])
            expectQuantity(t, rule.ID + " " + row.Month.String() + " report, credits, debits and adjustments", row.Balances[rule.ID], balance.String())
        }
    }
    following := NewSummary(book, book.Journal(month(t, "2026-12")), month(t, "2026-01"), month(t, "2026-12"), today)
    expectQuantity(t, "TT report of 2026 before its reset", following.Opening["TT"], summary.Totals.Balances["TT"].String())
    expectQuantity(t, "TT reset in January 2026", following.Rows[0].Adjustments["TT"], summary.Totals.Balances["TT"].Neg().String())
    expectQuantity(t, "report TP", summary.Opening["TPC"], "0")
    expectQuantity(t, "cumul CP", summary.Totals.Credits["CP"], "24.96")
    expectQuantity(t, "cumul TP", summary.Totals.Credits["TPC"], "7")
    expectQuantity(t, "TP taken", summary.Totals.Debits["TPC"], "1")
    expectQuantity(t, "TP balance at the end of the year", summary.Totals.Balances["TPC"], "6")
    expectQuantity(t, "total", summary.Totals.Indicators.Total, "261")
    expectQuantity(t, "holidays", summary.Totals.Indicators.Holidays, "10")
    expectQuantity(t, "working days", summary.Totals.Indicators.WorkingDays, "251")
    for _, row := range summary.Rows {
        i := row.Indicators
        sum := i.Missing.Add(i.Upcoming)
        for _, value := range i.ByCounter {
            sum = sum.Add(value)
        }
        expectQuantity(t, row.Month.String() + " invariant", sum, strconv.Itoa(row.Month.Days()))
    }
    if following.Rows[0].State != RowProvisional || following.Rows[1].State != RowFuture {
        t.Fatalf("states: %s %s", following.Rows[0].State, following.Rows[1].State)
    }
    expectQuantity(t, "future months show the forecast credit", following.Rows[1].Credits["CP"], "2.08")
    expectQuantity(t, "future balance keeps planned leave", following.Rows[1].Balances["CP"].Sub(following.Rows[0].Balances["CP"]), "1.08")
    may := NewSummary(book, book.Journal(month(t, "2026-12")), month(t, "2025-05"), month(t, "2025-05"), today)
    if len(may.Rows) != 1 {
        t.Fatalf("a month summary has one row, got %d", len(may.Rows))
    }
    expectQuantity(t, "report of May is the balance of April", may.Opening["CP"], summary.Rows[3].Balances["CP"].String())
    expectQuantity(t, "May balance", may.Rows[0].Balances["CP"], summary.Rows[4].Balances["CP"].String())
    later := NewSummary(book, book.Journal(month(t, "2027-12")), month(t, "2027-01"), month(t, "2027-12"), today)
    for _, rule := range later.Counters {
        expectQuantity(t, rule.ID + " January 2027 equals the journal", later.Rows[0].Balances[rule.ID], book.Journal(month(t, "2027-12")).Balance(rule.ID, day(t, "2027-01-31")).String())
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
