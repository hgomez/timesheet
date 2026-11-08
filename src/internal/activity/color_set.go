/*
 * color_set.go - Copyright (c) 2023-2026 - Olivier Poncet
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
// ColorSet declaration
// ---------------------------------------------------------------------------

type ColorSet map[string]Color

// ---------------------------------------------------------------------------
// ColorSet implementation
// ---------------------------------------------------------------------------

func (s ColorSet) OfCode(code ActivityCode) Color {
    if c := s[code.ID]; c.IsSet() {
        return c
    }
    return code.Color
}

func (s ColorSet) WithCode(id string, c Color) ColorSet {
    codes := ColorSet{}
    for key, value := range s {
        codes[key] = value
    }
    codes[id] = c
    return codes
}

func (s ColorSet) Invalid() (string, bool) {
    for id, c := range s {
        if !c.Valid() {
            return id, true
        }
    }
    return "", false
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
