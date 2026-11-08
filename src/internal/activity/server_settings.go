/*
 * server_settings.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "fmt"
    "net"
    "net/url"
    "slices"
    "strconv"
    "strings"
    "time"
)

// ---------------------------------------------------------------------------
// ServerSettings declaration
// ---------------------------------------------------------------------------

type ServerSettings struct {
    Addr           string   `json:"addr"`
    Origins        []string `json:"origins"`
    TrustedProxies []string `json:"trusted_proxies"`
    Timezone       string   `json:"timezone"`
}

const DefaultTimezone = "Europe/Paris"

// ---------------------------------------------------------------------------
// ServerSettings implementation
// ---------------------------------------------------------------------------

func NewDefaultServerSettings() ServerSettings {
    return ServerSettings{Addr: "127.0.0.1:8080", Origins: []string{"http://127.0.0.1:8080", "http://localhost:8080"}, TrustedProxies: []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}, Timezone: DefaultTimezone}
}

func (s ServerSettings) MarshalJSON() ([]byte, error) {
    return json.Marshal(struct {
        Addr           string   `json:"addr"`
        Origins        []string `json:"origins"`
        TrustedProxies []string `json:"trusted_proxies"`
        Timezone       string   `json:"timezone"`
    }{s.Addr, s.list(s.Origins), s.list(s.TrustedProxies), s.Timezone})
}

func (s ServerSettings) Validate() error {
    if _, port, err := net.SplitHostPort(s.Addr); (err != nil) || (port == "") {
        return s.fail("server.addr", "an address host:port is expected")
    }
    if len(s.Origins) == 0 {
        return s.fail("server.origins", "at least one origin is expected")
    }
    scheme := ""
    for i, origin := range s.Origins {
        parsed, err := url.Parse(origin)
        port := 443
        if (err == nil) && (parsed.Port() != "") {
            port, err = strconv.Atoi(parsed.Port())
        }
        if (err != nil) || (origin != strings.TrimSpace(origin)) || ((parsed.Scheme != "http") && (parsed.Scheme != "https")) || (parsed.Hostname() == "") || (port < 1) || (port > 65535) || ((parsed.Path != "") && (parsed.Path != "/")) || (parsed.RawQuery != "") || parsed.ForceQuery || (parsed.Fragment != "") || (parsed.User != nil) || ((scheme != "") && (parsed.Scheme != scheme)) {
            return s.fail(fmt.Sprintf("server.origins[%d]", i), "origins scheme://host[:port], all http or all https, are expected")
        }
        scheme = parsed.Scheme
    }
    if _, err := NewTrustedProxies(s.TrustedProxies); err != nil {
        return s.fail("server.trusted_proxies", "IP addresses or prefixes such as 10.0.0.0/8 are expected")
    }
    if _, err := time.LoadLocation(s.Timezone); (err != nil) || (s.Timezone == "") || (s.Timezone == "Local") {
        return s.fail("server.timezone", "an IANA time zone such as Europe/Paris is expected")
    }
    return nil
}

func (s ServerSettings) fail(path string, message string) error {
    return fmt.Errorf("%w: %s: %s", ErrInvalidConfiguration, path, message)
}

func (s ServerSettings) Location() *time.Location {
    location, err := time.LoadLocation(s.Timezone)
    if err != nil {
        return time.UTC
    }
    return location
}

func (s ServerSettings) Clone() ServerSettings {
    clone := s
    clone.Origins = slices.Clone(s.Origins)
    clone.TrustedProxies = slices.Clone(s.TrustedProxies)
    return clone
}

func (s ServerSettings) list(values []string) []string {
    if values == nil {
        return []string{}
    }
    return values
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
