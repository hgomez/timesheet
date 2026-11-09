/*
 * profile.go - Copyright (c) 2023-2026 - Olivier Poncet
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

import "errors"

// ---------------------------------------------------------------------------
// Profile declaration
// ---------------------------------------------------------------------------

type Profile string

const (
    ProfileGuest Profile = "GUEST"
    ProfileUser  Profile = "USER"
    ProfileAdmin Profile = "ADMIN"
)

var ErrInvalidProfile = errors.New("account: invalid profile (USER or ADMIN)")

// ---------------------------------------------------------------------------
// Profile implementation
// ---------------------------------------------------------------------------

func NewProfile(text string) (Profile, error) {
    profile := Profile(text)
    if !profile.Persistent() {
        return "", ErrInvalidProfile
    }
    return profile, nil
}

func (p Profile) Rank() int {
    switch p {
        case ProfileAdmin:
            return 2
        case ProfileUser:
            return 1
    }
    return 0
}

func (p Profile) AtLeast(required Profile) bool {
    return p.Rank() >= required.Rank()
}

func (p Profile) Persistent() bool {
    return (p == ProfileUser) || (p == ProfileAdmin)
}

func (p Profile) Label() string {
    switch p {
        case ProfileAdmin:
            return "Administrateur"
        case ProfileUser:
            return "Utilisateur"
    }
    return "Visiteur"
}

func (p *Profile) UnmarshalText(data []byte) error {
    profile, err := NewProfile(string(data))
    if err != nil {
        return err
    }
    *p = profile
    return nil
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
