/*
 * accrual_period.go - Copyright (c) 2023-2026 - Olivier Poncet
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
// AccrualPeriod declaration
// ---------------------------------------------------------------------------

type AccrualPeriod string

const (
    EveryMonth AccrualPeriod = "month"
    EveryYear  AccrualPeriod = "year"
)

// ---------------------------------------------------------------------------
// AccrualPeriod implementation
// ---------------------------------------------------------------------------

func (p AccrualPeriod) Valid() bool {
    return (p == EveryMonth) || (p == EveryYear)
}

func (p AccrualPeriod) Label() string {
    if p == EveryYear {
        return "par an"
    }
    return "par mois"
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
