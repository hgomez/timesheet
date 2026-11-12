/*
 * session.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "time"

    "timesheet/internal/account"
)

// ---------------------------------------------------------------------------
// Session declaration
// ---------------------------------------------------------------------------

type Session struct {
    login   account.Login
    csrf    string
    created time.Time
    seen    time.Time
}

// ---------------------------------------------------------------------------
// Session implementation
// ---------------------------------------------------------------------------

func NewSession(login account.Login, csrf string, now time.Time) *Session {
    return &Session{login: login, csrf: csrf, created: now, seen: now}
}

func (s *Session) Login() account.Login {
    return s.login
}

func (s *Session) CSRF() string {
    return s.csrf
}

func (s *Session) Expired(now time.Time, sliding time.Duration, absolute time.Duration) bool {
    return (now.Sub(s.seen) > sliding) || (now.Sub(s.created) > absolute) || s.created.After(now) || s.seen.After(now)
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
