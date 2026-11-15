/*
 * server_settings_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "errors"
    "strings"
    "testing"
    "time"
)

// ---------------------------------------------------------------------------
// ServerSettings tests
// ---------------------------------------------------------------------------

func TestServerSettingsValidationNamesTheField(t *testing.T) {
    cases := map[string]func(*ServerSettings){
        "server.timezone":            func(s *ServerSettings) { s.Timezone = "Mars/Olympus" },
        "server.listen":              func(s *ServerSettings) { s.Listen = "nonsense" },
        "server.origins":             func(s *ServerSettings) { s.Origins = nil },
        "server.origins[1]":          func(s *ServerSettings) { s.Origins = []string{"http://a.test:1", "https://b.test"} },
        "server.trusted_proxies":     func(s *ServerSettings) { s.TrustedProxies = []string{"10.0.0.0/8", "proxy.example"} },
        "server.log_level":           func(s *ServerSettings) { s.LogLevel = "loud" },
        "server.tls_cert":            func(s *ServerSettings) { s.TLSCert = "cert.pem" },
        "server.tls_min_version":     func(s *ServerSettings) { s.TLSMinVersion = "1.1" },
        "server.hsts_max_age":        func(s *ServerSettings) { s.HSTSMaxAge = Duration(time.Hour) },
        "server.read_header_timeout": func(s *ServerSettings) { s.ReadHeaderTimeout = Duration(500 * time.Millisecond) },
        "server.keep_alive_timeout":  func(s *ServerSettings) { s.KeepAliveTimeout = Duration(11 * time.Minute) },
        "server.shutdown_timeout":    func(s *ServerSettings) { s.ShutdownTimeout = Duration(2 * time.Minute) },
        "server.max_header_bytes":    func(s *ServerSettings) { s.MaxHeaderBytes = 1024 },
        "server.max_body_bytes":      func(s *ServerSettings) { s.MaxBodyBytes = 16 << 20 },
        "server.max_url_bytes":       func(s *ServerSettings) { s.MaxURLBytes = 100 },
    }
    for path, mutate := range cases {
        server := NewDefaultServerSettings().Clone()
        mutate(&server)
        err := server.Validate()
        if !errors.Is(err, ErrInvalidConfiguration) || !strings.Contains(err.Error(), path + ":") {
            t.Errorf("%s: got %v", path, err)
        }
    }
    for _, listen := range []string{"127.0.0.1:99999", "127.0.0.1:-1", "127.0.0.1:no-such-service"} {
        server := NewDefaultServerSettings().Clone()
        server.Listen = listen
        if err := server.Validate(); !errors.Is(err, ErrInvalidConfiguration) || !strings.Contains(err.Error(), "server.listen: a port") {
            t.Errorf("%s: got %v", listen, err)
        }
    }
    emptied := NewDefaultServerSettings().Clone()
    emptied.TrustedProxies = nil
    if data, err := json.Marshal(emptied); (err != nil) || !strings.Contains(string(data), `"trusted_proxies":[]`) || !strings.Contains(string(data), `"tls_cert":null`) || !strings.Contains(string(data), `"keep_alive_timeout":"1m"`) {
        t.Fatalf("an empty list is written as a list, an absent path as null, a duration in its short form: %s %v", data, err)
    }
}

func TestServerSettingsAcceptATLSPairAndALogOffset(t *testing.T) {
    server := NewDefaultServerSettings().Clone()
    server.TLSCert, server.TLSKey, server.LogLevel, server.TLSMinVersion = "cert.pem", "key.pem", "INFO+2", TLSVersion13
    if err := server.Validate(); err != nil {
        t.Fatal(err)
    }
    if level, err := server.Level(); (err != nil) || (level != 2) || !server.TLS() {
        t.Fatalf("got %v %v %v", level, err, server.TLS())
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
