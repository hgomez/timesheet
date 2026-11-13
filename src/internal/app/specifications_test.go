/*
 * specifications_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "os"
    "path/filepath"
    "regexp"
    "slices"
    "strings"
    "testing"

    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// specifications tests
// ---------------------------------------------------------------------------

func TestCommandsMatchTheSpecifications(t *testing.T) {
    data, err := os.ReadFile(filepath.Join("..", "..", "..", "doc", "SPECIFICATIONS.md"))
    if err != nil {
        t.Fatal(err)
    }
    section := string(data)[strings.Index(string(data), "## COMMAND LINE"):]
    section = section[:strings.Index(section, "\n## ")]
    documented := []string{}
    for _, line := range strings.Split(section, "\n") {
        fields := strings.Fields(strings.TrimPrefix(line, ProgramName + " "))
        if !strings.HasPrefix(line, ProgramName + " ") || (len(fields) == 0) {
            continue
        }
        switch fields[0] {
            case "help":
                documented = append(documented, "help", "version")
            case "config", "user", "session":
                documented = append(documented, fields[0] + " " + fields[1])
            default:
                documented = append(documented, fields[0])
        }
    }
    known := []string{"help", "version", "init", "run", "config show", "config set", "user add", "user profile", "user passwd", "user del", "user list", "session clear", "import"}
    slices.Sort(documented)
    slices.Sort(known)
    if !slices.Equal(documented, known) {
        t.Fatalf("commands differ:\ndocumented %v\nknown %v", documented, known)
    }
    usage := &strings.Builder{}
    NewUsage(ProgramName).Write(usage)
    for _, command := range known {
        if !strings.Contains(usage.String(), "  " + command) {
            t.Errorf("usage does not list %s", command)
        }
    }
}

func TestServerKeysMatchTheSpecifications(t *testing.T) {
    data, err := os.ReadFile(filepath.Join("..", "..", "..", "doc", "SPECIFICATIONS.md"))
    if err != nil {
        t.Fatal(err)
    }
    section := string(data)[strings.Index(string(data), "## SERVER CONFIGURATION"):]
    section = section[:strings.Index(section, "\n## ")]
    documented := []string{}
    prefix := ""
    names := regexp.MustCompile("`([a-z_]+)`")
    for _, line := range strings.Split(section, "\n") {
        if title, found := strings.CutPrefix(line, "### Section "); found {
            prefix = title + "."
            continue
        }
        if strings.HasPrefix(line, "### ") {
            prefix = ""
        }
        cells := strings.Split(line, " | ")
        if (prefix == "") || !strings.HasPrefix(line, "| `") || (len(cells) < 2) {
            continue
        }
        for _, match := range names.FindAllStringSubmatch(cells[0], -1) {
            documented = append(documented, prefix + match[1])
        }
    }
    known := activity.NewDefaultServerConfiguration().Keys()
    slices.Sort(documented)
    slices.Sort(known)
    if !slices.Equal(documented, known) {
        t.Fatalf("keys differ:\ndocumented %v\nknown %v", documented, known)
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
