/*
 * session_file.go - Copyright (c) 2023-2026 - Olivier Poncet
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

package store

import (
    "bytes"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "log/slog"
    "regexp"

    "timesheet/internal/account"
    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// SessionFile declaration
// ---------------------------------------------------------------------------

type SessionFile struct {
    file *SecureFile
}

const MaxSessionFileSize = 8 << 20

var ErrMalformedSessions = errors.New("store: malformed session file")

// ---------------------------------------------------------------------------
// SessionFile implementation
// ---------------------------------------------------------------------------

func NewSessionFile(database string, logger *slog.Logger) *SessionFile {
    return &SessionFile{file: NewSecureFile(database + activity.SessionsSuffix, logger, MaxSessionFileSize)}
}

func (f *SessionFile) Path() string {
    return f.file.Path()
}

func (f *SessionFile) Prepare() error {
    return f.file.Prepare()
}

func (f *SessionFile) LoadSessions() ([]account.SessionRecord, error) {
    data, found, err := f.file.Read()
    if err != nil {
        return nil, err
    }
    records := []account.SessionRecord{}
    if !found {
        return records, nil
    }
    decoder := json.NewDecoder(bytes.NewReader(data))
    decoder.DisallowUnknownFields()
    if err := decoder.Decode(&records); err != nil {
        return nil, fmt.Errorf("%w: %v", ErrMalformedSessions, err)
    }
    if _, err := decoder.Token(); err != io.EOF {
        return nil, fmt.Errorf("%w: data after the list", ErrMalformedSessions)
    }
    key := regexp.MustCompile(`^[0-9a-f]{64}$`)
    for i, record := range records {
        if !key.MatchString(record.Key) || !record.Login.Valid() || (record.CSRF == "") || record.Created.IsZero() || record.Seen.Before(record.Created) {
            return nil, fmt.Errorf("%w: sessions[%d]: invalid session", ErrMalformedSessions, i)
        }
    }
    return records, nil
}

func (f *SessionFile) SaveSessions(records []account.SessionRecord) error {
    if records == nil {
        records = []account.SessionRecord{}
    }
    data, err := json.MarshalIndent(records, "", "  ")
    if err != nil {
        return err
    }
    return f.file.Write(append(data, '\n'))
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
