/*
 * option_list.go - Copyright (c) 2023-2026 - Olivier Poncet
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

package app

import (
    "errors"
    "strings"
)

// ---------------------------------------------------------------------------
// OptionList declaration
// ---------------------------------------------------------------------------

type OptionList struct {
    values []string
}

// ---------------------------------------------------------------------------
// OptionList implementation
// ---------------------------------------------------------------------------

func NewOptionList() *OptionList {
    return &OptionList{}
}

func (l *OptionList) String() string {
    return strings.Join(l.values, ",")
}

func (l *OptionList) Set(value string) error {
    parts := []string{}
    for _, part := range strings.Split(value, ",") {
        if part = strings.TrimSpace(part); part != "" {
            parts = append(parts, part)
        }
    }
    if len(parts) == 0 {
        return errors.New("no value")
    }
    l.values = append(l.values, parts...)
    return nil
}

func (l *OptionList) Values() []string {
    return append([]string(nil), l.values...)
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
