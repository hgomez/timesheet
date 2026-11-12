/*
 * access_log.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "log/slog"
    "net/http"
    "time"

    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// AccessLog declaration
// ---------------------------------------------------------------------------

type AccessLog struct {
    next    http.Handler
    address *ClientAddress
    clock   activity.Clock
    logger  *slog.Logger
}

const (
    AnonymousUser    = "-"
    maxLoggedAgent   = 200
    accessLogMessage = "request"
)

// ---------------------------------------------------------------------------
// AccessLog implementation
// ---------------------------------------------------------------------------

func NewAccessLog(next http.Handler, address *ClientAddress, clock activity.Clock, logger *slog.Logger) *AccessLog {
    return &AccessLog{next: next, address: address, clock: clock, logger: logger}
}

func (l *AccessLog) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    started := l.clock.Now()
    entry := NewAccessEntry()
    tracker := NewTrackingWriter(w)
    defer l.record(tracker, r, entry, started)
    l.next.ServeHTTP(tracker, entry.Attach(r))
}

func (l *AccessLog) record(w *TrackingWriter, r *http.Request, entry *AccessEntry, started time.Time) {
    peer := l.address.Peer(r)
    client, known := l.address.Of(r)
    if !known {
        client = peer
    }
    l.logger.Info(accessLogMessage,
        "client", client.String(),
        "peer", peer.String(),
        "user", entry.User(),
        "method", r.Method,
        "host", r.Host,
        "path", r.URL.Path,
        "protocol", r.Proto,
        "status", w.Status(),
        "bytes", w.Bytes(),
        "duration", l.clock.Now().Sub(started).String(),
        "agent", l.agent(r),
    )
}

func (l *AccessLog) agent(r *http.Request) string {
    agent := r.UserAgent()
    if len(agent) > maxLoggedAgent {
        agent = agent[:maxLoggedAgent]
    }
    return agent
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
