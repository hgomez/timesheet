/*
 * session_settings.go - Copyright (c) 2023-2026 - Olivier Poncet
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

package activity

import (
    "time"
)

// ---------------------------------------------------------------------------
// SessionSettings declaration
// ---------------------------------------------------------------------------

type SessionSettings struct {
    IdleTimeout  Duration `json:"idle_timeout"`
    MaxLifetime  Duration `json:"max_lifetime"`
    PerAccount   int64    `json:"per_account"`
    SaveInterval Duration `json:"save_interval"`
}

const (
    minSessionIdle     = 5 * time.Minute
    maxSessionIdle     = 7 * 24 * time.Hour
    maxSessionLifetime = 30 * 24 * time.Hour
    maxSessionsPerUser = 1000
    minSessionSave     = 5 * time.Second
    maxSessionSave     = time.Hour
)

// ---------------------------------------------------------------------------
// SessionSettings implementation
// ---------------------------------------------------------------------------

func NewDefaultSessionSettings() SessionSettings {
    return SessionSettings{IdleTimeout: Duration(8 * time.Hour), MaxLifetime: Duration(24 * time.Hour), PerAccount: 64, SaveInterval: Duration(time.Minute)}
}

func (s SessionSettings) Validate() error {
    check := NewSettingCheck("sessions")
    if err := check.Duration("idle_timeout", s.IdleTimeout, minSessionIdle, maxSessionIdle); err != nil {
        return err
    }
    if err := check.Duration("max_lifetime", s.MaxLifetime, s.IdleTimeout.Std(), maxSessionLifetime); err != nil {
        return err
    }
    if err := check.Integer("per_account", s.PerAccount, 1, maxSessionsPerUser); err != nil {
        return err
    }
    return check.Duration("save_interval", s.SaveInterval, minSessionSave, maxSessionSave)
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
