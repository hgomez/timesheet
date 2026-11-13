/*
 * request_limits.go - Copyright (c) 2023-2026 - Olivier Poncet
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
// RequestLimits declaration
// ---------------------------------------------------------------------------

type RequestLimits struct {
    next    http.Handler
    errors  *ErrorPages
    maxURL  int
    maxBody int64
}

// ---------------------------------------------------------------------------
// RequestLimits implementation
// ---------------------------------------------------------------------------

func NewRequestLimits(next http.Handler, errors *ErrorPages, maxURL int, maxBody int64) *RequestLimits {
    return &RequestLimits{next: next, errors: errors, maxURL: maxURL, maxBody: maxBody}
}

func (l *RequestLimits) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if (len(r.RequestURI) > l.maxURL) || (len(r.URL.String()) > l.maxURL) {
        l.errors.Status(w, r, http.StatusRequestURITooLong)
        return
    }
    r.Body = http.MaxBytesReader(l.unwrap(w), r.Body, l.maxBody)
    l.next.ServeHTTP(w, r)
}

func (l *RequestLimits) unwrap(w http.ResponseWriter) http.ResponseWriter {
    for {
        wrapper, ok := w.(interface{ Unwrap() http.ResponseWriter })
        if !ok {
            return w
        }
        w = wrapper.Unwrap()
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
