/*
 * recovery_handler.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "fmt"
    "log/slog"
    "net/http"
)

// ---------------------------------------------------------------------------
// RecoveryHandler declaration
// ---------------------------------------------------------------------------

type RecoveryHandler struct {
    next   http.Handler
    errors *ErrorPages
    logger *slog.Logger
}

// ---------------------------------------------------------------------------
// RecoveryHandler implementation
// ---------------------------------------------------------------------------

func NewRecoveryHandler(next http.Handler, errors *ErrorPages, logger *slog.Logger) *RecoveryHandler {
    return &RecoveryHandler{next: next, errors: errors, logger: logger}
}

func (h *RecoveryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    snapshot := w.Header().Clone()
    tracker := NewTrackingWriter(w)
    defer h.recover(tracker, r, snapshot)
    h.next.ServeHTTP(tracker, r)
}

func (h *RecoveryHandler) recover(w *TrackingWriter, r *http.Request, snapshot http.Header) {
    value := recover()
    if value == nil {
        return
    }
    if value == http.ErrAbortHandler {
        panic(value)
    }
    h.logger.Error("handler panic", "path", r.URL.Path, "panic", fmt.Sprint(value))
    if w.Wrote() {
        panic(http.ErrAbortHandler)
    }
    header := w.Header()
    for key := range header {
        delete(header, key)
    }
    for key, values := range snapshot {
        header[key] = values
    }
    h.errors.Status(w, r, http.StatusInternalServerError)
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
