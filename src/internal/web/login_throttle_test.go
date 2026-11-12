/*
 * login_throttle_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "testing"
    "time"

    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// LoginThrottle tests
// ---------------------------------------------------------------------------

func TestThrottleBudgetNeverBlocksAFirstAttempt(t *testing.T) {
    clock := activity.NewManualClock(time.Date(2026, time.November, 8, 10, 0, 0, 0, time.UTC))
    throttle := NewLoginThrottle(clock, 64, 3)
    for i := 1; i <= 5; i++ {
        key := netip.MustParsePrefix("10.0.0." + string(rune('0' + i)) + "/32")
        if allowed, _ := throttle.Begin(key); !allowed {
            t.Fatalf("attempt %d refused by the global budget", i)
        }
        throttle.End(key, false)
    }
    over := netip.MustParsePrefix("10.0.0.5/32")
    if allowed, wait := throttle.Begin(over); allowed || (wait < 4 * time.Minute) {
        t.Fatalf("an address failing over budget must wait the longest delay: %v %v", allowed, wait)
    }
    legit := netip.MustParsePrefix("10.0.1.1/32")
    if allowed, _ := throttle.Begin(legit); !allowed {
        t.Fatal("a fresh address must still be able to log in")
    }
    throttle.End(legit, true)
}

func TestThrottleSuccessKeepsEarlierFailures(t *testing.T) {
    clock := activity.NewManualClock(time.Date(2026, time.November, 8, 10, 0, 0, 0, time.UTC))
    throttle := NewLoginThrottle(clock, 64, 100)
    key := netip.MustParsePrefix("10.0.0.9/32")
    for range 3 {
        throttle.Begin(key)
        throttle.End(key, false)
        clock.Advance(10 * time.Minute)
    }
    throttle.Begin(key)
    throttle.End(key, true)
    clock.Advance(10 * time.Minute)
    throttle.Begin(key)
    throttle.End(key, false)
    if _, wait := throttle.Begin(key); wait < 8 * time.Second {
        t.Fatalf("a success must not reset the delay: %v", wait)
    }
    clock.Advance(2 * time.Hour)
    if allowed, _ := throttle.Begin(key); !allowed {
        t.Fatal("old failures must be forgotten")
    }
    throttle.End(key, false)
    if _, wait := throttle.Begin(key); wait > time.Second {
        t.Fatalf("a single typo after a long time must cost one second: %v", wait)
    }
}

func TestThrottleTableFullEvictsInsteadOfRefusing(t *testing.T) {
    clock := activity.NewManualClock(time.Date(2026, time.November, 8, 10, 0, 0, 0, time.UTC))
    throttle := NewLoginThrottle(clock, 4, 1)
    for i := 1; i <= 4; i++ {
        key := netip.MustParsePrefix("10.0.2." + string(rune('0' + i)) + "/32")
        throttle.Begin(key)
        throttle.End(key, false)
    }
    if allowed, _ := throttle.Begin(netip.MustParsePrefix("10.0.3.1/32")); allowed {
        t.Fatal("a blocked entry must never be evicted")
    }
    clock.Advance(6 * time.Minute)
    if allowed, _ := throttle.Begin(netip.MustParsePrefix("10.0.3.1/32")); !allowed {
        t.Fatal("once the blocks are over, a full table must evict an entry, not refuse a fresh address")
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
