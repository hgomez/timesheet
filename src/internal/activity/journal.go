/*
 * journal.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "slices"
)

// ---------------------------------------------------------------------------
// Journal declaration
// ---------------------------------------------------------------------------

type Journal struct {
    lines    []JournalLine
    counters map[string][]JournalLine
}

// ---------------------------------------------------------------------------
// Journal implementation
// ---------------------------------------------------------------------------

func NewJournal(lines []JournalLine) *Journal {
    reversed := map[string]bool{}
    for _, line := range lines {
        if (line.Entry.Kind == KindReversal) && (line.Entry.Ref != "") {
            reversed[line.Entry.Ref] = true
        }
    }
    marked := make([]JournalLine, len(lines))
    counters := map[string][]JournalLine{}
    for i, line := range lines {
        line.Reversed = (line.Entry.Kind == KindReversal) || ((line.Entry.ID != "") && reversed[line.Entry.ID])
        marked[i] = line
        counters[line.Entry.Counter] = append(counters[line.Entry.Counter], line)
    }
    return &Journal{lines: marked, counters: counters}
}

func (j *Journal) Lines() []JournalLine {
    return slices.Clone(j.lines)
}

func (j *Journal) Balance(counter string, through Date) Quantity {
    balance := NewQuantity(0, 1)
    for _, line := range j.counters[counter] {
        if !line.Entry.Date.After(through) {
            balance = balance.Add(line.Entry.Amount)
        }
    }
    return balance
}

func (j *Journal) Before(counter string, d Date) Quantity {
    return j.Balance(counter, d.AddDays(-1))
}

func (j *Journal) Validated(counter string) Quantity {
    balance := NewQuantity(0, 1)
    for _, line := range j.counters[counter] {
        if !line.Provisional {
            balance = balance.Add(line.Entry.Amount)
        }
    }
    return balance
}

func (j *Journal) Month(counter string, m Month) []JournalLine {
    lines := []JournalLine{}
    for _, line := range j.counters[counter] {
        if line.Entry.Month == m {
            lines = append(lines, line)
        }
    }
    slices.SortStableFunc(lines, j.compare)
    return lines
}

func (j *Journal) compare(a JournalLine, b JournalLine) int {
    return a.Entry.Date.Compare(b.Entry.Date)
}

func (j *Journal) Sum(counter string, m Month, kinds ...EntryKind) Quantity {
    total := NewQuantity(0, 1)
    for _, line := range j.counters[counter] {
        if (line.Entry.Month == m) && !line.Reversed && slices.Contains(kinds, line.Entry.Kind) {
            total = total.Add(line.Entry.Amount)
        }
    }
    return total
}

func (j *Journal) Net(counter string, m Month, excluded ...EntryKind) Quantity {
    total := NewQuantity(0, 1)
    for _, line := range j.counters[counter] {
        if (line.Entry.Month == m) && !line.Reversed && !slices.Contains(excluded, line.Entry.Kind) {
            total = total.Add(line.Entry.Amount)
        }
    }
    return total
}

func (j *Journal) HasMovement(counter string, year int) bool {
    for _, line := range j.counters[counter] {
        if (line.Entry.Date.Year() == year) && !line.Reversed && !line.Entry.Amount.IsZero() {
            return true
        }
    }
    return false
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
