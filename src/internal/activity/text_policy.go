/*
 * text_policy.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "unicode"
    "unicode/utf8"
)

// ---------------------------------------------------------------------------
// TextPolicy declaration
// ---------------------------------------------------------------------------

type TextPolicy struct{}

// ---------------------------------------------------------------------------
// TextPolicy implementation
// ---------------------------------------------------------------------------

func NewTextPolicy() *TextPolicy {
    return &TextPolicy{}
}

func (p *TextPolicy) Accepts(text string, max int) bool {
    if !utf8.ValidString(text) || (utf8.RuneCountInString(text) > max) {
        return false
    }
    for _, r := range text {
        joiner := (r == '\u200c') || (r == '\u200d')
        if unicode.IsControl(r) || unicode.In(r, unicode.Zl, unicode.Zp) || (unicode.Is(unicode.Cf, r) && !joiner) {
            return false
        }
    }
    return true
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
