/*
 * quantity.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "errors"
    "math"
    "strconv"
    "strings"
)

// ---------------------------------------------------------------------------
// Quantity declaration
// ---------------------------------------------------------------------------

type Quantity struct {
    num int64
    den int64
}

const (
    maxQuantityText      = 32
    maxQuantityDecimals  = 6
    maxQuantityMagnitude = 1000000
    QuantityBase         = 63000000
)

var (
    ErrInvalidQuantity  = errors.New("activity: invalid quantity")
    ErrQuantityOverflow = errors.New("activity: quantity overflow")
)

// ---------------------------------------------------------------------------
// Quantity implementation
// ---------------------------------------------------------------------------

func NewQuantity(num int64, den int64) Quantity {
    if den == 0 {
        panic(ErrInvalidQuantity)
    }
    if den < 0 {
        num, den = -num, -den
    }
    divisor := Quantity{}.gcd(Quantity{}.magnitude(num), den)
    return Quantity{num: num / divisor, den: den / divisor}
}

func NewQuantityFromText(text string) (Quantity, error) {
    q := Quantity{den: 1}
    if err := q.UnmarshalText([]byte(text)); err != nil {
        return Quantity{den: 1}, err
    }
    return q, nil
}

func (q Quantity) magnitude(value int64) int64 {
    if value < 0 {
        return -value
    }
    return value
}

func (q Quantity) gcd(a int64, b int64) int64 {
    for b != 0 {
        a, b = b, a % b
    }
    if a == 0 {
        return 1
    }
    return a
}

func (q Quantity) normal() Quantity {
    if q.den == 0 {
        return Quantity{num: 0, den: 1}
    }
    return q
}

func (q Quantity) Add(other Quantity) Quantity {
    a, b := q.normal(), other.normal()
    common, ok0 := q.multiply(a.den / q.gcd(a.den, b.den), b.den)
    if !ok0 {
        panic(ErrQuantityOverflow)
    }
    left, ok1 := q.multiply(a.num, common / a.den)
    right, ok2 := q.multiply(b.num, common / b.den)
    sum := left + right
    if !ok1 || !ok2 || ((left > 0) && (right > 0) && (sum < 0)) || ((left < 0) && (right < 0) && (sum >= 0)) {
        panic(ErrQuantityOverflow)
    }
    return NewQuantity(sum, common)
}

func (q Quantity) multiply(a int64, b int64) (int64, bool) {
    if (a == 0) || (b == 0) {
        return 0, true
    }
    product := a * b
    return product, ((product / b) == a) && (a != math.MinInt64) && (b != math.MinInt64)
}

func (q Quantity) Sub(other Quantity) Quantity {
    return q.Add(other.Neg())
}

func (q Quantity) Neg() Quantity {
    a := q.normal()
    return Quantity{num: -a.num, den: a.den}
}

func (q Quantity) Abs() Quantity {
    if q.Sign() < 0 {
        return q.Neg()
    }
    return q.normal()
}

func (q Quantity) Sign() int {
    switch {
        case q.num > 0:
            return 1
        case q.num < 0:
            return -1
    }
    return 0
}

func (q Quantity) IsZero() bool {
    return q.num == 0
}

func (q Quantity) Cmp(other Quantity) int {
    return q.Sub(other).Sign()
}

func (q Quantity) Ceil() Quantity {
    a := q.normal()
    whole := a.num / a.den
    if (a.num % a.den) > 0 {
        whole++
    }
    return NewQuantity(whole, 1)
}

func (q Quantity) Equal(other Quantity) bool {
    return q.Cmp(other) == 0
}

func (q Quantity) String() string {
    a := q.normal()
    if a.den == 1 {
        return strconv.FormatInt(a.num, 10)
    }
    scale := int64(1)
    for digits := 1; digits <= maxQuantityDecimals; digits++ {
        scale *= 10
        if (scale % a.den) == 0 {
            scaled := q.magnitude(a.num) * (scale / a.den)
            text := strconv.FormatInt(scaled, 10)
            for len(text) <= digits {
                text = "0" + text
            }
            sign := ""
            if a.num < 0 {
                sign = "-"
            }
            return sign + text[:len(text) - digits] + "." + text[len(text) - digits:]
        }
    }
    return strconv.FormatInt(a.num, 10) + "/" + strconv.FormatInt(a.den, 10)
}

func (q Quantity) Format() string {
    a := q.normal()
    hundredths := ((q.magnitude(a.num) * 200) + a.den) / (2 * a.den)
    text := strconv.FormatInt(hundredths, 10)
    for len(text) < 3 {
        text = "0" + text
    }
    sign := ""
    if (a.num < 0) && (hundredths != 0) {
        sign = "-"
    }
    return sign + text[:len(text) - 2] + "," + text[len(text) - 2:]
}

func (q Quantity) FormatSigned() string {
    text := q.Format()
    if (q.Sign() > 0) && (text != "0,00") {
        return "+" + text
    }
    return text
}

func (q Quantity) FormatDays() string {
    a := q.normal()
    if a.den == 1 {
        return strconv.FormatInt(a.num, 10)
    }
    if a.den == 2 {
        return strings.Replace(q.String(), ".", ",", 1)
    }
    return q.Format()
}

func (q Quantity) MarshalText() ([]byte, error) {
    return []byte(q.String()), nil
}

func (q *Quantity) UnmarshalText(data []byte) error {
    text := strings.TrimSpace(string(data))
    if (text == "") || (len(text) > maxQuantityText) {
        return ErrInvalidQuantity
    }
    negative := strings.HasPrefix(text, "-")
    body := text
    if negative || strings.HasPrefix(text, "+") {
        body = text[1:]
    }
    if strings.HasPrefix(body, "-") || strings.HasPrefix(body, "+") {
        return ErrInvalidQuantity
    }
    if numerator, denominator, found := strings.Cut(body, "/"); found {
        if !q.digits(numerator) || !q.digits(denominator) || (numerator == "") || (denominator == "") {
            return ErrInvalidQuantity
        }
        if negative {
            numerator = "-" + numerator
        }
        num, err1 := strconv.ParseInt(numerator, 10, 64)
        den, err2 := strconv.ParseInt(denominator, 10, 64)
        if (err1 != nil) || (err2 != nil) || (den <= 0) || (den > QuantityBase) || ((QuantityBase % den) != 0) || (num == math.MinInt64) || (q.magnitude(num) > (maxQuantityMagnitude * den)) {
            return ErrInvalidQuantity
        }
        *q = NewQuantity(num, den)
        return nil
    }
    whole, fraction, _ := strings.Cut(body, ".")
    if ((whole == "") && (fraction == "")) || (len(fraction) > maxQuantityDecimals) || !q.digits(whole) || !q.digits(fraction) {
        return ErrInvalidQuantity
    }
    if whole == "" {
        whole = "0"
    }
    scale := int64(1)
    for range fraction {
        scale *= 10
    }
    integer, err := strconv.ParseInt(whole + fraction, 10, 64)
    if (err != nil) || (integer > (maxQuantityMagnitude * scale)) {
        return ErrInvalidQuantity
    }
    if negative {
        integer = -integer
    }
    *q = NewQuantity(integer, scale)
    return nil
}

func (q Quantity) digits(text string) bool {
    for _, r := range text {
        if (r < '0') || (r > '9') {
            return false
        }
    }
    return true
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
