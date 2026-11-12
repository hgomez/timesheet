/*
 * access_entry.go - Copyright (c) 2023-2026 - Olivier Poncet
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

package web

import (
    "context"
    "net/http"

    "timesheet/internal/account"
)

// ---------------------------------------------------------------------------
// AccessEntry declaration
// ---------------------------------------------------------------------------

type AccessEntry struct {
    login account.Login
}

type accessEntryKey struct{}

// ---------------------------------------------------------------------------
// AccessEntry implementation
// ---------------------------------------------------------------------------

func NewAccessEntry() *AccessEntry {
    return &AccessEntry{}
}

func (e *AccessEntry) Attach(r *http.Request) *http.Request {
    return r.WithContext(context.WithValue(r.Context(), accessEntryKey{}, e))
}

func (e *AccessEntry) From(r *http.Request) *AccessEntry {
    if entry, ok := r.Context().Value(accessEntryKey{}).(*AccessEntry); ok {
        return entry
    }
    return nil
}

func (e *AccessEntry) Identify(login account.Login) {
    if e != nil {
        e.login = login
    }
}

func (e *AccessEntry) User() string {
    if (e == nil) || (e.login == "") {
        return AnonymousUser
    }
    return string(e.login)
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
