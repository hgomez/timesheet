/*
 * password_policy.go - Copyright (c) 2023-2026 - Olivier Poncet
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

package account

import (
    "errors"
    "strings"
    "unicode"
    "unicode/utf8"
)

// ---------------------------------------------------------------------------
// PasswordPolicy declaration
// ---------------------------------------------------------------------------

type PasswordPolicy struct{}

const (
    MinPasswordLength   = 12
    MinPasswordDistinct = 6
    MaxPasswordBytes    = 256
)

var ErrWeakPassword = errors.New("account: the password needs at least 12 characters, 6 distinct characters, no surrounding spaces, 256 bytes at most, and must differ from the login")

// ---------------------------------------------------------------------------
// PasswordPolicy implementation
// ---------------------------------------------------------------------------

func NewPasswordPolicy() *PasswordPolicy {
    return &PasswordPolicy{}
}

func (p *PasswordPolicy) Check(login Login, password string) error {
    distinct := map[rune]bool{}
    printable := true
    for _, r := range password {
        distinct[r] = true
        printable = printable && !unicode.IsControl(r) && !unicode.In(r, unicode.Cf, unicode.Zl, unicode.Zp)
    }
    usable := printable && (strings.TrimSpace(password) == password) && (len(password) <= MaxPasswordBytes) && utf8.ValidString(password)
    if !usable || (utf8.RuneCountInString(password) < MinPasswordLength) || (len(distinct) < MinPasswordDistinct) || strings.EqualFold(password, string(login)) {
        return ErrWeakPassword
    }
    return nil
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
