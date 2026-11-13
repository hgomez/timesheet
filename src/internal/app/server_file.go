/*
 * server_file.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "io/fs"
    "os"
    "path/filepath"

    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// ServerFile declaration
// ---------------------------------------------------------------------------

type ServerFile struct {
    path string
}

const (
    maxServerFileSize   = 1 << 20
    serverFileMode      = 0o644
    serverDirectoryMode = 0o755
)

// ---------------------------------------------------------------------------
// ServerFile implementation
// ---------------------------------------------------------------------------

func NewServerFile(path string) *ServerFile {
    return &ServerFile{path: path}
}

func (f *ServerFile) Path() string {
    return f.path
}

func (f *ServerFile) Load() (activity.ServerConfiguration, error) {
    file, err := os.Open(f.path)
    if errors.Is(err, fs.ErrNotExist) {
        return activity.NewDefaultServerConfiguration(), nil
    }
    if err != nil {
        return activity.ServerConfiguration{}, err
    }
    defer file.Close()
    data, err := io.ReadAll(io.LimitReader(file, maxServerFileSize + 1))
    if err != nil {
        return activity.ServerConfiguration{}, err
    }
    if len(data) > maxServerFileSize {
        return activity.ServerConfiguration{}, fmt.Errorf("%s: larger than %d MiB", f.path, maxServerFileSize >> 20)
    }
    record := activity.NewOptionalServerConfiguration()
    decoder := json.NewDecoder(bytes.NewReader(data))
    decoder.DisallowUnknownFields()
    if err := decoder.Decode(&record); err != nil {
        return activity.ServerConfiguration{}, fmt.Errorf("%s: %w", f.path, err)
    }
    if _, err := decoder.Token(); err != io.EOF {
        return activity.ServerConfiguration{}, fmt.Errorf("%s: data after the configuration", f.path)
    }
    if err := record.Validate(); err != nil {
        return activity.ServerConfiguration{}, fmt.Errorf("%s: %w", f.path, err)
    }
    return record, nil
}

func (f *ServerFile) Encode(configuration activity.ServerConfiguration) ([]byte, error) {
    var data bytes.Buffer
    encoder := json.NewEncoder(&data)
    encoder.SetEscapeHTML(false)
    encoder.SetIndent("", "  ")
    if err := encoder.Encode(configuration); err != nil {
        return nil, err
    }
    return data.Bytes(), nil
}

func (f *ServerFile) Save(configuration activity.ServerConfiguration) error {
    if err := configuration.Validate(); err != nil {
        return err
    }
    data, err := f.Encode(configuration)
    if err != nil {
        return err
    }
    directory := filepath.Dir(f.path)
    if err := os.MkdirAll(directory, serverDirectoryMode); err != nil {
        return err
    }
    temp, err := os.CreateTemp(directory, "." + filepath.Base(f.path) + ".*.tmp")
    if err != nil {
        return err
    }
    name := temp.Name()
    if _, err := temp.Write(data); err == nil {
        err = temp.Chmod(serverFileMode)
    }
    if err == nil {
        err = temp.Sync()
    }
    if closing := temp.Close(); err == nil {
        err = closing
    }
    if err == nil {
        err = os.Rename(name, f.path)
    }
    if err != nil {
        os.Remove(name)
    }
    return err
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
