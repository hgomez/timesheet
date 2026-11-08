/*
 * entry_kind.go - Copyright (c) 2023-2026 - Olivier Poncet
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
// EntryKind declaration
// ---------------------------------------------------------------------------

type EntryKind string

const (
    KindDebit      EntryKind = "debit"
    KindCredit     EntryKind = "credit"
    KindAdjustment EntryKind = "adjustment"
    KindReversal   EntryKind = "reversal"
)

// ---------------------------------------------------------------------------
// EntryKind implementation
// ---------------------------------------------------------------------------

func (k EntryKind) Valid() bool {
    switch k {
        case KindDebit, KindCredit, KindAdjustment, KindReversal:
            return true
    }
    return false
}

func (k EntryKind) Label() string {
    switch k {
        case KindDebit:
            return "Absence"
        case KindCredit:
            return "Acquisition"
        case KindAdjustment:
            return "Ajustement"
    }
    return "Contre-passation"
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
