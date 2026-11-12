/*
 * source_policy_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "go/ast"
    "go/parser"
    "go/token"
    "io/fs"
    "os"
    "path/filepath"
    "regexp"
    "strconv"
    "strings"
    "testing"
)

// ---------------------------------------------------------------------------
// source policy tests
// ---------------------------------------------------------------------------

func TestNoTrustedTemplateContent(t *testing.T) {
    forbidden := map[string]bool{"HTML": true, "HTMLAttr": true, "JS": true, "JSStr": true, "URL": true, "CSS": true, "Srcset": true}
    err := filepath.WalkDir(filepath.Join("..", ".."), func(path string, entry fs.DirEntry, err error) error {
        if (err != nil) || entry.IsDir() || !strings.HasSuffix(path, ".go") {
            return err
        }
        file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
        if err != nil {
            return err
        }
        names := map[string]bool{}
        for _, spec := range file.Imports {
            imported, _ := strconv.Unquote(spec.Path.Value)
            if imported == "text/template" {
                t.Errorf("%s imports text/template", path)
            }
            if imported == "html/template" {
                names["template"] = true
            }
        }
        ast.Inspect(file, func(node ast.Node) bool {
            if selector, ok := node.(*ast.SelectorExpr); ok {
                if ident, ok := selector.X.(*ast.Ident); ok && names[ident.Name] && forbidden[selector.Sel.Name] {
                    t.Errorf("%s uses %s.%s", path, ident.Name, selector.Sel.Name)
                }
            }
            return true
        })
        return nil
    })
    if err != nil {
        t.Fatal(err)
    }
}

func TestGoFilesHaveOnlyAllowedComments(t *testing.T) {
    allowed := regexp.MustCompile(`^//go:(build|embed)|^// -{10,}$|^// .+ (declaration|implementation|tests)( \(.+\))?$|^// (End-Of-File|fixtures)$`)
    err := filepath.WalkDir(filepath.Join("..", ".."), func(path string, entry fs.DirEntry, err error) error {
        if (err != nil) || entry.IsDir() || !strings.HasSuffix(path, ".go") {
            return err
        }
        data, err := os.ReadFile(path)
        if err != nil {
            return err
        }
        for number, line := range strings.Split(string(data), "\n") {
            trimmed := strings.TrimSpace(line)
            if strings.HasPrefix(trimmed, "//") && !allowed.MatchString(trimmed) {
                t.Errorf("%s:%d: forbidden comment %q", path, number + 1, trimmed)
            }
            if strings.HasPrefix(line, "\t") {
                t.Errorf("%s:%d: indented with a tab", path, number + 1)
            }
        }
        return nil
    })
    if err != nil {
        t.Fatal(err)
    }
}

func licenceHeader(t *testing.T, name string) string {
    t.Helper()
    data, err := os.ReadFile("source_policy_test.go")
    if err != nil {
        t.Fatal(err)
    }
    header, _, found := strings.Cut(string(data), " */\n")
    if !found {
        t.Fatal("no licence header in source_policy_test.go")
    }
    return strings.Replace(header, "source_policy_test.go", name, 1) + " */"
}

func withoutLicence(t *testing.T, file string, text string, opening string, closing string) string {
    t.Helper()
    rest, found := strings.CutPrefix(text, opening + licenceHeader(t, filepath.Base(file)) + closing + "\n")
    if !found {
        t.Errorf("%s does not start with its licence header", file)
    }
    return rest
}

