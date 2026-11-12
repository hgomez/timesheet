/*
 * host_guard.go - Copyright (c) 2023-2026 - Olivier Poncet
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
// HostGuard declaration
// ---------------------------------------------------------------------------

type HostGuard struct {
    next    http.Handler
    origins *Origins
    errors  *ErrorPages
}

// ---------------------------------------------------------------------------
// HostGuard implementation
// ---------------------------------------------------------------------------

func NewHostGuard(next http.Handler, origins *Origins, errors *ErrorPages) *HostGuard {
    return &HostGuard{next: next, origins: origins, errors: errors}
}

func (g *HostGuard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if !g.origins.AllowsHost(r.Host) {
        g.errors.Status(w, r, http.StatusMisdirectedRequest)
        return
    }
    g.next.ServeHTTP(w, r)
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
