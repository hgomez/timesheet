/*
 * asset_server.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "io"
    "io/fs"
    "net/http"
    "path"
    "strings"
    "time"
)

// ---------------------------------------------------------------------------
// AssetServer declaration
// ---------------------------------------------------------------------------

type AssetServer struct {
    static fs.FS
    errors *ErrorPages
}

// ---------------------------------------------------------------------------
// AssetServer implementation
// ---------------------------------------------------------------------------

func NewAssetServer(static fs.FS, errors *ErrorPages) *AssetServer {
    return &AssetServer{static: static, errors: errors}
}

func (s *AssetServer) ServeStatic(w http.ResponseWriter, r *http.Request) {
    s.serve(w, r, r.PathValue("path"))
}

func (s *AssetServer) serve(w http.ResponseWriter, r *http.Request, name string) {
    contentType, known := s.contentType(name)
    if (name == "") || !fs.ValidPath(name) || !known {
        s.errors.NotFound(w, r)
        return
    }
    info, err := fs.Stat(s.static, name)
    if (err != nil) || !info.Mode().IsRegular() {
        s.errors.NotFound(w, r)
        return
    }
    file, err := s.static.Open(name)
    if err != nil {
        s.errors.NotFound(w, r)
        return
    }
    defer file.Close()
    content, ok := file.(io.ReadSeeker)
    if !ok {
        s.errors.NotFound(w, r)
        return
    }
    w.Header().Set("Content-Type", contentType)
    http.ServeContent(w, r, "", time.Time{}, content)
}

func (s *AssetServer) contentType(name string) (string, bool) {
    switch strings.ToLower(path.Ext(name)) {
        case ".css":
            return "text/css; charset=utf-8", true
        case ".js":
            return "text/javascript; charset=utf-8", true
        case ".png":
            return "image/png", true
        case ".gif":
            return "image/gif", true
        case ".jpg", ".jpeg":
            return "image/jpeg", true
        case ".svg":
            return "image/svg+xml", true
    }
    return "", false
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
