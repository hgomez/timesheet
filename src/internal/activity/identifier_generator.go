/*
 * identifier_generator.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "encoding/hex"
    "io"
)

// ---------------------------------------------------------------------------
// IdentifierGenerator declaration
// ---------------------------------------------------------------------------

type IdentifierGenerator struct {
    random io.Reader
}

const identifierBytes = 8

// ---------------------------------------------------------------------------
// IdentifierGenerator implementation
// ---------------------------------------------------------------------------

func NewIdentifierGenerator(random io.Reader) *IdentifierGenerator {
    return &IdentifierGenerator{random: random}
}

func (g *IdentifierGenerator) Next() (string, error) {
    buffer := make([]byte, identifierBytes)
    if _, err := io.ReadFull(g.random, buffer); err != nil {
        return "", err
    }
    return hex.EncodeToString(buffer), nil
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
