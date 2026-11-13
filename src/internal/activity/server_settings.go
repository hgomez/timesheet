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
    "log/slog"
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
    Listen            string       `json:"listen"`
    Origins           []string     `json:"origins"`
    TrustedProxies    []string     `json:"trusted_proxies"`
    Timezone          string       `json:"timezone"`
    LogLevel          string       `json:"log_level"`
    TLSCert           NullableText `json:"tls_cert"`
    TLSKey            NullableText `json:"tls_key"`
    TLSMinVersion     string       `json:"tls_min_version"`
    HSTSMaxAge        Duration     `json:"hsts_max_age"`
    ReadHeaderTimeout Duration     `json:"read_header_timeout"`
    ReadTimeout       Duration     `json:"read_timeout"`
    WriteTimeout      Duration     `json:"write_timeout"`
    KeepAliveTimeout  Duration     `json:"keep_alive_timeout"`
    ShutdownTimeout   Duration     `json:"shutdown_timeout"`
    MaxHeaderBytes    int64        `json:"max_header_bytes"`
    MaxBodyBytes      int64        `json:"max_body_bytes"`
    MaxURLBytes       int64        `json:"max_url_bytes"`
}

const (
    DefaultTimezone = "Europe/Paris"

    TLSVersion12 = "1.2"
    TLSVersion13 = "1.3"

    minRequestTimeout  = time.Second
    maxRequestTimeout  = 10 * time.Minute
    minShutdownTimeout = time.Second
    maxShutdownTimeout = time.Minute
    minHSTSMaxAge      = 24 * time.Hour
    maxHSTSMaxAge      = 2 * 8760 * time.Hour
    minHeaderBytes     = 4 << 10
    maxHeaderBytes     = 1 << 20
    minBodyBytes       = 256 << 10
    maxBodyBytes       = 8 << 20
    minURLBytes        = 1024
    maxURLBytes        = 65536
)

// ---------------------------------------------------------------------------
// ServerSettings implementation
// ---------------------------------------------------------------------------

func NewDefaultServerSettings() ServerSettings {
    settings := NewOptionalServerSettings()
    settings.Listen = "127.0.0.1:8080"
    settings.Origins = []string{"http://127.0.0.1:8080", "http://localhost:8080"}
    settings.TrustedProxies = []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}
    settings.Timezone = DefaultTimezone
    return settings
}

func NewOptionalServerSettings() ServerSettings {
    return ServerSettings{
        LogLevel:          "info",
        TLSMinVersion:     TLSVersion12,
        HSTSMaxAge:        Duration(8760 * time.Hour),
        ReadHeaderTimeout: Duration(5 * time.Second),
        ReadTimeout:       Duration(15 * time.Second),
        WriteTimeout:      Duration(30 * time.Second),
        KeepAliveTimeout:  Duration(time.Minute),
        ShutdownTimeout:   Duration(10 * time.Second),
        MaxHeaderBytes:    8 << 10,
        MaxBodyBytes:      256 << 10,
        MaxURLBytes:       2048,
    }
}

func (s ServerSettings) MarshalJSON() ([]byte, error) {
    type plain ServerSettings
    copied := plain(s)
    copied.Origins = s.list(s.Origins)
    copied.TrustedProxies = s.list(s.TrustedProxies)
    return json.Marshal(copied)
}

func (s ServerSettings) Validate() error {
    check := NewSettingCheck("server")
    if _, port, err := net.SplitHostPort(s.Listen); (err != nil) || (port == "") {
        return check.Fail("listen", "an address host:port is expected")
    }
    if len(s.Origins) == 0 {
        return check.Fail("origins", "at least one origin is expected")
    }
    scheme := ""
    for i, origin := range s.Origins {
        parsed, err := url.Parse(origin)
        port := 443
        if (err == nil) && (parsed.Port() != "") {
            port, err = strconv.Atoi(parsed.Port())
        }
        if (err != nil) || (origin != strings.TrimSpace(origin)) || ((parsed.Scheme != "http") && (parsed.Scheme != "https")) || (parsed.Hostname() == "") || (port < 1) || (port > 65535) || ((parsed.Path != "") && (parsed.Path != "/")) || (parsed.RawQuery != "") || parsed.ForceQuery || (parsed.Fragment != "") || (parsed.User != nil) || ((scheme != "") && (parsed.Scheme != scheme)) {
            return check.Fail(fmt.Sprintf("origins[%d]", i), "origins scheme://host[:port], all http or all https, are expected")
        }
        scheme = parsed.Scheme
    }
    if _, err := NewTrustedProxies(s.TrustedProxies); err != nil {
        return check.Fail("trusted_proxies", "IP addresses or prefixes such as 10.0.0.0/8 are expected")
    }
    if _, err := time.LoadLocation(s.Timezone); (err != nil) || (s.Timezone == "") || (s.Timezone == "Local") {
        return check.Fail("timezone", "an IANA time zone such as Europe/Paris is expected")
    }
    if _, err := s.Level(); err != nil {
        return check.Fail("log_level", "debug, info, warn or error, with an optional offset such as info+2, is expected")
    }
    if (s.TLSCert == "") != (s.TLSKey == "") {
        return check.Fail("tls_cert", "tls_cert and tls_key go together")
    }
    if (s.TLSMinVersion != TLSVersion12) && (s.TLSMinVersion != TLSVersion13) {
        return check.Fail("tls_min_version", "1.2 or 1.3 is expected")
    }
    if err := check.Duration("hsts_max_age", s.HSTSMaxAge, minHSTSMaxAge, maxHSTSMaxAge); err != nil {
        return err
    }
    timeouts := []struct {
        key   string
        value Duration
    }{{"read_header_timeout", s.ReadHeaderTimeout}, {"read_timeout", s.ReadTimeout}, {"write_timeout", s.WriteTimeout}, {"keep_alive_timeout", s.KeepAliveTimeout}}
    for _, timeout := range timeouts {
        if err := check.Duration(timeout.key, timeout.value, minRequestTimeout, maxRequestTimeout); err != nil {
            return err
        }
    }
    if err := check.Duration("shutdown_timeout", s.ShutdownTimeout, minShutdownTimeout, maxShutdownTimeout); err != nil {
        return err
    }
    if err := check.Integer("max_header_bytes", s.MaxHeaderBytes, minHeaderBytes, maxHeaderBytes); err != nil {
        return err
    }
    if err := check.Integer("max_body_bytes", s.MaxBodyBytes, minBodyBytes, maxBodyBytes); err != nil {
        return err
    }
    return check.Integer("max_url_bytes", s.MaxURLBytes, minURLBytes, maxURLBytes)
}

func (s ServerSettings) Level() (slog.Level, error) {
    var level slog.Level
    err := level.UnmarshalText([]byte(s.LogLevel))
    return level, err
}

func (s ServerSettings) TLS() bool {
    return s.TLSCert != ""
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
