/*
 * trusted_proxies_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "net/netip"
    "testing"
)

// ---------------------------------------------------------------------------
// TrustedProxies tests
// ---------------------------------------------------------------------------

func TestTrustedProxiesHoldAddressesAndPrefixes(t *testing.T) {
    proxies, err := NewTrustedProxies([]string{"192.168.0.0/16", "172.16.0.0/12", "10.0.0.0/8", "203.0.113.7", "::ffff:198.51.100.4", "2001:db8::/32", "10.9.8.7/8"})
    if (err != nil) || proxies.IsEmpty() {
        t.Fatalf("list refused: %v", err)
    }
    cases := map[string]bool{
        "192.168.1.10": true, "172.16.0.1": true, "172.31.255.254": true, "172.32.0.1": false, "10.255.0.1": true, "11.0.0.1": false,
        "203.0.113.7": true, "203.0.113.8": false, "198.51.100.4": true, "::ffff:10.1.2.3": true, "2001:db8:1::1": true, "2001:db9::1": false, "127.0.0.1": false,
    }
    for text, want := range cases {
        if got := proxies.Contains(netip.MustParseAddr(text)); got != want {
            t.Errorf("%s: trusted %v, want %v", text, got, want)
        }
    }
    none, err := NewTrustedProxies(nil)
    if (err != nil) || !none.IsEmpty() || none.Contains(netip.MustParseAddr("10.0.0.1")) {
        t.Fatalf("an empty list trusts nobody: %v", err)
    }
    for _, text := range []string{"10.0.0.0/33", "", " 10.0.0.1", "10.0.0.0/8,10.0.0.1", "proxy.example", "fe80::1%eth0", "10.0.0.1 10.0.0.2", "10.0.0.0/", "/8"} {
        if _, err := NewTrustedProxies([]string{"10.0.0.1", text}); !errors.Is(err, ErrInvalidTrustedProxies) {
            t.Errorf("%q: %v", text, err)
        }
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
