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

package app

import (
    "bytes"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "os"

    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// ImportFile declaration
// ---------------------------------------------------------------------------

type ImportFile struct {
    Format      string                             `json:"format"`
    Version     int                                `json:"version"`
    Settings    *ImportSettings                    `json:"settings,omitempty"`
    Days        map[activity.Date]activity.DayEntry `json:"days"`
    Adjustments []ImportAdjustment                 `json:"adjustments,omitempty"`
}

const (
    ImportFormat  = "timesheet-import"
    ImportVersion = 1
    MaxImportSize = 8 << 20
)

var ErrInvalidImportFile = errors.New("app: invalid import file")

// ---------------------------------------------------------------------------
// ImportFile implementation
// ---------------------------------------------------------------------------

func NewImportFileFromFile(path string) (ImportFile, error) {
    file, err := os.Open(path)
    if err != nil {
        return ImportFile{}, err
    }
    defer file.Close()
    data, err := io.ReadAll(io.LimitReader(file, MaxImportSize + 1))
    if err != nil {
        return ImportFile{}, err
    }
    if len(data) > MaxImportSize {
        return ImportFile{}, fmt.Errorf("%w: larger than %d MiB", ErrInvalidImportFile, MaxImportSize >> 20)
    }
    return NewImportFile(data)
}

func NewImportFile(data []byte) (ImportFile, error) {
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
        imported.Days = map[activity.Date]activity.DayEntry{}
    }
    return imported, nil
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
