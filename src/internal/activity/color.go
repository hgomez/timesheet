/*
 * color.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "math"
    "strconv"
)

// ---------------------------------------------------------------------------
// Color declaration
// ---------------------------------------------------------------------------

type Color string

const (
    DarkInk  Color = "#000000"
    LightInk Color = "#ffffff"
)

// ---------------------------------------------------------------------------
// Color implementation
// ---------------------------------------------------------------------------

func (c Color) Valid() bool {
    if (len(c) != 7) || (c[0] != '#') {
        return false
    }
    for _, digit := range []byte(c[1:]) {
        if ((digit < '0') || (digit > '9')) && ((digit < 'a') || (digit > 'f')) {
            return false
        }
    }
    return true
}

func (c Color) IsSet() bool {
    return c != ""
}

func (c Color) Ink() Color {
    if c.Contrast(DarkInk) >= c.Contrast(LightInk) {
        return DarkInk
    }
    return LightInk
}

func (c Color) Contrast(other Color) float64 {
    a, b := c.luminance(), other.luminance()
    if a < b {
        a, b = b, a
    }
    return (a + 0.05) / (b + 0.05)
}

func (c Color) luminance() float64 {
    if !c.Valid() {
        return 0
    }
    total := 0.0
    weights := []float64{0.2126, 0.7152, 0.0722}
    for i, weight := range weights {
        value, _ := strconv.ParseUint(string(c[(1 + (2 * i)):(3 + (2 * i))]), 16, 8)
        channel := float64(value) / 255
        if channel <= 0.04045 {
            channel = channel / 12.92
        } else {
            channel = math.Pow((channel + 0.055) / 1.055, 2.4)
        }
        total += weight * channel
    }
    return total
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
