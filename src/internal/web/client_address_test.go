/*
 * client_address_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "net/http/httptest"
    "net/netip"
    "testing"

    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// ClientAddress tests
// ---------------------------------------------------------------------------

func TestTrustedProxiesGiveTheForwardedClient(t *testing.T) {
    cases := []struct {
        proxies []string
        peer    string
        trusted bool
    }{
        {[]string{"127.0.0.1"}, "127.0.0.1:40000", true},
        {[]string{"::ffff:127.0.0.1"}, "127.0.0.1:40000", true},
        {[]string{"127.0.0.1"}, "[::ffff:127.0.0.1]:40000", true},
        {[]string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}, "10.20.30.40:40000", true},
        {[]string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}, "172.20.0.3:40000", true},
        {[]string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}, "198.51.100.20:40000", false},
        {nil, "10.20.30.40:40000", false},
    }
    for _, item := range cases {
        proxies, err := activity.NewTrustedProxies(item.proxies)
        if err != nil {
            t.Fatal(err)
        }
        request := httptest.NewRequest("POST", "/login", nil)
        request.RemoteAddr = item.peer
        request.Header.Set("X-Forwarded-For", "192.0.2.66, 203.0.113.9")
        client, ok := NewClientAddress(proxies, nil).Of(request)
        if !ok || ((client == netip.MustParseAddr("203.0.113.9")) != item.trusted) {
            t.Errorf("%v from %s: client %v %v", item.proxies, item.peer, client, ok)
        }
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
