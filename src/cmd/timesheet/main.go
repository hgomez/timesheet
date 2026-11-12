/*
 * main.go - Copyright (c) 2023-2026 - Olivier Poncet
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

package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"

    "timesheet/internal/app"
)

// ---------------------------------------------------------------------------
// main implementation
// ---------------------------------------------------------------------------

func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    context.AfterFunc(ctx, stop)
    code := app.NewApplication(os.Args[1:], os.LookupEnv, os.Stdin, os.Stdout, os.Stderr, time.Local).WithSignals(stop).Run(ctx)
    stop()
    os.Exit(code)
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
