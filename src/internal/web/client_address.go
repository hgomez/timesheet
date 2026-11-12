/*
 * client_address.go - Copyright (c) 2023-2026 - Olivier Poncet
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

package web

import (
    "net"
    "net/http"
    "net/netip"
    "strings"

    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// ClientAddress declaration
// ---------------------------------------------------------------------------

type ClientAddress struct {
    trusted activity.TrustedProxies
    local   []netip.Prefix
}

// ---------------------------------------------------------------------------
// ClientAddress implementation
// ---------------------------------------------------------------------------

func NewClientAddress(trusted activity.TrustedProxies, local []netip.Prefix) *ClientAddress {
    return &ClientAddress{trusted: trusted, local: local}
}

func (c *ClientAddress) Peer(r *http.Request) netip.Addr {
    host, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        host = r.RemoteAddr
    }
    address, err := netip.ParseAddr(host)
    if err != nil {
        return netip.Addr{}
    }
    return address.Unmap().WithZone("")
}

func (c *ClientAddress) Of(r *http.Request) (netip.Addr, bool) {
    peer := c.Peer(r)
    if !c.trusted.Contains(peer) {
        return peer, peer.IsValid()
    }
    lines := r.Header.Values("X-Forwarded-For")
    if len(lines) == 0 {
        return netip.Addr{}, false
    }
    elements := strings.Split(lines[len(lines) - 1], ",")
    element := strings.TrimSpace(elements[len(elements) - 1])
    address, err := netip.ParseAddr(element)
    if err != nil {
        withPort, err := netip.ParseAddrPort(element)
        if err != nil {
            return netip.Addr{}, false
        }
        address = withPort.Addr()
    }
    return address.Unmap().WithZone(""), true
}

func (c *ClientAddress) Key(r *http.Request) netip.Prefix {
    address, ok := c.Of(r)
    if !ok {
        return netip.PrefixFrom(netip.IPv6Unspecified(), 128)
    }
    if address.Is4() {
        return netip.PrefixFrom(address, 32)
    }
    if address.IsPrivate() || address.IsLinkLocalUnicast() || address.IsLoopback() || c.isLocal(address) {
        return netip.PrefixFrom(address, 128)
    }
    prefix, _ := address.Prefix(64)
    return prefix
}

func (c *ClientAddress) isLocal(address netip.Addr) bool {
    for _, prefix := range c.local {
        if prefix.Contains(address) {
            return true
        }
    }
    return false
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
