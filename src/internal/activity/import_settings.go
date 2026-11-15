/*
 * import_settings.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "bytes"
    "encoding/json"
    "slices"
)

// ---------------------------------------------------------------------------
// ImportSettings declaration
// ---------------------------------------------------------------------------

type ImportSettings struct {
    Start    *Date                    `json:"start"`
    Counters map[string]ImportCounter `json:"counters"`
    order    []string
}

// ---------------------------------------------------------------------------
// ImportSettings implementation
// ---------------------------------------------------------------------------

func (s ImportSettings) MarshalJSON() ([]byte, error) {
    ids := s.order
    if len(ids) == 0 {
        for id := range s.Counters {
            ids = append(ids, id)
        }
        slices.Sort(ids)
    }
    start, err := json.Marshal(s.Start)
    if err != nil {
        return nil, err
    }
    var data bytes.Buffer
    data.WriteString(`{"start":`)
    data.Write(start)
    data.WriteString(`,"counters":{`)
    for i, id := range ids {
        key, err := json.Marshal(id)
        if err != nil {
            return nil, err
        }
        value, err := json.Marshal(s.Counters[id])
        if err != nil {
            return nil, err
        }
        if i > 0 {
            data.WriteString(",")
        }
        data.Write(key)
        data.WriteString(":")
        data.Write(value)
    }
    data.WriteString("}}")
    return data.Bytes(), nil
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
