/*
 * account.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "time"
)

// ---------------------------------------------------------------------------
// Account declaration
// ---------------------------------------------------------------------------

const MaxNameLength = 60

type Account struct {
    Login              Login        `json:"login"`
    Name               string       `json:"name"`
    Profile            Profile      `json:"profile"`
    Password           PasswordHash `json:"password"`
    MustChangePassword bool         `json:"must_change_password"`
    State              State        `json:"state"`
    Created            time.Time    `json:"created"`
    LastLogin          time.Time    `json:"last_login"`
}

var (
    ErrUnknownAccount  = errors.New("account: unknown account")
    ErrDuplicateLogin  = errors.New("account: this login is already taken")
    ErrLastAdmin       = errors.New("account: the last administrator cannot be removed or demoted")
    ErrTooManyAccounts = errors.New("account: too many accounts")
    ErrInvalidName     = errors.New("account: invalid name")
)

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
