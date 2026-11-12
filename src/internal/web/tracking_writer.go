/*
 * tracking_writer.go - Copyright (c) 2023-2026 - Olivier Poncet
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
// TrackingWriter declaration
// ---------------------------------------------------------------------------

type TrackingWriter struct {
    http.ResponseWriter
    wrote  bool
    status int
    bytes  int64
}

// ---------------------------------------------------------------------------
// TrackingWriter implementation
// ---------------------------------------------------------------------------

func NewTrackingWriter(w http.ResponseWriter) *TrackingWriter {
    return &TrackingWriter{ResponseWriter: w}
}

func (t *TrackingWriter) WriteHeader(status int) {
    if t.status == 0 {
        t.status = status
    }
    t.wrote = true
    t.ResponseWriter.WriteHeader(status)
}

func (t *TrackingWriter) Write(data []byte) (int, error) {
    if t.status == 0 {
        t.status = http.StatusOK
    }
    t.wrote = true
    written, err := t.ResponseWriter.Write(data)
    t.bytes += int64(written)
    return written, err
}

func (t *TrackingWriter) Status() int {
    if t.status == 0 {
        return http.StatusOK
    }
    return t.status
}

func (t *TrackingWriter) Bytes() int64 {
    return t.bytes
}

func (t *TrackingWriter) Wrote() bool {
    return t.wrote
}

func (t *TrackingWriter) Unwrap() http.ResponseWriter {
    return t.ResponseWriter
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
