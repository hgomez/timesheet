/*
 * workspace_loader.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "net/http"
    "strconv"
    "time"

    "timesheet/internal/account"
    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// WorkspaceLoader declaration
// ---------------------------------------------------------------------------

type WorkspaceLoader struct {
    repository Repository
    today      *Today
}

// ---------------------------------------------------------------------------
// WorkspaceLoader implementation
// ---------------------------------------------------------------------------

func NewWorkspaceLoader(repository Repository, today *Today) *WorkspaceLoader {
    return &WorkspaceLoader{repository: repository, today: today}
}

func (l *WorkspaceLoader) Own(r *http.Request) (*Workspace, bool) {
    viewer := (&Viewer{}).From(r)
    if viewer == nil {
        return nil, false
    }
    return l.load(viewer, viewer.Login, "/dashboard", false)
}

func (l *WorkspaceLoader) Other(r *http.Request) (*Workspace, bool) {
    viewer := (&Viewer{}).From(r)
    login := account.Login(r.PathValue("login"))
    if (viewer == nil) || !login.Valid() {
        return nil, false
    }
    return l.load(viewer, login, "/admin/accounts/" + string(login), true)
}

func (l *WorkspaceLoader) load(viewer *Viewer, login account.Login, base string, readOnly bool) (*Workspace, bool) {
    owner, calendar, found := l.repository.Account(login)
    if !found {
        return nil, false
    }
    config := l.repository.Configuration()
    return &Workspace{Viewer: viewer, Owner: owner, Calendar: calendar, Config: config, Book: activity.NewBookkeeper(config, calendar), Today: l.today.Date(), Location: l.today.Location(), Base: base, ReadOnly: readOnly}, true
}

func (l *WorkspaceLoader) Month(r *http.Request) (activity.Month, bool) {
    year, ok := l.Year(r)
    number, err := strconv.Atoi(r.PathValue("month"))
    if !ok || (err != nil) || (number < 1) || (number > 12) {
        return activity.Month{}, false
    }
    m := activity.NewMonth(year, time.Month(number))
    if (r.PathValue("year") + "/" + r.PathValue("month")) != m.Path() {
        return activity.Month{}, false
    }
    return m, true
}

func (l *WorkspaceLoader) Year(r *http.Request) (int, bool) {
    year, err := strconv.Atoi(r.PathValue("year"))
    return year, (err == nil) && (year >= activity.MinYear) && (year <= activity.MaxYear) && (r.PathValue("year") == strconv.Itoa(year))
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
