/*
 * bookkeeper_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "strings"
    "testing"
    "time"
)

// ---------------------------------------------------------------------------
// Bookkeeper tests
// ---------------------------------------------------------------------------

func TestPartTimeComplementAccruesSevenTwelfthsAndResets(t *testing.T) {
    holidays := NewPublicHolidayCalendar(ZoneMetropole)
    calendar := authorCalendar(t, "2025-01")
    for m := month(t, "2025-01"); m.Year() == 2025; m = m.Next() {
        fillMonth(t, &calendar, m, holidays)
    }
    for _, text := range []string{"2025-02-11", "2025-06-10", "2025-07-08", "2025-07-09", "2025-10-14", "2025-12-09"} {
        calendar.Days[day(t, text)] = DayEntry{AM: "TPC", PM: "TPC", Comment: "conférence"}
    }
    calendar.Days[day(t, "2025-07-10")] = DayEntry{AM: "TPC", PM: "TT"}
    book := NewBookkeeper(authorConfiguration(), calendar)
    journal := book.Journal(month(t, "2026-01"))
    expectQuantity(t, "TP credited in 2025", journal.Sum("TPC", month(t, "2025-03"), KindCredit), "7/12")
    expectQuantity(t, "TP at 2025-12-31", journal.Balance("TPC", day(t, "2025-12-31")), "0.5")
    expectQuantity(t, "TP after the reset", journal.Balance("TPC", day(t, "2026-01-01")), "0")
    expectQuantity(t, "TP in October", journal.Balance("TPC", day(t, "2025-10-31")), "1/3")
    if journal.HasMovement("RTT", 2025) {
        t.Fatal("a disabled counter must not move")
    }
}

func TestPaidLeaveIsCarriedOverAndRoundedUp(t *testing.T) {
    holidays := NewPublicHolidayCalendar(ZoneMetropole)
    calendar := authorCalendar(t, "2025-01")
    calendar.Settings.Counters["CP"] = CounterSettings{Initial: q(t, "15")}
    calendar.Days[day(t, "2025-03-04")] = DayEntry{AM: "CP", PM: "CP"}
    calendar.Days[day(t, "2025-03-05")] = DayEntry{AM: "CP", PM: "CP"}
    calendar.Days[day(t, "2025-03-06")] = DayEntry{AM: "CP", PM: "CP"}
    for d := day(t, "2025-08-04"); d.Before(day(t, "2025-08-23")); d = d.AddDays(1) {
        if holidays.IsWorkingDay(d) {
            calendar.Days[d] = DayEntry{AM: "CP", PM: "CP"}
        }
    }
    journal := NewBookkeeper(authorConfiguration(), calendar).Journal(month(t, "2025-12"))
    expectQuantity(t, "January credits only the monthly accrual", journal.Sum("CP", month(t, "2025-01"), KindCredit), "2.08")
    expectQuantity(t, "initial balances are adjustments", journal.Sum("CP", month(t, "2025-01"), KindAdjustment), "15")
    expectQuantity(t, "balance on 31 May, rounded up", journal.Balance("CP", day(t, "2025-05-31")), "23")
    expectQuantity(t, "rounding adjustment of May", journal.Sum("CP", month(t, "2025-05"), KindAdjustment), "0.6")
    expectQuantity(t, "balance carried over on 1 June", journal.Balance("CP", day(t, "2025-06-01")), "23")
    expectQuantity(t, "nothing lost in June", journal.Sum("CP", month(t, "2025-06"), KindAdjustment), "0")
    expectQuantity(t, "June credit only", journal.Sum("CP", month(t, "2025-06"), KindCredit), "2.08")
    expectQuantity(t, "balance in August", journal.Balance("CP", day(t, "2025-08-31")), "15.24")
}

func TestAdjustmentOfTheLastDayFollowsTheRounding(t *testing.T) {
    calendar := authorCalendar(t, "2025-01")
    calendar.Adjustments = append(calendar.Adjustments, Adjustment{ID: "a1", Date: day(t, "2025-05-31"), Counter: "CP", Amount: q(t, "-0.28"), Reason: "bulletin"})
    journal := NewBookkeeper(authorConfiguration(), calendar).Journal(month(t, "2025-06"))
    expectQuantity(t, "balance on 31 May, rounded up then adjusted", journal.Balance("CP", day(t, "2025-05-31")), "10.72")
    expectQuantity(t, "balance on 1 June", journal.Balance("CP", day(t, "2025-06-01")), "10.72")
}

func TestDisabledCounterIsNotRounded(t *testing.T) {
    calendar := authorCalendar(t, "2025-01")
    disabled := false
    calendar.Settings.Counters["CP"] = CounterSettings{Enabled: &disabled, Initial: q(t, "3.5")}
    journal := NewBookkeeper(authorConfiguration(), calendar).Journal(month(t, "2025-06"))
    expectQuantity(t, "no rounding of a disabled counter", journal.Balance("CP", day(t, "2025-06-01")), "3.5")
}

func TestForecastStaysWithinTheLastYear(t *testing.T) {
    calendar := NewCalendar(month(t, "2099-06").First())
    book := NewBookkeeper(authorConfiguration(), calendar)
    today := day(t, "2099-07-01")
    cp, _ := authorConfiguration().Counter("CP")
    if position := NewPosition(book.Journal(book.Horizon(today)), cp, today); position.ForecastDate != day(t, "2099-12-31") {
        t.Fatalf("forecast date: %s", position.ForecastDate)
    }
}

func TestHalfDayDebitsHalfADay(t *testing.T) {
    calendar := authorCalendar(t, "2025-01")
    calendar.Settings.Counters["CP"] = CounterSettings{Initial: q(t, "0.25")}
    calendar.Days[day(t, "2025-01-07")] = DayEntry{AM: "CP", PM: "TT"}
    journal := NewBookkeeper(authorConfiguration(), calendar).Journal(month(t, "2025-01"))
    expectQuantity(t, "debit of a half-day", journal.Sum("CP", month(t, "2025-01"), KindDebit), "-0.5")
    expectQuantity(t, "balance", journal.Balance("CP", day(t, "2025-01-07")), "-0.25")
}

func TestIndicatorsMatchTheWorkbookDefinitions(t *testing.T) {
    holidays := NewPublicHolidayCalendar(ZoneMetropole)
    calendar := authorCalendar(t, "2025-01")
    fillMonth(t, &calendar, month(t, "2025-05"), holidays)
    calendar.Days[day(t, "2025-05-06")] = DayEntry{AM: "CP", PM: "CP"}
    calendar.Days[day(t, "2025-05-07")] = DayEntry{AM: "ABS", PM: "ABS", Comment: "ARRÊT MALADIE"}
    calendar.Days[day(t, "2025-05-10")] = DayEntry{AM: "TT", PM: "TT"}
    delete(calendar.Days, day(t, "2025-05-13"))
    book := NewBookkeeper(authorConfiguration(), calendar)
    got := book.Indicators(month(t, "2025-05"), day(t, "2025-11-08"))
    expectQuantity(t, "total", got.Total, "22")
    expectQuantity(t, "holidays", got.Holidays, "3")
    expectQuantity(t, "working days", got.WorkingDays, "19")
    expectQuantity(t, "part time", got.ByCounter["TP"], "4")
    expectQuantity(t, "taken", got.ByCounter["CP"], "1")
    expectQuantity(t, "absences", got.ByCounter["ABS"], "1")
    expectQuantity(t, "missing", got.Missing, "1")
    expectQuantity(t, "worked, the Saturday included", got.ByCounter["TT"], "13")
    sum := got.Missing.Add(got.Upcoming)
    for _, value := range got.ByCounter {
        sum = sum.Add(value)
    }
    expectQuantity(t, "invariant", sum, "31")
}

func TestCountedPartTimeCodeIsTakenFromItsCounter(t *testing.T) {
    holidays := NewPublicHolidayCalendar(ZoneMetropole)
    calendar := authorCalendar(t, "2025-01")
    fillMonth(t, &calendar, month(t, "2025-05"), holidays)
    calendar.Days[day(t, "2025-05-06")] = DayEntry{AM: "TPC", PM: "TPC"}
    book := NewBookkeeper(authorConfiguration(), calendar)
    got := book.Indicators(month(t, "2025-05"), day(t, "2025-11-08"))
    expectQuantity(t, "taken from TPC", got.ByCounter["TPC"], "1")
    expectQuantity(t, "part time keeps the weekly days only", got.ByCounter["TP"], "4")
    if kind := book.DayType(day(t, "2025-05-06"), Morning, day(t, "2025-11-08")); kind != DayCode {
        t.Fatalf("day type %s", kind)
    }
    if problems := book.CheckEntry(day(t, "2025-05-10"), DayEntry{}, DayEntry{AM: "TPC"}); len(problems) != 1 {
        t.Fatalf("a counted code on a weekend: %v", problems)
    }
    expectQuantity(t, "TPC debited", book.Journal(month(t, "2025-05")).Sum("TPC", month(t, "2025-05"), KindDebit), "-1")
}

func TestFedCounterReceivesHalfADayPerHalfDay(t *testing.T) {
    holidays := NewPublicHolidayCalendar(ZoneMetropole)
    calendar := authorCalendar(t, "2025-01")
    fillMonth(t, &calendar, month(t, "2025-05"), holidays)
    fillMonth(t, &calendar, month(t, "2025-12"), holidays)
    calendar.Days[day(t, "2025-05-10")] = DayEntry{AM: "TT", PM: "TT", Comment: "salon"}
    calendar.Days[day(t, "2025-05-07")] = DayEntry{AM: "ABS", PM: "TT", Comment: "rendez-vous"}
    calendar.Days[day(t, "2025-05-11")] = DayEntry{AM: "WE", PM: "WE"}
    book := NewBookkeeper(authorConfiguration(), calendar)
    journal := book.Journal(month(t, "2026-01"))
    worked := book.Indicators(month(t, "2025-05"), day(t, "2025-11-08")).ByCounter["TT"]
    expectQuantity(t, "TT credited with the worked half-days of May, the Saturday included", journal.Sum("TT", month(t, "2025-05"), KindCredit), worked.String())
    expectQuantity(t, "WE credited on every weekend day but the worked Saturday", journal.Sum("WE", month(t, "2025-05"), KindCredit), "8")
    expectQuantity(t, "ABS credited with the morning", journal.Sum("ABS", month(t, "2025-05"), KindCredit), "0.5")
    lines := journal.Month("ABS", month(t, "2025-05"))
    if (len(lines) != 1) || (lines[0].Entry.Label != "Absence, rendez-vous") {
        t.Fatalf("ABS lines: %+v", lines)
    }
    reset := journal.Month("TT", month(t, "2026-01"))
    if (len(reset) == 0) || !strings.HasPrefix(reset[0].Entry.Label, "Ajustement : remise à zéro du solde") || !journal.Balance("TT", day(t, "2026-01-01")).IsZero() {
        t.Fatalf("TT reset: %+v", reset)
    }
    disabled := false
    calendar.Settings.Counters["TT"] = CounterSettings{Enabled: &disabled}
    if !NewBookkeeper(authorConfiguration(), calendar).Journal(month(t, "2025-05")).Sum("TT", month(t, "2025-05"), KindCredit).IsZero() {
        t.Fatal("a disabled counter must receive nothing")
    }
}

func TestHolidaysFollowTheCalendar(t *testing.T) {
    holidays := NewPublicHolidayCalendar(ZoneMetropole)
    calendar := authorCalendar(t, "2025-01")
    fillMonth(t, &calendar, month(t, "2025-05"), holidays)
    calendar.Days[day(t, "2025-05-02")] = DayEntry{AM: "JF", PM: "JF", Comment: "pont"}
    book := NewBookkeeper(authorConfiguration(), calendar)
    got := book.Indicators(month(t, "2025-05"), day(t, "2025-11-08"))
    expectQuantity(t, "legal holidays from the calendar, the bridge day excepted", got.Holidays, "3")
    expectQuantity(t, "the bridge day stays a working day", got.WorkingDays, "19")
    expectQuantity(t, "JF counted on its counter", got.ByCounter["JF"], "4")
    calendar.Days[day(t, "2025-05-01")] = DayEntry{AM: "TT", PM: "TT"}
    worked := NewBookkeeper(authorConfiguration(), calendar).Indicators(month(t, "2025-05"), day(t, "2025-11-08"))
    expectQuantity(t, "a legal holiday worked still leaves the working days", worked.WorkingDays, "19")
    expectQuantity(t, "JF credited, two legal holidays and the bridge day once 1 May is worked", NewBookkeeper(authorConfiguration(), calendar).Journal(month(t, "2025-05")).Sum("JF", month(t, "2025-05"), KindCredit), "3")
    if problems := book.CheckEntry(day(t, "2025-05-03"), DayEntry{}, DayEntry{AM: "JF"}); len(problems) != 0 {
        t.Fatalf("JF on a Saturday: %v", problems)
    }
    if problems := book.CheckEntry(day(t, "2025-05-06"), DayEntry{}, DayEntry{AM: "JF"}); len(problems) != 0 {
        t.Fatalf("JF on a Tuesday: %v", problems)
    }
}

func TestLeaveIsRefusedOnWeekendsAndHolidays(t *testing.T) {
    calendar := authorCalendar(t, "2025-01")
    book := NewBookkeeper(authorConfiguration(), calendar)
    if problems := book.CheckEntry(day(t, "2026-11-11"), DayEntry{}, DayEntry{AM: "CP"}); len(problems) != 1 {
        t.Fatalf("leave on a holiday: %v", problems)
    }
    if problems := book.CheckEntry(day(t, "2026-11-07"), DayEntry{}, DayEntry{AM: "TT", Comment: "salon"}); len(problems) != 0 {
        t.Fatalf("work on a Saturday: %v", problems)
    }
    if problems := book.CheckEntry(day(t, "2026-11-09"), DayEntry{}, DayEntry{AM: "RTT"}); len(problems) != 1 {
        t.Fatalf("code of a disabled counter: %v", problems)
    }
    long := make([]rune, MaxCommentLength + 1)
    for i := range long {
        long[i] = 'x'
    }
    if problems := book.CheckEntry(day(t, "2026-11-09"), DayEntry{}, DayEntry{Comment: string(long)}); len(problems) != 1 {
        t.Fatalf("long comment: %v", problems)
    }
}

func TestPublicHolidaysFollowTheZoneOfTheAccount(t *testing.T) {
    goodFriday := day(t, "2026-04-03")
    metropole, alsace := ZoneMetropole, ZoneAlsaceMoselle
    config := authorConfiguration()
    calendar := authorCalendar(t, "2025-01")
    cases := []struct {
        name     string
        config   PublicHolidayZone
        own      *PublicHolidayZone
        zone     PublicHolidayZone
        refused  int
        holidays string
    }{
        {"default of an older database", "", nil, ZoneMetropole, 0, "1"},
        {"default of the configuration", ZoneAlsaceMoselle, nil, ZoneAlsaceMoselle, 1, "2"},
        {"own zone of the account", ZoneAlsaceMoselle, &metropole, ZoneMetropole, 0, "1"},
        {"own zone over the metropole", ZoneMetropole, &alsace, ZoneAlsaceMoselle, 1, "2"},
    }
    for _, c := range cases {
        config.PublicHolidays = c.config
        calendar.Settings.PublicHolidays = c.own
        book := NewBookkeeper(config, calendar)
        if book.PublicHolidayZone() != c.zone {
            t.Errorf("%s: zone %s, want %s", c.name, book.PublicHolidayZone(), c.zone)
        }
        if problems := book.CheckEntry(goodFriday, DayEntry{}, DayEntry{AM: "CP"}); len(problems) != c.refused {
            t.Errorf("%s: leave on Good Friday: %v", c.name, problems)
        }
        expectQuantity(t, c.name, book.Indicators(month(t, "2026-04"), day(t, "2026-05-01")).Holidays, c.holidays)
    }
}

func TestYearlyCreditAndPositions(t *testing.T) {
    calendar := NewCalendar(month(t, "2026-01").First())
    config := authorConfiguration()
    for i := range config.Counters {
        if config.Counters[i].ID == "RTT" {
            config.Counters[i].Accrual = Accrual{Every: EveryYear, Amount: NewQuantity(10, 1)}
        }
    }
    book := NewBookkeeper(config, calendar)
    today := day(t, "2026-11-08")
    journal := book.Journal(book.Horizon(today))
    rtt, _ := config.Counter("RTT")
    position := NewPosition(journal, rtt, today)
    expectQuantity(t, "RTT today", position.Today, "10")
    if position.ForecastDate != day(t, "2026-12-31") {
        t.Fatalf("RTT forecast date: %+v", position)
    }
    cp, _ := config.Counter("CP")
    position = NewPosition(journal, cp, today)
    expectQuantity(t, "CP today", position.Today, "21.4")
    if position.ForecastDate != NewDate(2027, time.May, 31) {
        t.Fatalf("forecast date: %s", position.ForecastDate)
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
