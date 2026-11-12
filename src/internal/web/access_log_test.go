/*
 * access_log_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "net/http"
    "net/url"
    "strings"
    "testing"
    "time"
)

// ---------------------------------------------------------------------------
// AccessLog tests
// ---------------------------------------------------------------------------

func TestEveryRequestIsLoggedWithItsUser(t *testing.T) {
    site := newTestSite(t)
    guest := site.client()
    _, _, body := guest.get("/login?notice=saved&secret=query-secret")
    guest.post("/login", url.Values{"csrf": {guest.token(body)}, "login": {"typed-login"}, "password": {"typed secret password"}})
    guest.get("/nowhere")
    request, _ := http.NewRequest(http.MethodGet, site.server.URL + "/", nil)
    request.Host = "other.example:8080"
    request.Header.Set("User-Agent", strings.Repeat("a", 500))
    if response, err := http.DefaultClient.Do(request); err == nil {
        response.Body.Close()
    }
    site.clock.Advance(10 * time.Minute)
    user := site.client().login("demo")
    user.get("/dashboard/")
    user.get("/static/css/theme.css")
    admin := site.client().login("root")
    admin.get("/admin/accounts/demo/year/2026")
    journal := site.logs.waitFor("path=/admin/accounts/demo/year/2026 ")
    lines := []string{}
    for _, line := range strings.Split(journal, "\n") {
        if strings.Contains(line, "msg=request") {
            lines = append(lines, line)
        }
    }
    expected := [][]string{
        {"method=GET", "path=/login ", "status=200", "user=- ", "client=127.0.0.1", "peer=127.0.0.1", "protocol=HTTP/1.1", "duration="},
        {"method=POST", "path=/login ", "status=401", "user=- "},
        {"method=GET", "path=/nowhere ", "status=404", "user=- "},
        {"method=GET", "path=/ ", "status=421", "host=other.example:8080", "agent=" + strings.Repeat("a", 200) + " "},
        {"method=POST", "path=/login ", "status=303", "user=demo "},
        {"method=GET", "path=/dashboard/ ", "status=200", "user=demo "},
        {"method=GET", "path=/static/css/theme.css ", "status=200", "user=- "},
        {"method=GET", "path=/admin/accounts/demo/year/2026 ", "status=200", "user=root "},
    }
    for _, fields := range expected {
        found := false
        for _, line := range lines {
            matches := true
            for _, field := range fields {
                matches = matches && strings.Contains(line + " ", field)
            }
            found = found || matches
        }
        if !found {
            t.Errorf("no request line with %v in:\n%s", fields, strings.Join(lines, "\n"))
        }
    }
    for _, secret := range []string{"query-secret", "notice=saved", "typed-login", "typed secret password", guest.token(body)} {
        if strings.Contains(journal, secret) {
            t.Errorf("the journal contains %q", secret)
        }
    }
    for _, line := range lines {
        if strings.Contains(line, "bytes=0") && strings.Contains(line, "path=/dashboard/ ") {
            t.Errorf("no size on %s", line)
        }
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
