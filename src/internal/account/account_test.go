/*
 * account_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "crypto/rand"
    "errors"
    "testing"
)

// ---------------------------------------------------------------------------
// Account tests
// ---------------------------------------------------------------------------

func TestLoginsAndProfiles(t *testing.T) {
    for _, text := range []string{"olivier", "o.poncet", "a1", "jean-marc_2"} {
        if _, err := NewLogin(text); err != nil {
            t.Errorf("%s refused", text)
        }
    }
    for _, text := range []string{"", "a", "Olivier", ".hidden", "../x", "a b", "abcdefghijklmnopqrstuvwxyzabcdefgh"} {
        if _, err := NewLogin(text); err == nil {
            t.Errorf("%q accepted", text)
        }
    }
    if _, err := NewProfile("GUEST"); !errors.Is(err, ErrInvalidProfile) {
        t.Fatal("GUEST must never be persisted")
    }
    if !ProfileAdmin.AtLeast(ProfileUser) || ProfileUser.AtLeast(ProfileAdmin) || ProfileGuest.AtLeast(ProfileUser) {
        t.Fatal("profile order")
    }
    for _, text := range []string{"", "ENABLED", "active", "removed"} {
        if _, err := NewState(text); !errors.Is(err, ErrInvalidState) {
            t.Errorf("state %q accepted", text)
        }
    }
    if !StateEnabled.Enabled() || StateDisabled.Enabled() || (StateEnabled.Label() != "actif") || (StateDisabled.Label() != "désactivé") {
        t.Fatal("states")
    }
}

func TestPasswordPolicy(t *testing.T) {
    policy := NewPasswordPolicy()
    if err := policy.Check("olivier", "correct horse battery"); err != nil {
        t.Fatal(err)
    }
    for _, password := range []string{"short", "aaaaaaaaaaaaaaaa", " leading space1", "olivierolivier", "ababababababab", "abcdefghijkl\x01", "abcdefghijkl\u202e"} {
        if err := policy.Check("olivierolivier", password); err == nil {
            t.Errorf("%q accepted", password)
        }
    }
}

func TestPasswordHash(t *testing.T) {
    hash, err := NewPasswordHash("correct horse battery", MinIterations, rand.Reader)
    if err != nil {
        t.Fatal(err)
    }
    if !hash.Valid() || !hash.Verify("correct horse battery") || hash.Verify("correct horse batterY") {
        t.Fatal("verification")
    }
    if PasswordHash("pbkdf2-sha256$10$AAAA$BBBB").Valid() || PasswordHash("plain").Verify("plain") {
        t.Fatal("malformed hashes must be refused")
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
