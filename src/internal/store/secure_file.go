/*
 * secure_file.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "errors"
    "log/slog"
    "os"
    "path/filepath"
    "strings"

    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// SecureFile declaration
// ---------------------------------------------------------------------------

type SecureFile struct {
    path    string
    logger  *slog.Logger
    maxSize int64
    lock    *os.File
}

var (
    ErrUnsupportedPlatform = errors.New("store: secure files require a Unix system")
    ErrInsecureFile        = errors.New("store: insecure file")
    ErrInsecureDirectory   = errors.New("store: insecure directory")
    ErrMissingDirectory    = errors.New("store: the directory does not exist")
    ErrFileTooLarge        = errors.New("store: file too large")
    ErrLocked              = errors.New("store: the file is used by another process")
)

// ---------------------------------------------------------------------------
// SecureFile implementation
// ---------------------------------------------------------------------------

func NewSecureFile(path string, logger *slog.Logger, maxSize int64) *SecureFile {
    return &SecureFile{path: path, logger: logger, maxSize: maxSize}
}

func (a *SecureFile) Path() string {
    if absolute, err := filepath.Abs(a.path); err == nil {
        return absolute
    }
    return a.path
}

func (a *SecureFile) Prepare() error {
    if err := a.inspectDirectory(true); err != nil {
        return err
    }
    directory := filepath.Dir(a.path)
    entries, err := os.ReadDir(directory)
    if err != nil {
        return nil
    }
    prefix, suffix := filepath.Base(a.path) + ".", activity.TemporarySuffix
    for _, entry := range entries {
        name := entry.Name()
        if !entry.Type().IsRegular() || !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, suffix) || !a.generated(name[len(prefix):max(len(prefix), len(name) - len(suffix))]) {
            continue
        }
        if err := os.Remove(filepath.Join(directory, name)); err == nil {
            a.logger.Warn("orphan temporary file removed", "file", name)
        }
    }
    return nil
}

func (a *SecureFile) generated(middle string) bool {
    for _, digit := range []byte(middle) {
        if (digit < '0') || (digit > '9') {
            return false
        }
    }
    return middle != ""
}

func (a *SecureFile) Unlock() {
    if a.lock != nil {
        a.lock.Close()
        a.lock = nil
    }
}

func (a *SecureFile) lockPath() string {
    return a.path + activity.LockSuffix
}

func (a *SecureFile) tempPattern() string {
    return filepath.Base(a.path) + ".*" + activity.TemporarySuffix
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
