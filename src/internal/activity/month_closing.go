/*
 * month_closing.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "maps"
    "slices"
    "strings"
    "time"
)

// ---------------------------------------------------------------------------
// MonthClosing declaration
// ---------------------------------------------------------------------------

type MonthClosing struct {
    bookkeeper  *Bookkeeper
    identifiers *IdentifierGenerator
}

var (
    ErrMonthNotOver      = errors.New("activity: the month is not over")
    ErrMonthBeforeStart  = errors.New("activity: the month precedes the start of tracking")
    ErrMonthValidated    = errors.New("activity: the month is already validated")
    ErrPreviousMonthOpen = errors.New("activity: the previous month is not validated")
    ErrMonthMissingDays  = errors.New("activity: the month has half-days without entry")
    ErrMonthAnomalies    = errors.New("activity: the month has invalid entries")
    ErrNotLastValidated  = errors.New("activity: only the last validated month can be reopened")
    ErrReasonRequired    = errors.New("activity: a reason is required")
)

// ---------------------------------------------------------------------------
// MonthClosing implementation
// ---------------------------------------------------------------------------

func NewMonthClosing(bookkeeper *Bookkeeper, identifiers *IdentifierGenerator) *MonthClosing {
    return &MonthClosing{bookkeeper: bookkeeper, identifiers: identifiers}
}

func (c *MonthClosing) Check(m Month, today Date) error {
    calendar := c.bookkeeper.Calendar()
    switch {
        case m.Before(calendar.Settings.Start.Month()):
            return ErrMonthBeforeStart
        case calendar.IsValidated(m):
            return ErrMonthValidated
        case !today.After(m.Last()):
            return ErrMonthNotOver
        case (m != calendar.Settings.Start.Month()) && !calendar.IsValidated(m.Previous()):
            return ErrPreviousMonthOpen
    }
    if !c.bookkeeper.Indicators(m, today).Missing.IsZero() {
        return ErrMonthMissingDays
    }
    for _, anomaly := range c.bookkeeper.Anomalies(m) {
        if anomaly.Blocking {
            return ErrMonthAnomalies
        }
    }
    return nil
}

func (c *MonthClosing) Validate(m Month, now time.Time, today Date) (Calendar, error) {
    if err := c.Check(m, today); err != nil {
        return Calendar{}, err
    }
    calendar := c.bookkeeper.Calendar().Clone()
    journal := c.bookkeeper.Journal(m)
    for _, line := range journal.Lines() {
        if !line.Provisional || (line.Entry.Month != m) {
            continue
        }
        id, err := c.identifiers.Next()
        if err != nil {
            return Calendar{}, err
        }
        entry := line.Entry
        entry.ID = id
        entry.Created = now
        calendar.Ledger = append(calendar.Ledger, entry)
    }
    rules := map[string]Quantity{}
    for _, rule := range c.bookkeeper.Configuration().Counters {
        if c.bookkeeper.Enabled(rule) {
            rules[rule.ID] = c.bookkeeper.Amount(rule)
        }
    }
    indicators := c.bookkeeper.Indicators(m, today)
    record := calendar.Months[m]
    record.Status = MonthValidated
    record.ValidatedAt = now
    record.Rules = rules
    record.Indicators = &indicators
    record.History = append(slices.Clone(record.History), MonthEvent{At: now, Action: ActionValidated})
    calendar.Months = maps.Clone(calendar.Months)
    calendar.Months[m] = record
    return calendar, nil
}

func (c *MonthClosing) CheckReopen(m Month) error {
    last, found := c.bookkeeper.Calendar().LastValidated()
    if !found || (last != m) {
        return ErrNotLastValidated
    }
    return nil
}

func (c *MonthClosing) Reopen(m Month, reason string, now time.Time) (Calendar, error) {
    if err := c.CheckReopen(m); err != nil {
        return Calendar{}, err
    }
    reason = strings.TrimSpace(reason)
    if (reason == "") || !NewTextPolicy().Accepts(reason, MaxReasonLength) {
        return Calendar{}, ErrReasonRequired
    }
    calendar := c.bookkeeper.Calendar().Clone()
    journal := NewJournal(c.linesOf(calendar.Ledger))
    for _, line := range journal.Lines() {
        if (line.Entry.Month != m) || line.Reversed {
            continue
        }
        id, err := c.identifiers.Next()
        if err != nil {
            return Calendar{}, err
        }
        original := line.Entry
        calendar.Ledger = append(calendar.Ledger, LedgerEntry{ID: id, Month: m, Date: original.Date, Counter: original.Counter, Kind: KindReversal, Amount: original.Amount.Neg(), Label: "Contre-passation : " + reason, Ref: original.ID, Created: now})
    }
    record := calendar.Months[m]
    record.Status = MonthOpen
    record.ValidatedAt = time.Time{}
    record.Rules = nil
    record.Indicators = nil
    record.History = append(slices.Clone(record.History), MonthEvent{At: now, Action: ActionReopened, Reason: reason})
    calendar.Months = maps.Clone(calendar.Months)
    calendar.Months[m] = record
    return calendar, nil
}

func (c *MonthClosing) linesOf(entries []LedgerEntry) []JournalLine {
    lines := make([]JournalLine, len(entries))
    for i, entry := range entries {
        lines[i] = JournalLine{Entry: entry}
    }
    return lines
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
