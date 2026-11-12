/*
 * origin_guard.go - Copyright (c) 2023-2026 - Olivier Poncet
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

import "net/http"

// ---------------------------------------------------------------------------
// OriginGuard declaration
// ---------------------------------------------------------------------------

type OriginGuard struct {
    next    http.Handler
    origins *Origins
    errors  *ErrorPages
}

// ---------------------------------------------------------------------------
// OriginGuard implementation
// ---------------------------------------------------------------------------

func NewOriginGuard(next http.Handler, origins *Origins, errors *ErrorPages) *OriginGuard {
    return &OriginGuard{next: next, origins: origins, errors: errors}
}

func (g *OriginGuard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if g.safe(r.Method) {
        g.next.ServeHTTP(w, r)
        return
    }
    origin := r.Header.Get("Origin")
    site := r.Header.Get("Sec-Fetch-Site")
    if (origin == "") || (origin == "null") || !g.origins.AllowsOrigin(origin) || ((site != "") && (site != "same-origin") && (site != "none")) {
        g.errors.Forbidden(w, r, ForbiddenOrigin)
        return
    }
    g.next.ServeHTTP(w, r)
}

func (g *OriginGuard) safe(method string) bool {
    switch method {
        case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
            return true
    }
    return false
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
