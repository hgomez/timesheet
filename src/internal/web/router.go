/*
 * router.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "slices"
    "strings"
)

// ---------------------------------------------------------------------------
// Router declaration
// ---------------------------------------------------------------------------

type Router struct {
    mux      *http.ServeMux
    errors   *ErrorPages
    handlers map[string]map[string]http.Handler
    order    []string
}

type methodDispatcher struct {
    router  *Router
    pattern string
}

// ---------------------------------------------------------------------------
// Router implementation
// ---------------------------------------------------------------------------

func NewRouter(fallback http.Handler, errors *ErrorPages) *Router {
    r := &Router{mux: http.NewServeMux(), errors: errors, handlers: map[string]map[string]http.Handler{}}
    r.mux.Handle("/", fallback)
    return r
}

func (r *Router) Handle(method string, pattern string, handler http.Handler) {
    if _, known := r.handlers[pattern]; !known {
        r.handlers[pattern] = map[string]http.Handler{}
        r.order = append(r.order, pattern)
        r.mux.Handle(pattern, &methodDispatcher{router: r, pattern: pattern})
    }
    r.handlers[pattern][method] = handler
}

func (r *Router) Routes() []Route {
    routes := []Route{}
    for _, pattern := range r.order {
        for method := range r.handlers[pattern] {
            routes = append(routes, Route{Method: method, Pattern: pattern})
        }
    }
    slices.SortFunc(routes, r.compare)
    return routes
}

func (r *Router) compare(a Route, b Route) int {
    if a.Pattern != b.Pattern {
        return strings.Compare(a.Pattern, b.Pattern)
    }
    return strings.Compare(a.Method, b.Method)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    r.mux.ServeHTTP(w, req)
}

func (d *methodDispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    handlers := d.router.handlers[d.pattern]
    method := r.Method
    if method == http.MethodHead {
        method = http.MethodGet
    }
    if handler, ok := handlers[method]; ok {
        handler.ServeHTTP(w, r)
        return
    }
    allowed := []string{}
    for name := range handlers {
        allowed = append(allowed, name)
        if name == http.MethodGet {
            allowed = append(allowed, http.MethodHead)
        }
    }
    slices.Sort(allowed)
    d.router.errors.MethodNotAllowed(w, r, allowed)
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
