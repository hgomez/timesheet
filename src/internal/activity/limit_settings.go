/*
 * limit_settings.go - Copyright (c) 2023-2026 - Olivier Poncet
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

// ---------------------------------------------------------------------------
// LimitSettings declaration
// ---------------------------------------------------------------------------

type LimitSettings struct {
    Accounts     int64 `json:"accounts"`
    AccountBytes int64 `json:"account_bytes"`
    Adjustments  int64 `json:"adjustments"`
}

const (
    maxAccounts     = 10000
    minAccountBytes = 64 << 10
    maxAccountBytes = 16 << 20
    maxAdjustments  = 100000
)

// ---------------------------------------------------------------------------
// LimitSettings implementation
// ---------------------------------------------------------------------------

func NewDefaultLimitSettings() LimitSettings {
    return LimitSettings{Accounts: 200, AccountBytes: 1 << 20, Adjustments: 1000}
}

func (s LimitSettings) Validate() error {
    check := NewSettingCheck("limits")
    if err := check.Integer("accounts", s.Accounts, 1, maxAccounts); err != nil {
        return err
    }
    if err := check.Integer("account_bytes", s.AccountBytes, minAccountBytes, maxAccountBytes); err != nil {
        return err
    }
    return check.Integer("adjustments", s.Adjustments, 1, maxAdjustments)
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
