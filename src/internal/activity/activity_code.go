/*
 * activity_code.go - Copyright (c) 2023-2026 - Olivier Poncet
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
// ActivityCode declaration
// ---------------------------------------------------------------------------

type ActivityCode struct {
    ID       string `json:"id"`
    Label    string `json:"label"`
    Consumes string `json:"consumes"`
    Feeds    string `json:"feeds"`
    Color    Color  `json:"color"`
    Archived bool   `json:"archived"`
}

// ---------------------------------------------------------------------------
// ActivityCode implementation
// ---------------------------------------------------------------------------

func (c ActivityCode) Matches(other ActivityCode) bool {
    return c.ID == other.ID
}

func (c ActivityCode) Consuming() bool {
    return c.Consumes != ""
}

func (c ActivityCode) Feeding() bool {
    return c.Feeds != ""
}

func (c ActivityCode) Counter() string {
    return c.Consumes + c.Feeds
}

func (c ActivityCode) CounterField() string {
    if c.Consuming() {
        return "consumes"
    }
    return "feeds"
}

func (c ActivityCode) MarshalJSON() ([]byte, error) {
    return json.Marshal(struct {
        ID       string       `json:"id"`
        Label    string       `json:"label"`
        Consumes NullableText `json:"consumes"`
        Feeds    NullableText `json:"feeds"`
        Color    Color        `json:"color"`
        Archived bool         `json:"archived"`
    }{c.ID, c.Label, NullableText(c.Consumes), NullableText(c.Feeds), c.Color, c.Archived})
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
