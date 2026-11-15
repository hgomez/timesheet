/*
 * import_file.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "errors"
    "fmt"
    "io"
    "slices"
)

// ---------------------------------------------------------------------------
// ImportFile declaration
// ---------------------------------------------------------------------------

type ImportFile struct {
    Format      string             `json:"format"`
    Version     int                `json:"version"`
    Settings    *ImportSettings    `json:"settings,omitempty"`
    Days        map[Date]DayEntry  `json:"days"`
    Adjustments []ImportAdjustment `json:"adjustments,omitempty"`
}

const (
    ImportFormat  = "timesheet-import"
    ImportVersion = 1
    MaxImportSize = 8 << 20
)

var ErrInvalidImportFile = errors.New("activity: invalid import file")

// ---------------------------------------------------------------------------
// ImportFile implementation
// ---------------------------------------------------------------------------

func NewImportFile(data []byte) (ImportFile, error) {
    if len(data) > MaxImportSize {
        return ImportFile{}, fmt.Errorf("%w: larger than %d MiB", ErrInvalidImportFile, MaxImportSize >> 20)
    }
    imported := ImportFile{}
    decoder := json.NewDecoder(bytes.NewReader(data))
    decoder.DisallowUnknownFields()
    if err := decoder.Decode(&imported); err != nil {
        return ImportFile{}, fmt.Errorf("%w: %v", ErrInvalidImportFile, err)
    }
    if _, err := decoder.Token(); err != io.EOF {
        return ImportFile{}, fmt.Errorf("%w: data after the object", ErrInvalidImportFile)
    }
    if (imported.Format != ImportFormat) || (imported.Version != ImportVersion) {
        return ImportFile{}, fmt.Errorf("%w: format %q version %d expected", ErrInvalidImportFile, ImportFormat, ImportVersion)
    }
    if imported.Days == nil {
        imported.Days = map[Date]DayEntry{}
    }
    return imported, nil
}

func NewImportFileOf(calendar Calendar) ImportFile {
    start := calendar.Settings.Start
    settings := &ImportSettings{Start: &start, Counters: map[string]ImportCounter{}}
    for id, stored := range calendar.Settings.Counters {
        initial := stored.Initial
        settings.Counters[id] = ImportCounter{Enabled: stored.Enabled, Amount: stored.Amount, Initial: &initial}
    }
    days := map[Date]DayEntry{}
    for d, entry := range calendar.Days {
        days[d] = entry
    }
    adjustments := []ImportAdjustment{}
    for _, adjustment := range calendar.Adjustments {
        adjustments = append(adjustments, ImportAdjustment{Date: adjustment.Date, Counter: adjustment.Counter, Amount: adjustment.Amount, Reason: adjustment.Reason})
    }
    slices.SortStableFunc(adjustments, func(a ImportAdjustment, b ImportAdjustment) int {
        return a.Date.Compare(b.Date)
    })
    return ImportFile{Format: ImportFormat, Version: ImportVersion, Settings: settings, Days: days, Adjustments: adjustments}
}

func (f ImportFile) Encode() ([]byte, error) {
    var data bytes.Buffer
    encoder := json.NewEncoder(&data)
    encoder.SetEscapeHTML(false)
    encoder.SetIndent("", "  ")
    if err := encoder.Encode(f); err != nil {
        return nil, err
    }
    return data.Bytes(), nil
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
