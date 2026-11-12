/*
 * counters_view.go - Copyright (c) 2023-2026 - Olivier Poncet
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

package web

import (
    "time"

    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// CountersView declaration
// ---------------------------------------------------------------------------

type CountersView struct {
    statements *StatementView
}

type counterBookRecord struct {
    Label   string
    Rule    string
    Recap   []counterRecapRecord
    Lines   []statementLineRecord
}

type counterRecapRecord struct {
    Month    string
    Path     string
    State    string
    Debited  string
    Credited string
    Adjusted string
    Balance  string
    Negative bool
}

// ---------------------------------------------------------------------------
// CountersView implementation
// ---------------------------------------------------------------------------

func NewCountersView() *CountersView {
    return &CountersView{statements: NewStatementView()}
}

func (v *CountersView) Build(ws *Workspace, journal *activity.Journal, year int) []counterBookRecord {
    books := []counterBookRecord{}
    for _, rule := range ws.Config.Counters {
        if ws.Book.Enabled(rule) || journal.HasMovement(rule.ID, year) {
            books = append(books, v.counter(ws, journal, rule, year))
        }
    }
    return books
}

func (v *CountersView) counter(ws *Workspace, journal *activity.Journal, rule activity.CounterRule, year int) counterBookRecord {
    book := counterBookRecord{Label: rule.Label, Rule: v.describe(ws, rule)}
    for number := time.January; number <= time.December; number++ {
        m := activity.NewMonth(year, number)
        if m.Before(ws.Calendar.Settings.Start.Month()) {
            continue
        }
        balance := journal.Balance(rule.ID, m.Last())
        recap := counterRecapRecord{Month: m.Title(), Path: ws.MonthPath(m), State: v.state(ws, m), Debited: journal.Sum(rule.ID, m, activity.KindDebit).Neg().Format(), Credited: journal.Sum(rule.ID, m, activity.KindCredit).Format(), Adjusted: journal.Sum(rule.ID, m, activity.KindAdjustment).FormatSigned(), Balance: balance.Format(), Negative: balance.Sign() < 0}
        book.Recap = append(book.Recap, recap)
        statement := v.statements.Build(journal, rule, m)
        book.Lines = append(book.Lines, statement.Lines...)
    }
    return book
}

func (v *CountersView) state(ws *Workspace, m activity.Month) string {
    switch {
        case ws.Calendar.IsValidated(m):
            return activity.RowValidated.Label()
        case m.After(ws.Today.Month()):
            return activity.RowFuture.Label()
    }
    return activity.RowProvisional.Label()
}

func (v *CountersView) describe(ws *Workspace, rule activity.CounterRule) string {
    text := ws.Book.Amount(rule).Format() + " j " + rule.Accrual.Every.Label() + ", période au " + rule.PeriodStart.Label()
    if rule.CarryOver {
        return text + ", solde reporté et arrondi à l'entier supérieur"
    }
    return text + ", solde remis à zéro"
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
