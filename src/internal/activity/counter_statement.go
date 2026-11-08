/*
 * counter_statement.go - Copyright (c) 2023-2026 - Olivier Poncet
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
// CounterStatement declaration
// ---------------------------------------------------------------------------

type CounterStatement struct {
    Rule        CounterRule
    Month       Month
    Opening     Quantity
    Lines       []StatementLine
    Cancelled   []JournalLine
    Closing     Quantity
    Provisional bool
}

// ---------------------------------------------------------------------------
// CounterStatement implementation
// ---------------------------------------------------------------------------

func NewCounterStatement(journal *Journal, rule CounterRule, m Month) CounterStatement {
    statement := CounterStatement{Rule: rule, Month: m, Opening: journal.Before(rule.ID, m.First())}
    balance := statement.Opening
    for _, line := range journal.Month(rule.ID, m) {
        balance = balance.Add(line.Entry.Amount)
        if line.Provisional {
            statement.Provisional = true
        }
        if line.Reversed {
            statement.Cancelled = append(statement.Cancelled, line)
            continue
        }
        statement.Lines = append(statement.Lines, StatementLine{Line: line, Balance: balance})
    }
    statement.Closing = balance
    return statement
}

func (s CounterStatement) Empty() bool {
    return (len(s.Lines) == 0) && (len(s.Cancelled) == 0) && s.Opening.IsZero()
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
