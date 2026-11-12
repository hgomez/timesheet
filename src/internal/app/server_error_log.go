/*
 * server_error_log.go - Copyright (c) 2023-2026 - Olivier Poncet
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

package app

import (
    "context"
    "io"
    "log/slog"
    "strings"
)

// ---------------------------------------------------------------------------
// ServerErrorLog declaration
// ---------------------------------------------------------------------------

type ServerErrorLog struct {
    logger *slog.Logger
}

var _ io.Writer = (*ServerErrorLog)(nil)

// ---------------------------------------------------------------------------
// ServerErrorLog implementation
// ---------------------------------------------------------------------------

func NewServerErrorLog(logger *slog.Logger) *ServerErrorLog {
    return &ServerErrorLog{logger: logger}
}

func (l *ServerErrorLog) Write(p []byte) (int, error) {
    message := strings.TrimSpace(string(p))
    l.logger.Log(context.Background(), l.level(message), message)
    return len(p), nil
}

func (l *ServerErrorLog) level(message string) slog.Level {
    if strings.Contains(message, "TLS handshake error") {
        return slog.LevelDebug
    }
    return slog.LevelWarn
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
