/*
 * statement_view.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// StatementView declaration
// ---------------------------------------------------------------------------

type StatementView struct{}

type statementRecord struct {
    Counter         string
    ShortLabel      string
    Opening         string
    OpeningNegative bool
    Lines           []statementLineRecord
    Cancelled       []statementLineRecord
    Closing         string
    ClosingNegative bool
    Provisional     bool
}

type statementLineRecord struct {
    Date        string
    DateLabel   string
    Label       string
    Kind        string
    Debit       string
    Credit      string
    Signed      string
    Balance     string
    Negative    bool
    Provisional bool
}

// ---------------------------------------------------------------------------
// StatementView implementation
// ---------------------------------------------------------------------------

func NewStatementView() *StatementView {
    return &StatementView{}
}

func (v *StatementView) Build(journal *activity.Journal, rule activity.CounterRule, m activity.Month) statementRecord {
    statement := activity.NewCounterStatement(journal, rule, m)
    record := statementRecord{Counter: rule.Label, ShortLabel: rule.ShortLabel, Opening: statement.Opening.Format(), OpeningNegative: statement.Opening.Sign() < 0, Closing: statement.Closing.Format(), ClosingNegative: statement.Closing.Sign() < 0, Provisional: statement.Provisional}
    for _, line := range statement.Lines {
        item := v.line(line.Line, line.Balance.Format())
        item.Negative = line.Balance.Sign() < 0
        record.Lines = append(record.Lines, item)
    }
    for _, line := range statement.Cancelled {
        record.Cancelled = append(record.Cancelled, v.line(line, ""))
    }
    return record
}

func (v *StatementView) line(line activity.JournalLine, balance string) statementLineRecord {
    entry := line.Entry
    record := statementLineRecord{Date: entry.Date.String(), DateLabel: entry.Date.Numeric(), Label: entry.Label, Kind: entry.Kind.Label(), Signed: entry.Amount.FormatSigned(), Balance: balance, Provisional: line.Provisional}
    if entry.Amount.Sign() < 0 {
        record.Debit = entry.Amount.Neg().Format()
    } else {
        record.Credit = entry.Amount.Format()
    }
    return record
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
