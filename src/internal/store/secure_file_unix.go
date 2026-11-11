//go:build unix && !aix && (!solaris || illumos)

/*
 * secure_file_unix.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "fmt"
    "io"
    "io/fs"
    "os"
    "path/filepath"
    "syscall"
)

// ---------------------------------------------------------------------------
// SecureFile implementation (unix)
// ---------------------------------------------------------------------------

func (a *SecureFile) Read() ([]byte, bool, error) {
    file, err := os.OpenFile(a.path, os.O_RDONLY | syscall.O_NOFOLLOW | syscall.O_NONBLOCK, 0)
    if errors.Is(err, fs.ErrNotExist) {
        return nil, false, nil
    }
    if errors.Is(err, syscall.ELOOP) {
        return nil, true, fmt.Errorf("%w: %s is a symbolic link", ErrInsecureFile, a.Path())
    }
    if err != nil {
        return nil, true, err
    }
    defer file.Close()
    if err := a.inspectFile(file, a.Path()); err != nil {
        return nil, true, err
    }
    data, err := io.ReadAll(io.LimitReader(file, a.maxSize + 1))
    if err != nil {
        return nil, true, err
    }
    if int64(len(data)) > a.maxSize {
        return nil, true, ErrFileTooLarge
    }
    return data, true, nil
}

func (a *SecureFile) Lock() error {
    if err := a.inspectDirectory(false); err != nil {
        return err
    }
    path := a.lockPath()
    file, err := os.OpenFile(path, os.O_RDWR | os.O_CREATE | syscall.O_NOFOLLOW | syscall.O_NONBLOCK, 0o600)
    if errors.Is(err, syscall.ELOOP) {
        return fmt.Errorf("%w: %s is a symbolic link", ErrInsecureFile, path)
    }
    if err != nil {
        return err
    }
    if err := a.inspectFile(file, path); err != nil {
        file.Close()
        return err
    }
    if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX | syscall.LOCK_NB); err != nil {
        file.Close()
        if errors.Is(err, syscall.EWOULDBLOCK) {
            return fmt.Errorf("%w: %s", ErrLocked, a.Path())
        }
        return err
    }
    a.Unlock()
    a.lock = file
    return nil
}

func (a *SecureFile) inspectFile(file *os.File, path string) error {
    info, err := file.Stat()
    if err != nil {
        return err
    }
    if !info.Mode().IsRegular() {
        return fmt.Errorf("%w: %s is not a regular file", ErrInsecureFile, path)
    }
    if stat, ok := info.Sys().(*syscall.Stat_t); ok && (int(stat.Uid) != os.Getuid()) {
        return fmt.Errorf("%w: %s belongs to another user", ErrInsecureFile, path)
    }
    if (info.Mode().Perm() & 0o077) != 0 {
        return fmt.Errorf("%w: %s is accessible to group or other (expected 0600)", ErrInsecureFile, path)
    }
    return nil
}

func (a *SecureFile) inspectDirectory(warn bool) error {
    directory := filepath.Dir(a.Path())
    info, err := os.Stat(directory)
    if errors.Is(err, fs.ErrNotExist) {
        return fmt.Errorf("%w: %s", ErrMissingDirectory, directory)
    }
    if err != nil {
        return err
    }
    if !info.IsDir() {
        return fmt.Errorf("%w: %s is not a directory", ErrMissingDirectory, directory)
    }
    if stat, ok := info.Sys().(*syscall.Stat_t); ok && (int(stat.Uid) != os.Getuid()) {
        return fmt.Errorf("%w: %s belongs to another user", ErrInsecureDirectory, directory)
    }
    mode := info.Mode()
    if ((mode.Perm() & 0o002) != 0) && ((mode & fs.ModeSticky) == 0) {
        return fmt.Errorf("%w: %s is writable by other", ErrInsecureDirectory, directory)
    }
    if warn && ((mode.Perm() & 0o020) != 0) {
        a.logger.Warn("directory is writable by its group", "directory", directory)
    }
    return nil
}

func (a *SecureFile) Write(data []byte) error {
    if int64(len(data)) > a.maxSize {
        return ErrFileTooLarge
    }
    if err := a.inspectDirectory(false); err != nil {
        return err
    }
    directory := filepath.Dir(a.path)
    temp, err := os.CreateTemp(directory, a.tempPattern())
    if err != nil {
        return err
    }
    name := temp.Name()
    if _, err := temp.Write(data); err != nil {
        temp.Close()
        os.Remove(name)
        return err
    }
    if err := temp.Sync(); err != nil {
        temp.Close()
        os.Remove(name)
        return err
    }
    if err := temp.Close(); err != nil {
        os.Remove(name)
        return err
    }
    if err := os.Rename(name, a.path); err != nil {
        os.Remove(name)
        return err
    }
    a.syncDirectory(directory)
    return nil
}

func (a *SecureFile) syncDirectory(directory string) {
    dir, err := os.Open(directory)
    if err == nil {
        err = dir.Sync()
        dir.Close()
    }
    if err != nil {
        a.logger.Warn("file written but its directory is not synchronised", "error", err.Error())
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
