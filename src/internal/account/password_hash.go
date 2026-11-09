/*
 * password_hash.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "crypto/pbkdf2"
    "crypto/sha256"
    "crypto/subtle"
    "encoding/base64"
    "errors"
    "io"
    "strconv"
    "strings"
)

// ---------------------------------------------------------------------------
// PasswordHash declaration
// ---------------------------------------------------------------------------

type PasswordHash string

const (
    DefaultIterations = 600000
    MinIterations     = 1000
    MaxIterations     = 2 * DefaultIterations
    passwordScheme    = "pbkdf2-sha256"
    saltSize          = 16
    keySize           = 32
)

var ErrInvalidPasswordHash = errors.New("account: invalid password hash")

// ---------------------------------------------------------------------------
// PasswordHash implementation
// ---------------------------------------------------------------------------

func NewPasswordHash(password string, iterations int, random io.Reader) (PasswordHash, error) {
    if (iterations < MinIterations) || (iterations > MaxIterations) {
        return "", ErrInvalidPasswordHash
    }
    salt := make([]byte, saltSize)
    if _, err := io.ReadFull(random, salt); err != nil {
        return "", err
    }
    key, err := pbkdf2.Key(sha256.New, password, salt, iterations, keySize)
    if err != nil {
        return "", err
    }
    encoding := base64.RawStdEncoding
    return PasswordHash(passwordScheme + "$" + strconv.Itoa(iterations) + "$" + encoding.EncodeToString(salt) + "$" + encoding.EncodeToString(key)), nil
}

func (h PasswordHash) parts() (int, []byte, []byte, error) {
    fields := strings.Split(string(h), "$")
    if (len(fields) != 4) || (fields[0] != passwordScheme) {
        return 0, nil, nil, ErrInvalidPasswordHash
    }
    iterations, err := strconv.Atoi(fields[1])
    if (err != nil) || (iterations < MinIterations) || (iterations > MaxIterations) {
        return 0, nil, nil, ErrInvalidPasswordHash
    }
    salt, err1 := base64.RawStdEncoding.DecodeString(fields[2])
    key, err2 := base64.RawStdEncoding.DecodeString(fields[3])
    if (err1 != nil) || (err2 != nil) || (len(salt) != saltSize) || (len(key) != keySize) {
        return 0, nil, nil, ErrInvalidPasswordHash
    }
    return iterations, salt, key, nil
}

func (h PasswordHash) Valid() bool {
    _, _, _, err := h.parts()
    return err == nil
}

func (h PasswordHash) Verify(password string) bool {
    iterations, salt, key, err := h.parts()
    if err != nil {
        return false
    }
    candidate, err := pbkdf2.Key(sha256.New, password, salt, iterations, keySize)
    return (err == nil) && (subtle.ConstantTimeCompare(candidate, key) == 1)
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
