/*
 * origins.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "errors"
    "net"
    "net/netip"
    "net/url"
    "slices"
    "strconv"
    "strings"
)

// ---------------------------------------------------------------------------
// Origins declaration
// ---------------------------------------------------------------------------

type Origins struct {
    secure bool
    hosts  map[string]bool
    list   []string
    names  []string
    ips    []netip.Addr
}

var (
    ErrInvalidOrigin = errors.New("web: invalid public origin")
    ErrMixedSchemes  = errors.New("web: public origins must all use the same scheme")
)

// ---------------------------------------------------------------------------
// Origins implementation
// ---------------------------------------------------------------------------

func NewPublicOrigins(public []string) (*Origins, error) {
    o := &Origins{hosts: map[string]bool{}}
    scheme := ""
    for _, raw := range public {
        parsed, err := url.Parse(strings.TrimSpace(raw))
        if (err != nil) || ((parsed.Scheme != "http") && (parsed.Scheme != "https")) || (parsed.Host == "") || ((parsed.Path != "") && (parsed.Path != "/")) || (parsed.RawQuery != "") || parsed.ForceQuery || (parsed.Fragment != "") || (parsed.User != nil) {
            return nil, ErrInvalidOrigin
        }
        if (scheme != "") && (parsed.Scheme != scheme) {
            return nil, ErrMixedSchemes
        }
        scheme = parsed.Scheme
        o.secure = scheme == "https"
        host, port, ok := o.split(parsed.Host)
        if !ok {
            return nil, ErrInvalidOrigin
        }
        o.add(host, port)
    }
    if len(o.list) == 0 {
        return nil, ErrInvalidOrigin
    }
    return o, nil
}

func NewLocalOrigins(addresses []netip.Addr, hostname string, port int, secure bool) *Origins {
    o := &Origins{secure: secure, hosts: map[string]bool{}}
    o.add("localhost", port)
    o.add("127.0.0.1", port)
    o.add("::1", port)
    for _, address := range addresses {
        address = address.Unmap()
        if address.Is4() && !address.IsUnspecified() && !address.IsMulticast() {
            o.add(address.String(), port)
        }
    }
    if name := o.canonicalName(hostname); (name != "") && (name != "localhost") {
        o.add(name, port)
        if !strings.HasSuffix(name, ".local") {
            o.add(name + ".local", port)
        }
    }
    return o
}

func (o *Origins) Secure() bool {
    return o.secure
}

func (o *Origins) List() []string {
    return slices.Clone(o.list)
}

func (o *Origins) DNSNames() []string {
    return slices.Clone(o.names)
}

func (o *Origins) IPAddresses() []netip.Addr {
    return slices.Clone(o.ips)
}

func (o *Origins) AllowsHost(hostHeader string) bool {
    host, port, ok := o.split(hostHeader)
    return ok && o.hosts[o.key(host, port)]
}

func (o *Origins) AllowsOrigin(origin string) bool {
    parsed, err := url.Parse(origin)
    if (err != nil) || (parsed.Path != "") || (parsed.RawQuery != "") || parsed.ForceQuery || (parsed.Fragment != "") || (parsed.User != nil) {
        return false
    }
    if parsed.Scheme != o.scheme() {
        return false
    }
    return o.AllowsHost(parsed.Host)
}

func (o *Origins) scheme() string {
    if o.secure {
        return "https"
    }
    return "http"
}

func (o *Origins) defaultPort() int {
    if o.secure {
        return 443
    }
    return 80
}

func (o *Origins) add(host string, port int) {
    key := o.key(host, port)
    if o.hosts[key] {
        return
    }
    o.hosts[key] = true
    if address, err := netip.ParseAddr(host); err == nil {
        o.ips = append(o.ips, address)
    } else {
        o.names = append(o.names, host)
    }
    o.list = append(o.list, o.scheme() + "://" + key)
}

func (o *Origins) split(hostport string) (string, int, bool) {
    host, portText, err := net.SplitHostPort(hostport)
    port := o.defaultPort()
    if err != nil {
        host = strings.TrimSuffix(strings.TrimPrefix(hostport, "["), "]")
    } else {
        parsed, err := strconv.Atoi(portText)
        if (err != nil) || (parsed < 1) || (parsed > 65535) {
            return "", 0, false
        }
        port = parsed
    }
    if address, err := netip.ParseAddr(host); err == nil {
        if address.Zone() != "" {
            return "", 0, false
        }
        return address.Unmap().String(), port, true
    }
    name := o.canonicalName(host)
    return name, port, name != ""
}

func (o *Origins) canonicalName(host string) string {
    name := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
    if (name == "") || (len(name) > 253) {
        return ""
    }
    for _, r := range name {
        valid := ((r >= 'a') && (r <= 'z')) || ((r >= '0') && (r <= '9')) || (r == '-') || (r == '.')
        if !valid {
            return ""
        }
    }
    return name
}

func (o *Origins) key(host string, port int) string {
    shown := host
    if strings.Contains(host, ":") {
        shown = "[" + host + "]"
    }
    if port == o.defaultPort() {
        return shown
    }
    return shown + ":" + strconv.Itoa(port)
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
