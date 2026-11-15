//go:build unix && !aix && (!solaris || illumos)

/*
 * secure_file_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "errors"
    "log/slog"
    "os"
    "path/filepath"
    "strings"
    "testing"
)

// ---------------------------------------------------------------------------
// SecureFile tests
// ---------------------------------------------------------------------------

const testMaxBytes = 64 << 20

func TestSecureFileLiteralNamesAndLateChecks(t *testing.T) {
    directory := t.TempDir()
    unrelated := filepath.Join(directory, "axb.json.keep.tmp")
    os.WriteFile(unrelated, []byte("x"), 0o600)
    file := NewSecureFile(filepath.Join(directory, "a*b.json"), slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)), testMaxBytes)
    if err := file.Prepare(); err != nil {
        t.Fatal(err)
    }
    if _, err := os.Stat(unrelated); err != nil {
        t.Fatalf("unrelated file removed by a glob match")
    }
    os.Chmod(directory, 0o770)
    logs := &bytes.Buffer{}
    shared := NewSecureFile(filepath.Join(directory, "shared.json"), slog.New(slog.NewTextHandler(logs, nil)), testMaxBytes)
    shared.Prepare()
    for i := 0; i < 3; i++ {
        if err := shared.Write([]byte("{}")); err != nil {
            t.Fatal(err)
        }
    }
    if count := strings.Count(logs.String(), "writable by its group"); count != 1 {
        t.Fatalf("%d group-writable warnings, want 1", count)
    }
    os.Chmod(directory, 0o777)
    defer os.Chmod(directory, 0o700)
    if err := file.Write([]byte("{}")); !errors.Is(err, ErrInsecureDirectory) {
        t.Fatalf("write in a world-writable directory: %v", err)
    }
}

func TestSecureFileLock(t *testing.T) {
    directory := t.TempDir()
    logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
    first := NewSecureFile(filepath.Join(directory, "data.json"), logger, testMaxBytes)
    second := NewSecureFile(filepath.Join(directory, "data.json"), logger, testMaxBytes)
    if err := first.Lock(); err != nil {
        t.Fatal(err)
    }
    if err := second.Lock(); !errors.Is(err, ErrLocked) {
        t.Fatalf("second lock: %v", err)
    }
    first.Unlock()
    if err := second.Lock(); err != nil {
        t.Fatalf("lock after release: %v", err)
    }
    second.Unlock()
    info, err := os.Stat(filepath.Join(directory, "data.json.lock"))
    if (err != nil) || (info.Mode().Perm() != 0o600) {
        t.Fatalf("lock file: %v", err)
    }
}

func TestSecureFileRemovesOnlyItsOwnTemporaryFiles(t *testing.T) {
    directory := t.TempDir()
    orphan := filepath.Join(directory, "db.json.4242.tmp")
    foreign := []string{filepath.Join(directory, "db.json.bak.4242.tmp"), filepath.Join(directory, "db.json.notes.tmp"), filepath.Join(directory, "db.json..tmp")}
    for _, name := range append(foreign, orphan) {
        os.WriteFile(name, []byte("x"), 0o600)
    }
    file := NewSecureFile(filepath.Join(directory, "db.json"), slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)), testMaxBytes)
    if err := file.Prepare(); err != nil {
        t.Fatal(err)
    }
    if _, err := os.Stat(orphan); err == nil {
        t.Fatal("orphan temporary file kept")
    }
    for _, name := range foreign {
        if _, err := os.Stat(name); err != nil {
            t.Errorf("%s removed", filepath.Base(name))
        }
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
