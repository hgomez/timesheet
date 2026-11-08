/*
 * day_entry.go - Copyright (c) 2023-2026 - Olivier Poncet
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

import "encoding/json"

// ---------------------------------------------------------------------------
// DayEntry declaration
// ---------------------------------------------------------------------------

type DayEntry struct {
    AM      string `json:"am"`
    PM      string `json:"pm"`
    Comment string `json:"comment"`
}

const (
    MaxCommentLength = 500
)

// ---------------------------------------------------------------------------
// DayEntry implementation
// ---------------------------------------------------------------------------

func (e DayEntry) Empty() bool {
    return (e.AM == "") && (e.PM == "") && (e.Comment == "")
}

func (e DayEntry) Code(half HalfDay) string {
    if half == Morning {
        return e.AM
    }
    return e.PM
}

func (e DayEntry) Halves() []HalfDay {
    return []HalfDay{Morning, Afternoon}
}

func (e DayEntry) MarshalJSON() ([]byte, error) {
    return json.Marshal(struct {
        AM      NullableText `json:"am"`
        PM      NullableText `json:"pm"`
        Comment NullableText `json:"comment"`
    }{NullableText(e.AM), NullableText(e.PM), NullableText(e.Comment)})
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