func TestWebSourcesStartWithTheLicenceHeader(t *testing.T) {
    static := NewEmbeddedAssets().Static()
    templates := NewEmbeddedAssets().Templates()
    sheets, _ := fs.Glob(static, "css/*.css")
    scripts, _ := fs.Glob(static, "js/*.js")
    views, _ := fs.Glob(templates, "*.html")
    partials, _ := fs.Glob(templates, "partials/*.html")
    if (len(sheets) != 2) || (len(scripts) == 0) || (len(views) < 10) || (len(partials) != 3) {
        t.Fatalf("sheets %d, scripts %d, views %d, partials %d", len(sheets), len(scripts), len(views), len(partials))
    }
    for _, file := range append(sheets, scripts...) {
        data, _ := fs.ReadFile(static, file)
        withoutLicence(t, file, string(data), "", "\n")
    }
    for _, file := range append(views, partials...) {
        data, _ := fs.ReadFile(templates, file)
        withoutLicence(t, file, string(data), "{{- ", " -}}")
    }
}

func TestJavaScriptHasNoSink(t *testing.T) {
    comments := regexp.MustCompile(`(?s)/\*.*?\*/|//[^\n]*`)
    sinks := regexp.MustCompile(`innerHTML|outerHTML|insertAdjacentHTML|document\.write|\beval\s*\(|\bFunction\s*\(|javascript:|setAttribute\s*\(\s*['"](style|on)`)
    static := NewEmbeddedAssets().Static()
    files, _ := fs.Glob(static, "js/*.js")
    if len(files) == 0 {
        t.Fatal("no script found")
    }
    for _, file := range files {
        data, _ := fs.ReadFile(static, file)
        text := withoutLicence(t, file, string(data), "", "\n")
        if match := sinks.FindString(comments.ReplaceAllString(text, "")); match != "" {
            t.Errorf("%s uses %s", file, match)
        }
        if comments.MatchString(text) {
            t.Errorf("%s has a comment besides its licence header", file)
        }
    }
}

func TestTemplatesAreSemanticAndInert(t *testing.T) {
    templates := NewEmbeddedAssets().Templates()
    views, _ := fs.Glob(templates, "*.html")
    partials, _ := fs.Glob(templates, "partials/*.html")
    if (len(views) < 10) || (len(partials) != 3) {
        t.Fatalf("views %d, partials %d", len(views), len(partials))
    }
    forbidden := regexp.MustCompile(`(?i)<div\b|<style\b|[\s"'/]style\s*=|[\s"'/]id\s*=|[\s"'/]for\s*=|[\s"'/]on[a-z]+\s*=|<!--|\{\{-?\s*/\*|javascript:|<script(\s+type="module")?>[^<]`)
    scripts := regexp.MustCompile(`(?i)<script\b[^>]*>`)
    for _, file := range append(views, partials...) {
        data, _ := fs.ReadFile(templates, file)
        text := withoutLicence(t, file, string(data), "{{- ", " -}}")
        if match := forbidden.FindString(text); match != "" {
            t.Errorf("%s contains %q", file, match)
        }
        for _, tag := range scripts.FindAllString(text, -1) {
            if (tag != `<script type="application/json">`) && !regexp.MustCompile(`^<script type="module" src="/static/js/[a-z_]+_page\.js">$`).MatchString(tag) {
                t.Errorf("%s: script tag %s", file, tag)
            }
        }
        if strings.HasPrefix(file, "partials/") || (file == "layout.html") {
            continue
        }
        if count := len(regexp.MustCompile(`(?i)<h1[\s>]`).FindAllString(text, -1)); count != 1 {
            t.Errorf("%s defines %d h1", file, count)
        }
    }
}

func TestStylesheetsAvoidImportantAndLiteralSpacing(t *testing.T) {
    static := NewEmbeddedAssets().Static()
    files, _ := fs.Glob(static, "css/*.css")
    if len(files) != 2 {
        t.Fatalf("expected two style sheets, got %v", files)
    }
    for _, file := range files {
        data, _ := fs.ReadFile(static, file)
        text := withoutLicence(t, file, string(data), "", "\n")
        if strings.Contains(text, "!important") {
            t.Errorf("%s uses !important", file)
        }
        if strings.Contains(text, "/*") {
            t.Errorf("%s has a comment besides its licence header", file)
        }
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
