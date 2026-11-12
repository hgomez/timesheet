//go:build linux

/*
 * password_reader_linux.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "os"
    "os/signal"
    "syscall"
    "unsafe"
)

// ---------------------------------------------------------------------------
// PasswordReader implementation (linux)
// ---------------------------------------------------------------------------

func (p *PasswordReader) hideEcho(file *os.File) (func(), error) {
    var state syscall.Termios
    if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), syscall.TCGETS, uintptr(unsafe.Pointer(&state))); errno != 0 {
        return nil, errno
    }
    hidden := state
    hidden.Lflag &^= syscall.ECHO
    if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), syscall.TCSETS, uintptr(unsafe.Pointer(&hidden))); errno != 0 {
        return nil, errno
    }
    restorer := &echoRestorer{file: file, state: state, signals: make(chan os.Signal, 1), done: make(chan struct{})}
    signal.Notify(restorer.signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
    go restorer.watch()
    return restorer.Restore, nil
}

type echoRestorer struct {
    file    *os.File
    state   syscall.Termios
    signals chan os.Signal
    done    chan struct{}
}

func (r *echoRestorer) watch() {
    select {
        case received := <-r.signals:
            r.reset()
            os.Exit(128 + int(received.(syscall.Signal)))
        case <-r.done:
    }
}

func (r *echoRestorer) reset() {
    syscall.Syscall(syscall.SYS_IOCTL, r.file.Fd(), syscall.TCSETS, uintptr(unsafe.Pointer(&r.state)))
}

func (r *echoRestorer) Restore() {
    signal.Stop(r.signals)
    close(r.done)
    r.reset()
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
