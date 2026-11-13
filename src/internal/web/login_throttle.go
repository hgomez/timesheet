/*
 * login_throttle.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "net/netip"
    "sync"
    "time"

    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// LoginThrottle declaration
// ---------------------------------------------------------------------------

type LoginThrottle struct {
    mu           sync.Mutex
    clock        activity.Clock
    entries      map[netip.Prefix]*throttleEntry
    maxKeys      int
    budget       int
    initialDelay time.Duration
    maxDelay     time.Duration
    doublings    int
    window       time.Duration
    forget       time.Duration
    windowStart  time.Time
    newFailures  int
}

type throttleEntry struct {
    failures     int
    blockedUntil time.Time
    lastFailure  time.Time
    inFlight     bool
    overBudget   bool
}

const inFlightRetry = time.Second

// ---------------------------------------------------------------------------
// LoginThrottle implementation
// ---------------------------------------------------------------------------

func NewLoginThrottle(clock activity.Clock, settings activity.LoginSettings) *LoginThrottle {
    throttle := &LoginThrottle{clock: clock, entries: map[netip.Prefix]*throttleEntry{}, maxKeys: int(settings.TrackedAddresses), budget: int(settings.NewAddressesPerWindow), initialDelay: settings.InitialDelay.Std(), maxDelay: settings.MaxDelay.Std(), window: settings.BudgetWindow.Std(), forget: settings.ForgetAfter.Std()}
    for delay := throttle.initialDelay; delay < throttle.maxDelay; delay *= 2 {
        throttle.doublings++
    }
    return throttle
}

func (t *LoginThrottle) Begin(key netip.Prefix) (bool, time.Duration) {
    t.mu.Lock()
    defer t.mu.Unlock()
    now := t.clock.Now()
    if entry, known := t.entries[key]; known && !entry.inFlight && (now.Sub(entry.lastFailure) >= t.forget) && !now.Before(entry.blockedUntil) {
        delete(t.entries, key)
    }
    if entry, known := t.entries[key]; known {
        if entry.inFlight {
            return false, inFlightRetry
        }
        if now.Before(entry.blockedUntil) {
            return false, entry.blockedUntil.Sub(now)
        }
        entry.inFlight = true
        return true, 0
    }
    if now.Sub(t.windowStart) >= t.window {
        t.windowStart = now
        t.newFailures = 0
    }
    if len(t.entries) >= t.maxKeys {
        t.expire(now)
        if (len(t.entries) >= t.maxKeys) && !t.evict(now) {
            return false, t.window
        }
    }
    overBudget := t.newFailures >= t.budget
    t.newFailures++
    t.entries[key] = &throttleEntry{inFlight: true, overBudget: overBudget}
    return true, 0
}

func (t *LoginThrottle) End(key netip.Prefix, success bool) {
    t.mu.Lock()
    defer t.mu.Unlock()
    entry, known := t.entries[key]
    if !known {
        return
    }
    entry.inFlight = false
    if success {
        if entry.failures == 0 {
            if t.newFailures > 0 {
                t.newFailures--
            }
            delete(t.entries, key)
        }
        return
    }
    entry.failures++
    entry.lastFailure = t.clock.Now()
    delay := t.initialDelay << min(entry.failures - 1, t.doublings)
    if (delay > t.maxDelay) || entry.overBudget {
        delay = t.maxDelay
    }
    entry.blockedUntil = t.clock.Now().Add(delay)
}

func (t *LoginThrottle) expire(now time.Time) {
    for key, entry := range t.entries {
        if !entry.inFlight && (now.Sub(entry.blockedUntil) >= t.maxDelay) {
            delete(t.entries, key)
        }
    }
}

func (t *LoginThrottle) evict(now time.Time) bool {
    var oldest netip.Prefix
    found := false
    for key, entry := range t.entries {
        if entry.inFlight || now.Before(entry.blockedUntil) {
            continue
        }
        if !found {
            oldest, found = key, true
            continue
        }
        candidate := t.entries[oldest]
        if (entry.overBudget && !candidate.overBudget) || ((entry.overBudget == candidate.overBudget) && entry.lastFailure.Before(candidate.lastFailure)) {
            oldest = key
        }
    }
    if found {
        delete(t.entries, oldest)
    }
    return found
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
