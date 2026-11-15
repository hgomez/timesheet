/*
 * quantity_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "encoding/json"
    "testing"
)

// ---------------------------------------------------------------------------
// Quantity tests
// ---------------------------------------------------------------------------

func TestQuantityParsesDecimalsAndFractions(t *testing.T) {
    cases := map[string]string{"-7/12": "-7/12", "+7/12": "7/12", "2.08": "2.08", "7/12": "7/12", "14/24": "14/24", "8/12": "8/12", "0": "0", "-1.5": "-1.5", "0.50": "0.50", "25/12": "25/12", "3/4": "3/4", "+2": "2", ".5": "0.5"}
    for text, canonical := range cases {
        if got := q(t, text).String(); got != canonical {
            t.Errorf("%s: got %s, want %s", text, got, canonical)
        }
    }
    for _, text := range []string{"", "abc", "1/0", "1/-2", "1.1234567", "2,08", "1e3", "--1", "1/2/3", "99999999999", "-9223372036854775808/1", "1/999999", "1/11", "1/63000001", "-+1", "+-1", "+1/+2", "1/+2", "-1/-2", "-/2", "1/"} {
        if _, err := NewQuantityFromText(text); err == nil {
            t.Errorf("%q accepted", text)
        }
    }
}

func TestQuantityKeepsItsWrittenFormAndComputesCanonically(t *testing.T) {
    if !q(t, "8/12").Equal(q(t, "2/3")) || !q(t, "0.50").Equal(q(t, "1/2")) || (q(t, "8/12").Cmp(q(t, "2/3")) != 0) {
        t.Fatal("equal values must compare equal whatever their written form")
    }
    if NewQuantity(8, 12) != q(t, "8/12") {
        t.Fatal("NewQuantity must keep the written form")
    }
    written := map[string]Quantity{"1": q(t, "8/12").Add(q(t, "4/12")), "2/3": NewQuantity(0, 1).Add(q(t, "8/12")), "0.5": NewQuantity(0, 1).Add(NewQuantity(1, 2)), "0.25": q(t, "0.50").Sub(q(t, "0.25")), "-8/12": q(t, "8/12").Neg(), "3": q(t, "8/3").Ceil()}
    for want, got := range written {
        if got.String() != want {
            t.Errorf("got %s, want %s", got, want)
        }
    }
    if (q(t, "2/4").FormatDays() != "0,5") || (q(t, "4/2").FormatDays() != "2") || (q(t, "8/12").Format() != "0,67") {
        t.Fatalf("display of written forms: %s %s %s", q(t, "2/4").FormatDays(), q(t, "4/2").FormatDays(), q(t, "8/12").Format())
    }
}

func TestQuantityArithmeticIsExact(t *testing.T) {
    total := NewQuantity(0, 1)
    tp := NewQuantity(0, 1)
    for range 12 {
        total = total.Add(q(t, "2.08"))
        tp = tp.Add(q(t, "7/12"))
    }
    expectQuantity(t, "12 x 2.08", total, "24.96")
    expectQuantity(t, "12 x 7/12", tp, "7")
    expectQuantity(t, "sub", q(t, "1").Sub(q(t, "7/12")), "5/12")
}

func TestQuantityStaysReadableAfterArithmetic(t *testing.T) {
    total := NewQuantity(0, 1)
    for _, text := range []string{"1/7", "7/12", "1/9", "2.08", "1/64", "1/15625", "-1/63000000", "3/14"} {
        total = total.Add(q(t, text))
        if _, err := NewQuantityFromText(total.String()); err != nil {
            t.Fatalf("%s cannot be read back: %v", total, err)
        }
    }
    defer func() {
        if recover() != ErrQuantityOverflow {
            t.Fatal("overflow must panic")
        }
    }()
    NewQuantity(9223372036854775807, 1).Add(NewQuantity(1, 1))
}

func TestQuantityRoundsUpToTheWholeDay(t *testing.T) {
    cases := map[string]string{"11.71": "12", "18": "18", "307/12": "26", "0": "0", "-0.5": "0", "-1.5": "-1", "1/63000000": "1"}
    for text, want := range cases {
        expectQuantity(t, text, q(t, text).Ceil(), want)
    }
}

func TestQuantityFormatsInFrench(t *testing.T) {
    cases := map[string]string{"7/12": "0,58", "2.08": "2,08", "-0.17": "-0,17", "0": "0,00", "24.96": "24,96", "-1/1000": "0,00", "35/12": "2,92"}
    for text, want := range cases {
        if got := q(t, text).Format(); got != want {
            t.Errorf("%s: got %s, want %s", text, got, want)
        }
    }
    signed := map[string]string{"0.31": "+0,31", "-0.24": "-0,24", "0": "0,00", "1/1000": "0,00", "-1/1000": "0,00"}
    for text, want := range signed {
        if got := q(t, text).FormatSigned(); got != want {
            t.Errorf("signed %s: got %s, want %s", text, got, want)
        }
    }
    if got := q(t, "1.5").FormatDays(); got != "1,5" {
        t.Fatalf("days: %s", got)
    }
}

func TestQuantityJSONRoundTrip(t *testing.T) {
    data, err := json.Marshal(map[string]Quantity{"a": q(t, "7/12"), "b": q(t, "2.08")})
    if (err != nil) || (string(data) != `{"a":"7/12","b":"2.08"}`) {
        t.Fatalf("marshal: %s %v", data, err)
    }
    back := map[string]Quantity{}
    if err := json.Unmarshal(data, &back); (err != nil) || !back["a"].Equal(q(t, "7/12")) {
        t.Fatalf("unmarshal: %v %v", back, err)
    }
}

func TestQuantityDetectsADenominatorOverflow(t *testing.T) {
    defer func() {
        if recover() != ErrQuantityOverflow {
            t.Fatal("no overflow reported")
        }
    }()
    NewQuantity(1, 4294967311).Add(NewQuantity(1, 4294967357))
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
