/*
 * bookkeeper.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "time"
)

// ---------------------------------------------------------------------------
// Bookkeeper declaration
// ---------------------------------------------------------------------------

type Bookkeeper struct {
    config      Configuration
    calendar    Calendar
    holidays    *PublicHolidayCalendar
    codes       map[string]ActivityCode
    adjustments map[Date][]Adjustment
}

type ledgerWriter struct {
    month    Month
    balances map[string]Quantity
    entries  []LedgerEntry
}

// ---------------------------------------------------------------------------
// Bookkeeper implementation
// ---------------------------------------------------------------------------

func NewBookkeeper(config Configuration, calendar Calendar) *Bookkeeper {
    b := &Bookkeeper{config: config, calendar: calendar, codes: map[string]ActivityCode{}, adjustments: map[Date][]Adjustment{}}
    for _, code := range config.Activities {
        b.codes[code.ID] = code
    }
    for _, adjustment := range calendar.Adjustments {
        b.adjustments[adjustment.Date] = append(b.adjustments[adjustment.Date], adjustment)
    }
    b.holidays = NewPublicHolidayCalendar(b.PublicHolidayZone())
    return b
}

func (b *Bookkeeper) Configuration() Configuration {
    return b.config
}

func (b *Bookkeeper) Calendar() Calendar {
    return b.calendar
}

func (b *Bookkeeper) PublicHolidays() *PublicHolidayCalendar {
    return b.holidays
}

func (b *Bookkeeper) PublicHolidayZone() PublicHolidayZone {
    if b.calendar.Settings.PublicHolidays != nil {
        return *b.calendar.Settings.PublicHolidays
    }
    return b.config.PublicHolidayZone()
}

func (b *Bookkeeper) Enabled(rule CounterRule) bool {
    if rule.Archived {
        return false
    }
    if settings, ok := b.calendar.Settings.Counters[rule.ID]; ok && (settings.Enabled != nil) {
        return *settings.Enabled
    }
    return rule.EnabledByDefault
}

func (b *Bookkeeper) Amount(rule CounterRule) Quantity {
    if settings, ok := b.calendar.Settings.Counters[rule.ID]; ok && (settings.Amount != nil) {
        return *settings.Amount
    }
    return rule.Accrual.Amount
}

func (b *Bookkeeper) Horizon(today Date, shown ...Month) Month {
    horizon := today.Month()
    for _, m := range shown {
        if m.After(horizon) {
            horizon = m
        }
    }
    for _, rule := range b.config.Counters {
        if end := rule.PeriodStart.NextAfter(today).AddDays(-1).Month(); end.After(horizon) {
            horizon = end
        }
    }
    if last := NewMonth(MaxYear, time.December); horizon.After(last) {
        horizon = last
    }
    return horizon
}

func (b *Bookkeeper) Journal(through Month) *Journal {
    lines := []JournalLine{}
    balances := map[string]Quantity{}
    for _, entry := range b.calendar.Ledger {
        lines = append(lines, JournalLine{Entry: entry})
        balances[entry.Counter] = balances[entry.Counter].Add(entry.Amount)
    }
    if !b.calendar.HasHistory() {
        for _, entry := range b.openings() {
            lines = append(lines, JournalLine{Entry: entry, Provisional: true})
            balances[entry.Counter] = balances[entry.Counter].Add(entry.Amount)
        }
    }
    for m := b.calendar.FirstOpen(); !m.After(through); m = m.Next() {
        for _, entry := range b.month(m, balances) {
            lines = append(lines, JournalLine{Entry: entry, Provisional: true})
        }
    }
    return NewJournal(lines)
}

func (b *Bookkeeper) openings() []LedgerEntry {
    start := b.calendar.Settings.Start
    entries := []LedgerEntry{}
    for _, rule := range b.config.Counters {
        if initial := b.calendar.Settings.Counters[rule.ID].Initial; !initial.IsZero() {
            entries = append(entries, LedgerEntry{Month: start.Month(), Date: start, Counter: rule.ID, Kind: KindAdjustment, Amount: initial, Label: "Ajustement : solde initial"})
        }
    }
    return entries
}

func (b *Bookkeeper) month(m Month, balances map[string]Quantity) []LedgerEntry {
    writer := &ledgerWriter{month: m, balances: balances}
    trackStart := b.calendar.Settings.Start
    wholeMonth := !trackStart.After(m.First())
    last := m.Last()
    for d := m.First(); !d.After(last); d = d.AddDays(1) {
        if d.Before(trackStart) {
            continue
        }
        for _, rule := range b.config.Counters {
            if !rule.PeriodStart.Matches(d) {
                continue
            }
            if d != trackStart {
                for _, entry := range b.reset(rule, d, balances[rule.ID]) {
                    writer.record(entry)
                }
            }
        }
        debits, credits := b.movements(d)
        for _, rule := range b.config.Counters {
            if entry, found := debits[rule.ID]; found {
                writer.record(entry)
            }
        }
        for _, rule := range b.config.Counters {
            if rule.PeriodStart.Matches(d) && (rule.Accrual.Every == EveryYear) && b.Enabled(rule) && !b.Amount(rule).IsZero() {
                writer.record(LedgerEntry{Date: d, Counter: rule.ID, Kind: KindCredit, Amount: b.Amount(rule), Label: "Acquisition annuelle " + strconv.Itoa(d.Year())})
            }
        }
        for _, rule := range b.config.Counters {
            if entry, found := credits[rule.ID]; found && b.Enabled(rule) {
                writer.record(entry)
            }
        }
        if (d == last) && wholeMonth {
            for _, rule := range b.config.Counters {
                if (rule.Accrual.Every == EveryMonth) && b.Enabled(rule) && !b.Amount(rule).IsZero() {
                    writer.record(LedgerEntry{Date: d, Counter: rule.ID, Kind: KindCredit, Amount: b.Amount(rule), Label: "Acquisition " + m.Label()})
                }
            }
        }
        b.roundAcquired(writer, d, balances)
        for _, adjustment := range b.adjustments[d] {
            writer.record(LedgerEntry{Date: d, Counter: adjustment.Counter, Kind: KindAdjustment, Amount: adjustment.Amount, Label: "Ajustement : " + adjustment.Reason, Ref: adjustment.ID})
        }
    }
    return writer.entries
}

func (b *Bookkeeper) roundAcquired(writer *ledgerWriter, d Date, balances map[string]Quantity) {
    for _, rule := range b.config.Counters {
        if !rule.CarryOver || !b.Enabled(rule) || !rule.PeriodStart.Matches(d.AddDays(1)) {
            continue
        }
        balance := balances[rule.ID]
        if rounding := balance.Ceil().Sub(balance); (balance.Sign() > 0) && !rounding.IsZero() {
            writer.record(LedgerEntry{Date: d, Counter: rule.ID, Kind: KindAdjustment, Amount: rounding, Label: "Ajustement : arrondi du solde reporté à l'entier supérieur"})
        }
    }
}

func (w *ledgerWriter) record(entry LedgerEntry) {
    entry.Month = w.month
    w.balances[entry.Counter] = w.balances[entry.Counter].Add(entry.Amount)
    w.entries = append(w.entries, entry)
}

func (b *Bookkeeper) reset(rule CounterRule, d Date, balance Quantity) []LedgerEntry {
    if rule.CarryOver || balance.IsZero() {
        return nil
    }
    label := "Ajustement : perte du solde au " + d.Numeric()
    if !b.config.Consumed(rule.ID) {
        label = "Ajustement : remise à zéro du solde au " + d.Numeric()
    }
    return []LedgerEntry{{Date: d, Counter: rule.ID, Kind: KindAdjustment, Amount: balance.Neg(), Label: label}}
}

func (b *Bookkeeper) code(id string) (ActivityCode, bool) {
    code, known := b.codes[id]
    return code, known
}

func (b *Bookkeeper) movements(d Date) (map[string]LedgerEntry, map[string]LedgerEntry) {
    entry, found := b.calendar.Days[d]
    if !found || !b.calendar.Tracks(d) {
        return nil, nil
    }
    debits, credits := map[string]LedgerEntry{}, map[string]LedgerEntry{}
    half := NewQuantity(1, 2)
    checked, working := false, false
    for _, part := range entry.Halves() {
        code, known := b.code(entry.Code(part))
        if !known {
            continue
        }
        movements, kind, amount := credits, KindCredit, half
        if code.Consuming() {
            if !checked {
                checked, working = true, b.holidays.IsWorkingDay(d)
            }
            if !working {
                continue
            }
            movements, kind, amount = debits, KindDebit, half.Neg()
        }
        movement, seen := movements[code.Counter()]
        if !seen {
            movement = LedgerEntry{Date: d, Counter: code.Counter(), Kind: kind, Amount: NewQuantity(0, 1), Label: code.Label}
            if entry.Comment != "" {
                movement.Label += ", " + entry.Comment
            }
        }
        movement.Amount = movement.Amount.Add(amount)
        movements[code.Counter()] = movement
    }
    return debits, credits
}

func (b *Bookkeeper) DayType(d Date, half HalfDay, today Date) DayType {
    if !b.calendar.Tracks(d) {
        return DayUntracked
    }
    if _, known := b.code(b.calendar.Days[d].Code(half)); known {
        return DayCode
    }
    if d.Before(today) {
        return DayMissing
    }
    return DayEmpty
}

func (b *Bookkeeper) Indicators(m Month, today Date) MonthlyIndicators {
    half := NewQuantity(1, 2)
    one := NewQuantity(1, 1)
    indicators := MonthlyIndicators{ByCounter: map[string]Quantity{}, ByCode: map[string]Quantity{}}
    for d := m.First(); !d.After(m.Last()); d = d.AddDays(1) {
        weekday := !d.IsWeekend()
        if weekday {
            indicators.Total = indicators.Total.Add(one)
            if b.holidays.IsPublicHoliday(d) {
                indicators.Holidays = indicators.Holidays.Add(one)
            }
        }
        if !b.calendar.Tracks(d) {
            continue
        }
        entry := b.calendar.Days[d]
        for _, h := range entry.Halves() {
            id := entry.Code(h)
            code, known := b.code(id)
            if !known {
                if d.Before(today) {
                    indicators.Missing = indicators.Missing.Add(half)
                } else {
                    indicators.Upcoming = indicators.Upcoming.Add(half)
                }
                continue
            }
            indicators.ByCode[id] = indicators.ByCode[id].Add(half)
            indicators.ByCounter[code.Counter()] = indicators.ByCounter[code.Counter()].Add(half)
        }
    }
    indicators.WorkingDays = indicators.Total.Sub(indicators.Holidays)
    return indicators
}

func (b *Bookkeeper) Anomalies(m Month) []Anomaly {
    anomalies := []Anomaly{}
    for d := m.First(); !d.After(m.Last()); d = d.AddDays(1) {
        if !b.calendar.Tracks(d) {
            continue
        }
        entry := b.calendar.Days[d]
        for _, h := range entry.Halves() {
            id := entry.Code(h)
            if id == "" {
                continue
            }
            code, known := b.code(id)
            if !known {
                anomalies = append(anomalies, Anomaly{Date: d, Message: h.Label() + " : code inconnu « " + id + " »", Blocking: true})
                continue
            }
            if code.Consuming() && !b.holidays.IsWorkingDay(d) {
                anomalies = append(anomalies, Anomaly{Date: d, Message: h.Label() + " : " + b.workingDaysOnly(code), Blocking: true})
                continue
            }
            if rule, ok := b.config.Counter(code.Consumes); ok && code.Consuming() && !b.Enabled(rule) {
                anomalies = append(anomalies, Anomaly{Date: d, Message: h.Label() + " : le compteur « " + rule.Label + " » n'est pas activé", Blocking: false})
            }
        }
    }
    return anomalies
}

func (b *Bookkeeper) Choices(previous string) []ActivityCode {
    choices := []ActivityCode{}
    for _, code := range b.config.Activities {
        if code.ID == previous {
            choices = append(choices, code)
            continue
        }
        if code.Archived {
            continue
        }
        if rule, ok := b.config.Counter(code.Consumes); ok && code.Consuming() && !b.Enabled(rule) {
            continue
        }
        choices = append(choices, code)
    }
    return choices
}

func (b *Bookkeeper) workingDaysOnly(code ActivityCode) string {
    return "« " + code.Label + " » est décompté et ne peut pas être posé un week-end ou un jour férié"
}

func (b *Bookkeeper) CheckEntry(d Date, previous DayEntry, next DayEntry) []string {
    problems := []string{}
    for _, h := range next.Halves() {
        id := next.Code(h)
        if (id == "") || (id == previous.Code(h)) {
            continue
        }
        allowed := false
        for _, code := range b.Choices(previous.Code(h)) {
            if code.ID == id {
                allowed = true
                if code.Consuming() && !b.holidays.IsWorkingDay(d) {
                    problems = append(problems, h.Label() + " : " + b.workingDaysOnly(code))
                }
            }
        }
        if !allowed {
            problems = append(problems, h.Label() + " : code d'activité non disponible")
        }
    }
    if !NewTextPolicy().Accepts(next.Comment, MaxCommentLength) {
        problems = append(problems, "commentaire : une seule ligne de 500 caractères au plus, sans caractère de contrôle")
    }
    return problems
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
