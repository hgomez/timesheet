/*
 * security_headers.go - Copyright (c) 2023-2026 - Olivier Poncet
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
// SecurityHeaders declaration
// ---------------------------------------------------------------------------

type SecurityHeaders struct {
    next   http.Handler
    secure bool
}

const contentSecurityPolicy = "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self'; connect-src 'self'; font-src 'self'; form-action 'self'; base-uri 'none'; frame-ancestors 'none'; require-trusted-types-for 'script'; trusted-types 'none'"

// ---------------------------------------------------------------------------
// SecurityHeaders implementation
// ---------------------------------------------------------------------------

func NewSecurityHeaders(next http.Handler, secure bool) *SecurityHeaders {
    return &SecurityHeaders{next: next, secure: secure}
}

func (s *SecurityHeaders) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    header := w.Header()
    header.Set("Content-Security-Policy", contentSecurityPolicy)
    header.Set("X-Content-Type-Options", "nosniff")
    header.Set("X-Frame-Options", "DENY")
    header.Set("Referrer-Policy", "same-origin")
    header.Set("Cross-Origin-Resource-Policy", "same-origin")
    header.Set("Cross-Origin-Opener-Policy", "same-origin")
    header.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=(), usb=(), interest-cohort=()")
    header.Set("Cache-Control", "no-store")
    header.Set("Pragma", "no-cache")
    header.Set("Expires", "0")
    if s.secure {
        header.Set("Strict-Transport-Security", "max-age=31536000")
    }
    s.next.ServeHTTP(w, r)
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
