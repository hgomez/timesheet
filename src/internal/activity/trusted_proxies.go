/*
 * trusted_proxies.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "strings"
)

// ---------------------------------------------------------------------------
// TrustedProxies declaration
// ---------------------------------------------------------------------------

type TrustedProxies struct {
    prefixes []netip.Prefix
}

const mappedPrefixBits = 96

var ErrInvalidTrustedProxies = errors.New("activity: invalid trusted proxies (IP addresses or prefixes such as 10.0.0.0/8)")

// ---------------------------------------------------------------------------
// TrustedProxies implementation
// ---------------------------------------------------------------------------

func NewTrustedProxies(elements []string) (TrustedProxies, error) {
    proxies := TrustedProxies{}
    for _, element := range elements {
        prefix, err := proxies.parse(element)
        if err != nil {
            return TrustedProxies{}, ErrInvalidTrustedProxies
        }
        proxies.prefixes = append(proxies.prefixes, prefix)
    }
    return proxies, nil
}

func (p TrustedProxies) parse(element string) (netip.Prefix, error) {
    if strings.Contains(element, "/") {
        prefix, err := netip.ParsePrefix(element)
        if err != nil {
            return netip.Prefix{}, err
        }
        return netip.PrefixFrom(prefix.Addr().Unmap(), p.bits(prefix)).Masked(), nil
    }
    address, err := netip.ParseAddr(element)
    if (err != nil) || (address.Zone() != "") {
        return netip.Prefix{}, ErrInvalidTrustedProxies
    }
    address = address.Unmap()
    return netip.PrefixFrom(address, address.BitLen()), nil
}

func (p TrustedProxies) bits(prefix netip.Prefix) int {
    if prefix.Addr().Is4In6() && (prefix.Bits() >= mappedPrefixBits) {
        return prefix.Bits() - mappedPrefixBits
    }
    return prefix.Bits()
}

func (p TrustedProxies) IsEmpty() bool {
    return len(p.prefixes) == 0
}

func (p TrustedProxies) Contains(address netip.Addr) bool {
    address = address.Unmap().WithZone("")
    for _, prefix := range p.prefixes {
        if prefix.Contains(address) {
            return true
        }
    }
    return false
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
