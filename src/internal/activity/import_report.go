/*
 * import_report.go - Copyright (c) 2023-2026 - Olivier Poncet
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
// ImportReport declaration
// ---------------------------------------------------------------------------

type ImportReport struct {
    Calendar    Calendar
    Problems    []ImportProblem
    Warnings    []ImportProblem
    Settings    bool
    Counters    int
    Dates       []Date
    Missing     []ImportMissing
    Adjustments int
    Validated   []Month
    Stopped     Month
    StopReason  error
}

// ---------------------------------------------------------------------------
// ImportReport implementation
// ---------------------------------------------------------------------------

func (r ImportReport) First() Month {
    return r.Dates[0].Month()
}

func (r ImportReport) Last() Month {
    return r.Dates[len(r.Dates) - 1].Month()
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
