/*
 * palette_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "io/fs"
    "regexp"
    "strings"
    "testing"

    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// palette tests
// ---------------------------------------------------------------------------

func TestBothPalettesKeepReadableContrasts(t *testing.T) {
    data, err := fs.ReadFile(NewEmbeddedAssets().Static(), "css/theme.css")
    if err != nil {
        t.Fatal(err)
    }
    text := string(data)
    blocks := map[string]string{
        "light": text[strings.Index(text, ":root {"):strings.Index(text, `:root[data-theme="dark"] {`)],
        "dark":  text[strings.Index(text, `:root[data-theme="dark"] {`):strings.Index(text, "@media (prefers-color-scheme: dark)")],
    }
    system := text[strings.Index(text, "@media (prefers-color-scheme: dark)"):strings.Index(text, "*, *::before, *::after {")]
    declaration := regexp.MustCompile(`(--[a-z-]+): (#[0-9a-f]{6});`)
    if dark := declaration.FindAllString(blocks["dark"], -1); strings.Join(dark, "") != strings.Join(declaration.FindAllString(system, -1), "") {
        t.Error("the system dark palette must repeat the explicit dark palette")
    }
    pairs := [][2]string{
        {"--ink", "--background"}, {"--ink", "--surface"}, {"--ink", "--surface-raised"}, {"--ink", "--surface-sunken"}, {"--ink", "--footer"},
        {"--ink-soft", "--background"}, {"--ink-soft", "--surface"}, {"--ink-soft", "--surface-sunken"}, {"--ink-soft", "--footer"},
        {"--accent", "--surface"}, {"--accent", "--background"}, {"--accent", "--accent-soft"}, {"--accent-hover", "--surface"}, {"--on-accent", "--accent"}, {"--on-accent", "--accent-hover"},
        {"--success", "--success-soft"}, {"--warning", "--warning-soft"}, {"--danger", "--danger-soft"}, {"--danger", "--surface"}, {"--on-danger", "--danger"},
        {"--bar-ink", "--bar"}, {"--bar-ink-soft", "--bar"}, {"--bar-ink", "--bar-hover"},
    }
    for name, block := range blocks {
        tokens := map[string]activity.Color{}
        for _, match := range declaration.FindAllStringSubmatch(block, -1) {
            tokens[match[1]] = activity.Color(match[2])
        }
        for _, pair := range pairs {
            ink, inkFound := tokens[pair[0]]
            ground, groundFound := tokens[pair[1]]
            if !inkFound || !groundFound {
                t.Errorf("%s: missing %s or %s", name, pair[0], pair[1])
                continue
            }
            if contrast := ink.Contrast(ground); contrast < 4.5 {
                t.Errorf("%s: %s on %s has a contrast of %.2f", name, pair[0], pair[1], contrast)
            }
        }
    }
}

func TestButtonsKeepTheirInkOnTheAccent(t *testing.T) {
    data, err := fs.ReadFile(NewEmbeddedAssets().Static(), "css/components.css")
    if err != nil {
        t.Fatal(err)
    }
    rules := regexp.MustCompile(`(?s)([^{}]+)\{([^{}]*)\}`)
    for _, rule := range rules.FindAllStringSubmatch(string(data), -1) {
        if !strings.Contains(rule[2], "color: inherit") {
            continue
        }
        for _, selector := range strings.Split(rule[1], ",") {
            if strings.TrimSpace(selector) == "button" {
                t.Errorf("rule %q resets the ink of every button", strings.TrimSpace(rule[1]))
            }
        }
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
