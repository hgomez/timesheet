/*
 * state.go - Copyright (c) 2023-2026 - Olivier Poncet
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
// State declaration
// ---------------------------------------------------------------------------

type State string

const (
    StateEnabled  State = "enabled"
    StateDisabled State = "disabled"
)

var ErrInvalidState = errors.New("account: invalid state (enabled or disabled)")

// ---------------------------------------------------------------------------
// State implementation
// ---------------------------------------------------------------------------

func NewState(text string) (State, error) {
    state := State(text)
    if !state.Valid() {
        return "", ErrInvalidState
    }
    return state, nil
}

func (s State) Valid() bool {
    return (s == StateEnabled) || (s == StateDisabled)
}

func (s State) Enabled() bool {
    return s == StateEnabled
}

func (s State) Label() string {
    if s.Enabled() {
        return "actif"
    }
    return "désactivé"
}

func (s *State) UnmarshalText(data []byte) error {
    state, err := NewState(string(data))
    if err != nil {
        return err
    }
    *s = state
    return nil
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
