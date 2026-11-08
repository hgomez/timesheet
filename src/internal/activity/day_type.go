/*
 * day_type.go - Copyright (c) 2023-2026 - Olivier Poncet
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

import "strings"

// ---------------------------------------------------------------------------
// DayType declaration
// ---------------------------------------------------------------------------

type DayType string

const (
    DayCode      DayType = "code"
    DayMissing   DayType = "missing"
    DayEmpty     DayType = "empty"
    DayUntracked DayType = "untracked"
)

// ---------------------------------------------------------------------------
// DayType implementation
// ---------------------------------------------------------------------------

func (t DayType) Slug() string {
    return strings.ReplaceAll(string(t), "_", "-")
}

func (t DayType) Label() string {
    switch t {
        case DayCode:
            return "Saisi"
        case DayMissing:
            return "Non renseigné"
        case DayUntracked:
            return "Hors suivi"
    }
    return "À venir"
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
