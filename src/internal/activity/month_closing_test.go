/*
 * month_closing_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "errors"
    "testing"
)

// ---------------------------------------------------------------------------
// MonthClosing tests
// ---------------------------------------------------------------------------

func TestValidationRequiresAClosedCompleteMonth(t *testing.T) {
    holidays := NewHolidayCalendar()
    config := authorConfiguration()
    calendar := authorCalendar(t, "2026-09")
    today := day(t, "2026-11-08")
    closing := newClosing(config, calendar)
    if err := closing.Check(month(t, "2026-11"), today); !errors.Is(err, ErrMonthNotOver) {
        t.Fatalf("current month: %v", err)
    }
    if err := closing.Check(month(t, "2026-08"), today); !errors.Is(err, ErrMonthBeforeStart) {
        t.Fatalf("before start: %v", err)
    }
    if err := closing.Check(month(t, "2026-10"), today); !errors.Is(err, ErrPreviousMonthOpen) {
        t.Fatalf("previous open: %v", err)
    }
    if err := closing.Check(month(t, "2026-09"), today); !errors.Is(err, ErrMonthMissingDays) {
        t.Fatalf("missing days: %v", err)
    }
    fillMonth(t, &calendar, month(t, "2026-09"), holidays)
    calendar.Days[day(t, "2026-09-05")] = DayEntry{AM: "CP", PM: "WE"}
    if err := newClosing(config, calendar).Check(month(t, "2026-09"), today); !errors.Is(err, ErrMonthAnomalies) {
        t.Fatalf("anomaly: %v", err)
    }
    delete(calendar.Days, day(t, "2026-09-05"))
    if err := newClosing(config, calendar).Check(month(t, "2026-09"), today); !errors.Is(err, ErrMonthMissingDays) {
        t.Fatalf("an empty Saturday must block the validation: %v", err)
    }
    calendar.Days[day(t, "2026-09-05")] = DayEntry{AM: "WE", PM: "WE"}
    if err := newClosing(config, calendar).Check(month(t, "2026-09"), today); err != nil {
        t.Fatalf("complete month: %v", err)
    }
}

func TestTrackingMayStartWithinAMonth(t *testing.T) {
    holidays := NewHolidayCalendar()
    config := authorConfiguration()
    calendar := authorCalendar(t, "2023-11")
    calendar.Settings.Start = day(t, "2023-11-27")
    for d := day(t, "2023-11-27"); !d.After(day(t, "2023-11-30")); d = d.AddDays(1) {
        calendar.Days[d] = DayEntry{AM: "TT", PM: "TT"}
    }
    calendar.Days[day(t, "2023-11-20")] = DayEntry{AM: "CP", PM: "CP"}
    calendar.Days[day(t, "2023-11-25")] = DayEntry{AM: "CP", PM: "CP"}
    fillMonth(t, &calendar, month(t, "2023-12"), holidays)
    today := day(t, "2023-12-05")
    book := NewBookkeeper(config, calendar, holidays)
    expectQuantity(t, "missing before the start", book.Indicators(month(t, "2023-11"), today).Missing, "0")
    expectQuantity(t, "codes before the start", book.Indicators(month(t, "2023-11"), today).ByCode["CP"], "0")
    if anomalies := book.Anomalies(month(t, "2023-11")); len(anomalies) != 0 {
        t.Fatalf("anomalies before the start: %+v", anomalies)
    }
    for _, text := range []string{"2023-11-20", "2023-11-21"} {
        if dayType := book.DayType(day(t, text), Morning, today); dayType != DayUntracked {
            t.Fatalf("%s before the start: %v", text, dayType)
        }
    }
    if err := newClosing(config, calendar).Check(month(t, "2023-11"), today); err != nil {
        t.Fatalf("partial start month: %v", err)
    }
    journal := book.Journal(month(t, "2023-12"))
    expectQuantity(t, "CP credited in the start month", journal.Sum("CP", month(t, "2023-11"), KindCredit), "0")
    expectQuantity(t, "TPC credited in the start month", journal.Sum("TPC", month(t, "2023-11"), KindCredit), "0")
    expectQuantity(t, "CP debited before the start", journal.Sum("CP", month(t, "2023-11"), KindDebit), "0")
    expectQuantity(t, "CP credited the next month", journal.Sum("CP", month(t, "2023-12"), KindCredit), "2.08")
}

func TestValidationFreezesTheProvisionalEntries(t *testing.T) {
    holidays := NewHolidayCalendar()
    config := authorConfiguration()
    calendar := authorCalendar(t, "2026-09")
    calendar.Settings.Counters["CP"] = CounterSettings{Initial: q(t, "18.28")}
    fillMonth(t, &calendar, month(t, "2026-09"), holidays)
    fillMonth(t, &calendar, month(t, "2026-10"), holidays)
    calendar.Days[day(t, "2026-10-13")] = DayEntry{AM: "CP", PM: "CP", Comment: "Vacances"}
    today := day(t, "2026-11-08")
    before := NewBookkeeper(config, calendar, holidays).Journal(month(t, "2026-11"))
    validated, err := newClosing(config, calendar).Validate(month(t, "2026-09"), fixedNow(), today)
    if err != nil {
        t.Fatal(err)
    }
    validated, err = newClosing(config, validated).Validate(month(t, "2026-10"), fixedNow(), today)
    if err != nil {
        t.Fatal(err)
    }
    if !validated.IsValidated(month(t, "2026-10")) || (validated.Months[month(t, "2026-10")].Indicators == nil) {
        t.Fatal("October not validated")
    }
    for _, entry := range validated.Ledger {
        if (entry.ID == "") || entry.Created.IsZero() {
            t.Fatalf("frozen entry without identity: %+v", entry)
        }
    }
    after := NewBookkeeper(config, validated, holidays).Journal(month(t, "2026-11"))
    for _, id := range []string{"CP", "TPC"} {
        if !after.Balance(id, day(t, "2026-10-31")).Equal(before.Balance(id, day(t, "2026-10-31"))) {
            t.Fatalf("%s: frozen balance differs from the provisional one", id)
        }
    }
    expectQuantity(t, "validated CP", after.Validated("CP"), "21.44")
    changed := config.Clone()
    changed.Counters[3].Accrual.Amount = q(t, "3")
    reworked := NewBookkeeper(changed, validated, holidays).Journal(month(t, "2026-11"))
    expectQuantity(t, "rule change spares validated months", reworked.Balance("CP", day(t, "2026-10-31")), "21.44")
    expectQuantity(t, "rule change applies to open months", reworked.Balance("CP", day(t, "2026-11-30")), "24.44")
}

func TestReopenReversesOnlyTheLastValidatedMonth(t *testing.T) {
    holidays := NewHolidayCalendar()
    config := authorConfiguration()
    calendar := authorCalendar(t, "2026-09")
    fillMonth(t, &calendar, month(t, "2026-09"), holidays)
    fillMonth(t, &calendar, month(t, "2026-10"), holidays)
    today := day(t, "2026-11-08")
    calendar, _ = newClosing(config, calendar).Validate(month(t, "2026-09"), fixedNow(), today)
    calendar, _ = newClosing(config, calendar).Validate(month(t, "2026-10"), fixedNow(), today)
    reference := NewBookkeeper(config, calendar, holidays).Journal(month(t, "2026-10")).Balance("TPC", day(t, "2026-10-31"))
    if _, err := newClosing(config, calendar).Reopen(month(t, "2026-09"), "erreur", fixedNow()); !errors.Is(err, ErrNotLastValidated) {
        t.Fatalf("older month: %v", err)
    }
    if _, err := newClosing(config, calendar).Reopen(month(t, "2026-10"), " ", fixedNow()); !errors.Is(err, ErrReasonRequired) {
        t.Fatalf("no reason: %v", err)
    }
    reopened, err := newClosing(config, calendar).Reopen(month(t, "2026-10"), "oubli d'un jour", fixedNow())
    if err != nil {
        t.Fatal(err)
    }
    if reopened.IsValidated(month(t, "2026-10")) || (len(reopened.Ledger) <= len(calendar.Ledger)) {
        t.Fatal("October still validated or no reversal written")
    }
    journal := NewBookkeeper(config, reopened, holidays).Journal(month(t, "2026-10"))
    statement := NewCounterStatement(journal, config.Counters[6], month(t, "2026-10"))
    if (len(statement.Cancelled) == 0) || !statement.Provisional {
        t.Fatalf("statement: %+v", statement)
    }
    revalidated, err := newClosing(config, reopened).Validate(month(t, "2026-10"), fixedNow(), today)
    if err != nil {
        t.Fatal(err)
    }
    again := NewBookkeeper(config, revalidated, holidays).Journal(month(t, "2026-10")).Balance("TPC", day(t, "2026-10-31"))
    if !again.Equal(reference) {
        t.Fatalf("revalidation changed the balance: %s != %s", again, reference)
    }
    summary := NewSummary(NewBookkeeper(config, revalidated, holidays), NewBookkeeper(config, revalidated, holidays).Journal(month(t, "2026-12")), month(t, "2026-01"), month(t, "2026-12"), today)
    expectQuantity(t, "TP credits once in October", summary.Rows[9].Credits["TPC"], "7/12")
}

func TestCalendarCloneSharesNothing(t *testing.T) {
    config := authorConfiguration()
    holidays := NewHolidayCalendar()
    calendar := authorCalendar(t, "2026-09")
    fillMonth(t, &calendar, month(t, "2026-09"), holidays)
    validated, err := newClosing(config, calendar).Validate(month(t, "2026-09"), fixedNow(), day(t, "2026-11-08"))
    if err != nil {
        t.Fatal(err)
    }
    clone := validated.Clone()
    record := clone.Months[month(t, "2026-09")]
    record.Rules["CP"] = q(t, "99")
    record.Indicators.ByCode["TT"] = q(t, "99")
    record.History[0].Action = ActionReopened
    *clone.Settings.Counters["TPC"].Enabled = false
    *clone.Settings.Counters["TPC"].Amount = q(t, "99")
    original := validated.Months[month(t, "2026-09")]
    if original.Rules["CP"].Equal(q(t, "99")) || original.Indicators.ByCode["TT"].Equal(q(t, "99")) || (original.History[0].Action != ActionValidated) || !*validated.Settings.Counters["TPC"].Enabled || !validated.Settings.Counters["TPC"].Amount.Equal(q(t, "7/12")) {
        t.Fatal("the clone shares data with its source")
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
