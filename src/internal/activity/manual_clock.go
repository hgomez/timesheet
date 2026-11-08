/*
 * manual_clock.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "sync"
    "time"
)

// ---------------------------------------------------------------------------
// ManualClock declaration
// ---------------------------------------------------------------------------

type ManualClock struct {
    mu  sync.Mutex
    now time.Time
}

var _ Clock = (*ManualClock)(nil)

// ---------------------------------------------------------------------------
// ManualClock implementation
// ---------------------------------------------------------------------------

func NewManualClock(now time.Time) *ManualClock {
    return &ManualClock{now: now}
}

func (c *ManualClock) Now() time.Time {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.now
}

func (c *ManualClock) Set(now time.Time) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.now = now
}

func (c *ManualClock) Advance(d time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.now = c.now.Add(d)
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
