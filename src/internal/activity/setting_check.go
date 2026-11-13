/*
 * setting_check.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "fmt"
    "time"
)

// ---------------------------------------------------------------------------
// SettingCheck declaration
// ---------------------------------------------------------------------------

type SettingCheck struct {
    section string
}

// ---------------------------------------------------------------------------
// SettingCheck implementation
// ---------------------------------------------------------------------------

func NewSettingCheck(section string) SettingCheck {
    return SettingCheck{section: section}
}

func (c SettingCheck) Fail(key string, message string) error {
    return fmt.Errorf("%w: %s.%s: %s", ErrInvalidConfiguration, c.section, key, message)
}

func (c SettingCheck) Duration(key string, value Duration, low time.Duration, high time.Duration) error {
    if (value.Std() < low) || (value.Std() > high) || ((value.Std() % time.Millisecond) != 0) {
        return c.Fail(key, "a duration from " + Duration(low).String() + " to " + Duration(high).String() + " is expected")
    }
    return nil
}

func (c SettingCheck) Integer(key string, value int64, low int64, high int64) error {
    if (value < low) || (value > high) {
        return c.Fail(key, fmt.Sprintf("an integer from %d to %d is expected", low, high))
    }
    return nil
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
