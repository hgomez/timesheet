/*
 * usage_error.go - Copyright (c) 2023-2026 - Olivier Poncet
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

// ---------------------------------------------------------------------------
// UsageError declaration
// ---------------------------------------------------------------------------

type UsageError struct {
    reason string
}

var _ error = (*UsageError)(nil)

// ---------------------------------------------------------------------------
// UsageError implementation
// ---------------------------------------------------------------------------

func NewUsageError(reason string) *UsageError {
    return &UsageError{reason: reason}
}

func (e *UsageError) Error() string {
    return e.reason
}

func (e *UsageError) Is(target error) bool {
    return target == ErrUsage
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
