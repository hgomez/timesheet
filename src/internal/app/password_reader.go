/*
 * password_reader.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "bufio"
    "errors"
    "fmt"
    "io"
    "os"
    "strings"
)

// ---------------------------------------------------------------------------
// PasswordReader declaration
// ---------------------------------------------------------------------------

type PasswordReader struct {
    stdin     io.Reader
    stderr    io.Writer
    lookupEnv func(string) (string, bool)
}

const PasswordVariable = "TIMESHEET_PASSWORD"

var (
    ErrPasswordMismatch = errors.New("app: the two passwords differ")
    ErrEmptyPassword    = errors.New("app: " + PasswordVariable + " is set but empty")
)

// ---------------------------------------------------------------------------
// PasswordReader implementation
// ---------------------------------------------------------------------------

func NewPasswordReader(stdin io.Reader, stderr io.Writer, lookupEnv func(string) (string, bool)) *PasswordReader {
    return &PasswordReader{stdin: stdin, stderr: stderr, lookupEnv: lookupEnv}
}

func (p *PasswordReader) Read() (string, error) {
    if value, set := p.lookupEnv(PasswordVariable); set {
        if value == "" {
            return "", ErrEmptyPassword
        }
        return value, nil
    }
    if file, ok := p.stdin.(*os.File); ok {
        if restore, err := p.hideEcho(file); err == nil {
            defer restore()
            first, err := p.prompt(file, "Password: ")
            if err != nil {
                return "", err
            }
            second, err := p.prompt(file, "Confirmation: ")
            if err != nil {
                return "", err
            }
            if first != second {
                return "", ErrPasswordMismatch
            }
            return first, nil
        }
        if info, err := file.Stat(); (err == nil) && ((info.Mode() & os.ModeCharDevice) != 0) {
            fmt.Fprintln(p.stderr, "warning: the password is read with echo on this system; prefer " + PasswordVariable + " or a pipe")
        }
    }
    return p.line(p.stdin)
}

func (p *PasswordReader) prompt(file *os.File, label string) (string, error) {
    fmt.Fprint(p.stderr, label)
    value, err := p.line(file)
    fmt.Fprintln(p.stderr)
    return value, err
}

func (p *PasswordReader) line(source io.Reader) (string, error) {
    reader := bufio.NewReader(io.LimitReader(source, 4096))
    text, err := reader.ReadString('\n')
    if (err != nil) && !errors.Is(err, io.EOF) {
        return "", err
    }
    return strings.TrimRight(text, "\r\n"), nil
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
