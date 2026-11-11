//go:build !unix || aix || (solaris && !illumos)

/*
 * secure_file_other.go - Copyright (c) 2023-2026 - Olivier Poncet
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

package store

// ---------------------------------------------------------------------------
// SecureFile implementation (other platforms)
// ---------------------------------------------------------------------------

func (a *SecureFile) Read() ([]byte, bool, error) {
    return nil, false, ErrUnsupportedPlatform
}

func (a *SecureFile) Lock() error {
    return ErrUnsupportedPlatform
}

func (a *SecureFile) inspectDirectory(warn bool) error {
    return ErrUnsupportedPlatform
}

func (a *SecureFile) Write(data []byte) error {
    return ErrUnsupportedPlatform
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
