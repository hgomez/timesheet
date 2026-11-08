/*
 * month_record.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "encoding/json"
    "maps"
    "slices"
    "time"
)

// ---------------------------------------------------------------------------
// MonthRecord declaration
// ---------------------------------------------------------------------------

type MonthRecord struct {
    Status      MonthStatus         `json:"status"`
    ValidatedAt time.Time           `json:"validated_at"`
    Rules       map[string]Quantity `json:"rules"`
    Indicators  *MonthlyIndicators  `json:"indicators"`
    History     []MonthEvent        `json:"history"`
}

// ---------------------------------------------------------------------------
// MonthRecord implementation
// ---------------------------------------------------------------------------

func (r MonthRecord) Clone() MonthRecord {
    clone := r
    clone.Rules = maps.Clone(r.Rules)
    clone.History = slices.Clone(r.History)
    if r.Indicators != nil {
        indicators := *r.Indicators
        indicators.ByCounter = maps.Clone(r.Indicators.ByCounter)
        indicators.ByCode = maps.Clone(r.Indicators.ByCode)
        clone.Indicators = &indicators
    }
    return clone
}

func (r MonthRecord) MarshalJSON() ([]byte, error) {
    return json.Marshal(struct {
        Status      MonthStatus         `json:"status"`
        ValidatedAt NullableTime        `json:"validated_at"`
        Rules       map[string]Quantity `json:"rules"`
        Indicators  *MonthlyIndicators  `json:"indicators"`
        History     []MonthEvent        `json:"history"`
    }{r.Status, NullableTime(r.ValidatedAt), r.Rules, r.Indicators, r.History})
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
