/*
 * application_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "bytes"
    "context"
    "crypto/rand"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "testing"
    "time"

    "timesheet/internal/account"
    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// Application tests
// ---------------------------------------------------------------------------

func TestUsageAndArguments(t *testing.T) {
    if run := runCommand(t, "", "help"); (run.code != ExitOK) || !strings.Contains(run.stdout, "import LOGIN FILE") {
        t.Fatalf("help: %+v", run)
    }
    for _, args := range [][]string{{"bogus"}, {"user"}, {"user", "add"}, {"run", "--profile=ADMIN"}, {"run", "--tls-cert=x"}, {"import", "a"}, {"run", "--trusted-proxy=nope"}, {"import", "a", "b", "--from=2026-10", "--to=2026-09"}, {"user", "list", "--database="}, {"user", "list", "--database=dir/"}} {
        if run := runCommand(t, "", args...); run.code != ExitUsage {
            t.Errorf("%v: code %d", args, run.code)
        }
    }
    if run := runCommand(t, "", "version"); !strings.HasPrefix(run.stdout, "timesheet ") {
        t.Fatalf("version: %q", run.stdout)
    }
}

func TestInitAndAccounts(t *testing.T) {
    directory, database := newTestDirectory(t)
    config := "--config=" + filepath.Join(directory, "timesheet.json")
    template := "--template=" + filepath.Join("..", "..", "..", "share", "timesheet", "timesheet-database-template.json")
    if run := runCommand(t, "", "init", template, database); run.code != ExitOK {
        t.Fatalf("init: %+v", run)
    }
    if run := runCommand(t, "", "init", template, database); run.code != ExitError {
        t.Fatal("a second init must fail")
    }
    if run := runCommand(t, "", "run", database); (run.code != ExitError) || !strings.Contains(run.stderr, "--profile=ADMIN") {
        t.Fatalf("run without admin: %+v", run)
    }
    if run := runCommand(t, "short", "user", "add", "root", "--profile=ADMIN", database); run.code != ExitError {
        t.Fatal("weak password accepted")
    }
    if run := runCommand(t, "correct horse battery", "user", "add", "root", "--profile=ADMIN", database); run.code != ExitOK {
        t.Fatalf("user add: %+v", run)
    }
    if run := runCommand(t, "", "user", "profile", "root", "USER", database); (run.code != ExitError) || !strings.Contains(run.stderr, "last administrator") {
        t.Fatalf("demote the last admin: %+v", run)
    }
    if run := runCommand(t, "", "user", "list", database); !strings.Contains(run.stdout, "root") || !strings.Contains(run.stdout, "ADMIN") || !strings.Contains(run.stdout, "ÉTAT") || !strings.Contains(run.stdout, "enabled") {
        t.Fatalf("list: %+v", run)
    }
    if run := runCommand(t, "", "config", "set", "server.timezone", "Mars/Olympus", database, config); run.code != ExitError {
        t.Fatal("invalid time zone accepted")
    }
    if run := runCommand(t, "", "config", "set", "server.origins", "http://a.test:8080, http://b.test:8080", database, config); run.code != ExitOK {
        t.Fatalf("config set: %+v", run)
    }
    if run := runCommand(t, "", "config", "show", database, config); !strings.Contains(run.stdout, "http://b.test:8080") {
        t.Fatalf("config show: %s", run.stdout)
    }
    _ = directory
}

func TestImportReadsTheJSONFile(t *testing.T) {
    directory, database := newTestDirectory(t)
    template := "--template=" + filepath.Join("..", "..", "..", "share", "timesheet", "timesheet-database-template.json")
    runCommand(t, "", "init", template, database)
    runCommand(t, "correct horse battery", "user", "add", "olivier", database)
    file := filepath.Join(directory, "activite.json")
    writeTestImport(t, file)
    run := runCommand(t, "", "import", "olivier", file, "--dry-run", database)
    if (run.code != ExitOK) || !strings.Contains(run.stdout, "Réglages lus : début du suivi 2025-01-01, 2 compteur(s).") || !strings.Contains(run.stdout, "journée(s) lue(s), de 2025-01 à 2025-01") || !strings.Contains(run.stdout, "1 ajustement(s) lu(s).") || !strings.Contains(run.stdout, "Simulation") {
        t.Fatalf("dry run: %+v", run)
    }
    if _, calendar, _ := openTestDatabase(t, database).Account("olivier"); len(calendar.Days) != 0 {
        t.Fatal("a dry run must not write")
    }
    if run := runCommand(t, "", "import", "olivier", file, "--dry-run", "--validate", database); (run.code != ExitOK) || !strings.Contains(run.stdout, "2025-01 validé") || !strings.Contains(run.stdout, "Simulation") {
        t.Fatalf("dry run with validation: %+v", run)
    }
    if _, calendar, _ := openTestDatabase(t, database).Account("olivier"); len(calendar.Months) != 0 {
        t.Fatal("a dry run must not validate")
    }
    run = runCommand(t, "", "import", "olivier", file, "--validate", database)
    if (run.code != ExitOK) || !strings.Contains(run.stdout, "2025-01 validé") || !strings.Contains(run.stdout, "Import enregistré") {
        t.Fatalf("import: %+v", run)
    }
    _, calendar, _ := openTestDatabase(t, database).Account("olivier")
    if (calendar.Settings.Start.String() != "2025-01-01") || (calendar.Settings.Counters["CP"].Initial.String() != "2.29") || (calendar.Settings.Counters["TPC"].Amount.String() != "7/12") || (calendar.Days[activity.NewDate(2025, time.January, 7)].Comment != "Vacances") || (len(calendar.Adjustments) != 1) || (calendar.Adjustments[0].ID == "") || !calendar.IsValidated(activity.NewMonth(2025, time.January)) || (activity.NewDateOf(calendar.Months[activity.NewMonth(2025, time.January)].ValidatedAt) != activity.NewDate(2025, time.January, 31)) {
        t.Fatalf("imported calendar: %+v", calendar.Settings)
    }
    if run := runCommand(t, "", "import", "olivier", file, "--replace", database); (run.code != ExitError) || !strings.Contains(run.stdout, "2025-01-02 : le mois est validé") {
        t.Fatalf("import over a validated month: %+v", run)
    }
    moved := filepath.Join(directory, "moved.json")
    writeTestFile(t, moved, `{"format": "timesheet-import", "version": 1, "settings": {"start": "2024-06-01", "counters": {"CP": {"initial": "1"}}}, "days": {}}`)
    if run := runCommand(t, "", "import", "olivier", moved, database); (run.code != ExitError) || !strings.Contains(run.stdout, "début du suivi : un mois est déjà validé") || !strings.Contains(run.stdout, "Congés payés : un mois est déjà validé") {
        t.Fatalf("settings over a validated month: %+v", run)
    }
}

func TestImportRefusesInvalidFiles(t *testing.T) {
    directory, database := newTestDirectory(t)
    template := "--template=" + filepath.Join("..", "..", "..", "share", "timesheet", "timesheet-database-template.json")
    runCommand(t, "", "init", template, database)
    runCommand(t, "correct horse battery", "user", "add", "olivier", database)
    header := `{"format": "timesheet-import", "version": 1, `
    cases := map[string]string{
        header + `"days": {}, "extra": true}`: "unknown field",
        `{"format": "other", "version": 1, "days": {}}`: "format",
        `{"format": "timesheet-import", "version": 2, "days": {}}`: "version",
        header + `"days": {"2026-11-03": {"am": "NOPE", "pm": "", "comment": ""}}}`: "2026-11-03 : Matin : code d'activité non disponible",
        header + `"days": {"2026-11-07": {"am": "CP", "pm": "CP", "comment": ""}}}`: "2026-11-07 : Matin : « Congés payés » est décompté",
        header + `"days": {"2026-11-11": {"am": "CP", "pm": "", "comment": ""}}}`: "2026-11-11 : Matin : « Congés payés » est décompté",
        header + `"days": {"2029-01-03": {"am": "TT", "pm": "TT", "comment": ""}}}`: "au-delà de l'horizon",
        header + `"settings": {"counters": {"NOPE": {"initial": "0"}}}, "days": {}}`: "compteur \"NOPE\" inconnu",
        header + `"settings": {"counters": {"CP": {"amount": "400", "initial": "0"}}}, "days": {}}`: "montant d'acquisition invalide",
        header + `"days": {}} {}`: "data after the object",
        header + `"days": {}}`: "aucune journée à importer",
        header + `"days": {}, "adjustments": [{"date": "2026-11-03", "counter": "TPC", "bucket": "current", "amount": "1", "reason": "bulletin"}]}`: "unknown field",
        header + `"days": {}, "adjustments": [{"date": "2026-11-03", "counter": "NOPE", "amount": "1", "reason": "bulletin"}]}`: "compteur \"NOPE\" inconnu ou archivé",
        header + `"days": {}, "adjustments": [{"date": "2026-11-03", "counter": "CP", "amount": "0", "reason": "bulletin"}]}`: "quantité non nulle",
        header + `"days": {}, "adjustments": [{"date": "2026-11-03", "counter": "CP", "amount": "1", "reason": ""}]}`: "motif obligatoire",
        header + `"days": {}, "adjustments": [{"date": "2029-01-03", "counter": "CP", "amount": "1", "reason": "bulletin"}]}`: "dans un mois ouvert, à partir du début du suivi",
    }
    for content, expected := range cases {
        file := filepath.Join(directory, "refused.json")
        writeTestFile(t, file, content)
        if run := runCommand(t, "", "import", "olivier", file, database); (run.code != ExitError) || !strings.Contains(run.stdout + run.stderr, expected) {
            t.Errorf("%s: %+v", content, run)
        }
    }
    large := filepath.Join(directory, "large.json")
    writeTestFile(t, large, header + `"days": {}, "pad": "` + strings.Repeat("x", MaxImportSize) + `"}`)
    if run := runCommand(t, "", "import", "olivier", large, database); (run.code != ExitError) || !strings.Contains(run.stderr, "larger than 8 MiB") {
        t.Errorf("large file: %+v", run)
    }
    if run := runCommand(t, "", "import", "olivier", large, "--mapping=share/import-ods.json", database); run.code != ExitUsage {
        t.Errorf("--mapping: %+v", run)
    }
    if _, calendar, _ := openTestDatabase(t, database).Account("olivier"); len(calendar.Days) != 0 {
        t.Fatal("a refused import must not write")
    }
}

func TestImportReplacesOnlyWhenAsked(t *testing.T) {
    directory, database := newTestDirectory(t)
    template := "--template=" + filepath.Join("..", "..", "..", "share", "timesheet", "timesheet-database-template.json")
    runCommand(t, "", "init", template, database)
    runCommand(t, "correct horse battery", "user", "add", "olivier", database)
    file := filepath.Join(directory, "day.json")
    writeTestFile(t, file, `{"format": "timesheet-import", "version": 1, "days": {"2026-11-03": {"am": "TT", "pm": "TT", "comment": "salon"}}, "adjustments": [{"date": "2026-11-03", "counter": "TPC", "amount": "-0.5", "reason": "bulletin"}]}`)
    if run := runCommand(t, "", "import", "olivier", file, database); run.code != ExitOK {
        t.Fatalf("first import: %+v", run)
    }
    if run := runCommand(t, "", "import", "olivier", file, database); (run.code != ExitError) || !strings.Contains(run.stdout, "une saisie existe déjà") {
        t.Fatalf("second import: %+v", run)
    }
    if run := runCommand(t, "", "import", "olivier", file, "--replace", database); (run.code != ExitOK) || !strings.Contains(run.stdout, "ajustement du 2026-11-03 déjà présent, ignoré") {
        t.Fatalf("replace: %+v", run)
    }
    if _, calendar, _ := openTestDatabase(t, database).Account("olivier"); len(calendar.Adjustments) != 1 {
        t.Fatalf("adjustments after a second import: %+v", calendar.Adjustments)
    }
}

func TestImportKeepsAdjustmentsInsideTheTracking(t *testing.T) {
    directory, database := newTestDirectory(t)
    template := "--template=" + filepath.Join("..", "..", "..", "share", "timesheet", "timesheet-database-template.json")
    runCommand(t, "", "init", template, database)
    runCommand(t, "correct horse battery", "user", "add", "olivier", database)
    file := filepath.Join(directory, "adjusted.json")
    writeTestFile(t, file, `{"format": "timesheet-import", "version": 1, "settings": {"start": "2026-01-01"}, "days": {}, "adjustments": [{"date": "2026-01-15", "counter": "CP", "amount": "5", "reason": "bulletin"}]}`)
    if run := runCommand(t, "", "import", "olivier", file, database); run.code != ExitOK {
        t.Fatalf("first import: %+v", run)
    }
    partial := filepath.Join(directory, "partial.json")
    writeTestFile(t, partial, `{"format": "timesheet-import", "version": 1, "settings": {"counters": {"CP": {"initial": "2.29"}}}, "days": {}}`)
    runCommand(t, "", "import", "olivier", partial, database)
    writeTestFile(t, partial, `{"format": "timesheet-import", "version": 1, "settings": {"counters": {"CP": {"enabled": false}}}, "days": {}}`)
    if run := runCommand(t, "", "import", "olivier", partial, database); run.code != ExitOK {
        t.Fatalf("partial settings: %+v", run)
    }
    if _, calendar, _ := openTestDatabase(t, database).Account("olivier"); calendar.Settings.Counters["CP"].Initial.String() != "2.29" {
        t.Fatalf("an absent initial balance must be kept: %s", calendar.Settings.Counters["CP"].Initial)
    }
    moved := filepath.Join(directory, "moved.json")
    writeTestFile(t, moved, `{"format": "timesheet-import", "version": 1, "settings": {"start": "2026-03-01"}, "days": {}}`)
    if run := runCommand(t, "", "import", "olivier", moved, database); (run.code != ExitError) || !strings.Contains(run.stdout, "des ajustements sont datés avant 2026-03") {
        t.Fatalf("start moved after an adjustment: %+v", run)
    }
}

func TestUserCommandsCloseTheSessionsOfTheAccount(t *testing.T) {
    directory, database := newTestDirectory(t)
    template := "--template=" + filepath.Join("..", "..", "..", "share", "timesheet", "timesheet-database-template.json")
    runCommand(t, "", "init", template, database)
    runCommand(t, "correct horse battery", "user", "add", "root", "--profile=ADMIN", database)
    sessions := filepath.Join(directory, ".ts.json.sessions")
    record := func(key string, login string) string {
        return `{"key": "` + strings.Repeat(key, 64) + `", "login": "` + login + `", "csrf": "t", "created": "2026-11-11T21:00:00Z", "seen": "2026-11-11T21:00:00Z"}`
    }
    for _, command := range [][]string{{"user", "passwd", "bob"}, {"user", "profile", "bob", "ADMIN"}, {"user", "del", "bob"}} {
        runCommand(t, "correct horse battery", "user", "add", "bob", database)
        writeTestFile(t, sessions, "[" + record("a", "root") + ", " + record("b", "bob") + "]")
        if run := runCommand(t, "another horse battery", append(command, database)...); run.code != ExitOK {
            t.Fatalf("%v: %+v", command, run)
        }
        if data, _ := os.ReadFile(sessions); strings.Contains(string(data), `"bob"`) || !strings.Contains(string(data), `"root"`) {
            t.Fatalf("%v must close the sessions of bob only: %s", command, data)
        }
        if command[1] != "del" {
            runCommand(t, "", "user", "del", "bob", database)
        }
    }
}

func TestVersionIsTheReleaseOfTheChangelog(t *testing.T) {
    info := NewBuildInfo()
    if (info.Version() != ReleaseVersion) || !strings.HasPrefix(info.String(), "timesheet " + ReleaseVersion) {
        t.Fatalf("version %q, description %q, want %q", info.Version(), info.String(), ReleaseVersion)
    }
    if run := runCommand(t, "", "version"); !strings.HasPrefix(run.stdout, "timesheet " + ReleaseVersion) {
        t.Fatalf("version command: %s", run.stdout)
    }
    data, err := os.ReadFile(filepath.Join("..", "..", "..", "debian", "changelog"))
    if err != nil {
        t.Fatal(err)
    }
    entry := regexp.MustCompile(`^timesheet \(([^)]+)\)`).FindSubmatch(data)
    if entry == nil {
        t.Fatal("no version in the first entry of debian/changelog")
    }
    if packaged := string(entry[1]); (packaged != ReleaseVersion) && !strings.HasPrefix(packaged, ReleaseVersion + "+") {
        t.Fatalf("debian/changelog announces %s, the program %s", packaged, ReleaseVersion)
    }
}

func TestSessionClearEmptiesTheSessionFile(t *testing.T) {
    directory, database := newTestDirectory(t)
    template := "--template=" + filepath.Join("..", "..", "..", "share", "timesheet", "timesheet-database-template.json")
    runCommand(t, "", "init", template, database)
    sessions := filepath.Join(directory, ".ts.json.sessions")
    writeTestFile(t, sessions, `[{"key": "` + strings.Repeat("ab", 32) + `", "login": "olivier", "csrf": "t", "created": "2026-11-11T21:00:00Z", "seen": "2026-11-11T21:00:00Z"}]`)
    if run := runCommand(t, "", "session", "clear", database); (run.code != ExitOK) || !strings.Contains(run.stdout, "Toutes les sessions sont fermées.") {
        t.Fatalf("session clear: %+v", run)
    }
    if data, _ := os.ReadFile(sessions); strings.TrimSpace(string(data)) != "[]" {
        t.Fatalf("session file: %s", data)
    }
}

func TestRunServesTheSite(t *testing.T) {
    _, database := newTestDirectory(t)
    template := "--template=" + filepath.Join("..", "..", "..", "share", "timesheet", "timesheet-database-template.json")
    runCommand(t, "", "init", template, database)
    runCommand(t, "correct horse battery", "user", "add", "root", "--profile=ADMIN", database)
    stderr := &bytes.Buffer{}
    application := NewApplication([]string{"run", "--listen=127.0.0.1:0", database}, testEnv(""), strings.NewReader(""), io.Discard, stderr, time.UTC).WithIterations(account.MinIterations).WithClock(activity.NewSystemClock())
    ctx, cancel := context.WithCancel(context.Background())
    done := make(chan int, 1)
    go func() {
        done <- application.Run(ctx)
    }()
    select {
        case <-application.Ready():
        case code := <-done:
            t.Fatalf("run stopped: %d %s", code, stderr.String())
    }
    response, err := http.Get("http://" + application.Listen() + "/")
    if err != nil {
        t.Fatal(err)
    }
    response.Body.Close()
    if response.StatusCode != http.StatusOK {
        t.Fatalf("home: %d", response.StatusCode)
    }
    if run := runCommand(t, "correct horse battery", "user", "add", "late", database); (run.code != ExitError) || !strings.Contains(run.stderr, "stop the server") {
        t.Fatalf("user add while running: %+v", run)
    }
    cancel()
    if code := <-done; code != ExitOK {
        t.Fatalf("exit code %d: %s", code, stderr.String())
    }
    for _, secret := range []string{"correct horse battery"} {
        if strings.Contains(stderr.String(), secret) {
            t.Fatal("a secret reached the log")
        }
    }
}

func TestCommandEdgeCases(t *testing.T) {
    directory, database := newTestDirectory(t)
    config := "--config=" + filepath.Join(directory, "timesheet.json")
    template := filepath.Join("..", "..", "..", "share", "timesheet", "timesheet-database-template.json")
    trailing := filepath.Join(directory, "template.json")
    data, _ := os.ReadFile(template)
    os.WriteFile(trailing, append(data, []byte("{}")...), 0o600)
    if run := runCommand(t, "", "init", "--template=" + trailing, database); run.code != ExitError {
        t.Fatalf("template with trailing data: %+v", run)
    }
    runCommand(t, "", "init", "--template=" + template, database)
    if run := runCommand(t, "", "config", "set", "server.bogus", "x", database, config); run.code != ExitUsage {
        t.Fatalf("unknown key: %+v", run)
    }
    for _, value := range [][]string{{"server.listen", "nonsense"}, {"server.origins", "javascript:alert(1)"}} {
        if run := runCommand(t, "", "config", "set", value[0], value[1], database, config); run.code != ExitError {
            t.Fatalf("config set %v: %+v", value, run)
        }
    }
    if run := runCommand(t, "", "user", "add", "root", "--profile=ADMIN", database); run.code != ExitError {
        t.Fatalf("empty input accepted: %+v", run)
    }
    stderr := &bytes.Buffer{}
    env := func(name string) (string, bool) { return "", name == PasswordVariable }
    code := NewApplication([]string{"user", "add", "root", database}, env, strings.NewReader("correct horse battery\n"), io.Discard, stderr, time.UTC).WithIterations(account.MinIterations).Run(context.Background())
    if (code != ExitError) || !strings.Contains(stderr.String(), "empty") {
        t.Fatalf("empty TIMESHEET_PASSWORD: %d %s", code, stderr.String())
    }
    runCommand(t, "correct horse battery", "user", "add", "root", "--profile=ADMIN", database)
    if run := runCommand(t, "", "run", "--listen=127.0.0.1:0", "--origin=http://127.0.0.1:1", "--tls-cert=a.pem", "--tls-key=b.pem", database); (run.code != ExitError) || !strings.Contains(run.stderr, "https") {
        t.Fatalf("TLS with http origins: %+v", run)
    }
}

func TestImportValidatesOnlyImportedMonths(t *testing.T) {
    directory, database := newTestDirectory(t)
    template := "--template=" + filepath.Join("..", "..", "..", "share", "timesheet", "timesheet-database-template.json")
    runCommand(t, "", "init", template, database)
    runCommand(t, "correct horse battery", "user", "add", "olivier", database)
    file := filepath.Join(directory, "activite.json")
    writeTestImport(t, file)
    run := runCommand(t, "", "import", "olivier", file, "--validate", database)
    if strings.Contains(run.stdout, "2025-02") || !strings.Contains(run.stdout, "2025-01 validé") {
        t.Fatalf("validation outside the imported range: %s", run.stdout)
    }
}

func TestImportValidationStopsAtAMonthLeftOutOfTheFile(t *testing.T) {
    directory, database := newTestDirectory(t)
    template := "--template=" + filepath.Join("..", "..", "..", "share", "timesheet", "timesheet-database-template.json")
    runCommand(t, "", "init", template, database)
    runCommand(t, "correct horse battery", "user", "add", "olivier", database)
    days := []string{}
    for _, m := range []activity.Month{activity.NewMonth(2026, time.January), activity.NewMonth(2026, time.March)} {
        for d := m.First(); !d.After(m.Last()); d = d.AddDays(1) {
            days = append(days, `"` + d.String() + `": {"am": "TT", "pm": "TT"}`)
        }
    }
    file := filepath.Join(directory, "gap.json")
    writeTestFile(t, file, `{"format": "timesheet-import", "version": 1, "settings": {"start": "2026-01-01"}, "days": {` + strings.Join(days, ", ") + `}}`)
    run := runCommand(t, "", "import", "olivier", file, "--validate", database)
    if (run.code != ExitOK) || !strings.Contains(run.stdout, "2026-01 validé") || !strings.Contains(run.stdout, "validation arrêtée à 2026-02 : ce mois n'a pas été importé") || strings.Contains(run.stdout, "2026-03 validé") {
        t.Fatalf("validation across a month left out: %+v", run)
    }
}

func TestImportReportNeutralisesControlCharacters(t *testing.T) {
    directory, database := newTestDirectory(t)
    template := "--template=" + filepath.Join("..", "..", "..", "share", "timesheet", "timesheet-database-template.json")
    runCommand(t, "", "init", template, database)
    runCommand(t, "correct horse battery", "user", "add", "olivier", database)
    file := filepath.Join(directory, "escape.json")
    writeTestFile(t, file, `{"format": "timesheet-import", "version": 1, "settings": {"counters": {"\u001b[2J": {"initial": "0"}}}, "days": {}, "adjustments": [{"date": "2026-11-03", "counter": "\u001b]0;x\u0007", "amount": "1", "reason": "bulletin"}]}`)
    if run := runCommand(t, "", "import", "olivier", file, database); (run.code != ExitError) || strings.ContainsAny(run.stdout + run.stderr, "\x1b\x07") {
        t.Fatalf("control characters echoed: %q", run.stdout)
    }
    writeTestFile(t, file, `{"format": "timesheet-import", "version": 1, "settings": {"counters": {"\u001b[2J": {"initial": 5}}}, "days": {}}`)
    if run := runCommand(t, "", "import", "olivier", file, database); (run.code != ExitError) || strings.ContainsAny(run.stdout + run.stderr, "\x1b") {
        t.Fatalf("control characters echoed by a decoding error: %q", run.stderr)
    }
}

func TestConfigSetChecksThePortOfTheAddress(t *testing.T) {
    directory, database := newTestDirectory(t)
    config := "--config=" + filepath.Join(directory, "timesheet.json")
    template := "--template=" + filepath.Join("..", "..", "..", "share", "timesheet", "timesheet-database-template.json")
    runCommand(t, "", "init", template, database)
    if run := runCommand(t, "", "config", "set", "server.listen", "127.0.0.1:99999", database, config); run.code != ExitError {
        t.Fatalf("port out of range: %+v", run)
    }
    if run := runCommand(t, "", "config", "set", "server.listen", "127.0.0.1:http", database, config); run.code != ExitOK {
        t.Fatalf("service name: %+v", run)
    }
    if run := runCommand(t, "", "config", "show", database, config); (run.code != ExitOK) || !strings.Contains(run.stdout, `"listen": "127.0.0.1:http"`) {
        t.Fatalf("database with a service name: %+v", run)
    }
}

func TestImportSkipsAnAdjustmentRepeatedInTheFile(t *testing.T) {
    directory, database := newTestDirectory(t)
    template := "--template=" + filepath.Join("..", "..", "..", "share", "timesheet", "timesheet-database-template.json")
    runCommand(t, "", "init", template, database)
    runCommand(t, "correct horse battery", "user", "add", "olivier", database)
    file := filepath.Join(directory, "twice.json")
    adjustment := `{"date": "2026-11-03", "counter": "CP", "amount": "1", "reason": "bulletin"}`
    writeTestFile(t, file, `{"format": "timesheet-import", "version": 1, "days": {}, "adjustments": [` + adjustment + `, ` + adjustment + `]}`)
    if run := runCommand(t, "", "import", "olivier", file, database); (run.code != ExitOK) || !strings.Contains(run.stdout, "déjà présent, ignoré") {
        t.Fatalf("repeated adjustment: %+v", run)
    }
    if _, calendar, _ := openTestDatabase(t, database).Account("olivier"); len(calendar.Adjustments) != 1 {
        t.Fatalf("adjustments: %+v", calendar.Adjustments)
    }
}

func TestDatabaseNamesNeverCollideWithTheirCompanionFiles(t *testing.T) {
    directory, _ := newTestDirectory(t)
    if run := runCommand(t, "", "user", "list", "--database=" + filepath.Join(directory, ".ts.json.sessions")); run.code != ExitUsage {
        t.Fatalf("hidden database name: %+v", run)
    }
}

func TestImportValidatesAReopenedMonthNow(t *testing.T) {
    directory, database := newTestDirectory(t)
    template := "--template=" + filepath.Join("..", "..", "..", "share", "timesheet", "timesheet-database-template.json")
    runCommand(t, "", "init", template, database)
    runCommand(t, "correct horse battery", "user", "add", "olivier", database)
    file := filepath.Join(directory, "activite.json")
    writeTestImport(t, file)
    if run := runCommand(t, "", "import", "olivier", file, "--validate", database); run.code != ExitOK {
        t.Fatalf("import: %+v", run)
    }
    opened := openTestDatabase(t, database)
    _, calendar, _ := opened.Account("olivier")
    january := activity.NewMonth(2025, time.January)
    closing := activity.NewMonthClosing(activity.NewBookkeeper(opened.Configuration(), calendar, activity.NewHolidayCalendar()), activity.NewIdentifierGenerator(rand.Reader))
    reopenedAt := time.Date(2026, time.June, 1, 8, 0, 0, 0, time.UTC)
    reopened, err := closing.Reopen(january, "correction", reopenedAt)
    if err != nil {
        t.Fatal(err)
    }
    if err := opened.SaveCalendar("olivier", calendar.Revision, reopened); err != nil {
        t.Fatal(err)
    }
    if run := runCommand(t, "", "import", "olivier", file, "--replace", "--validate", database); (run.code != ExitOK) || !strings.Contains(run.stdout, "2025-01 validé") {
        t.Fatalf("second import: %+v", run)
    }
    _, calendar, _ = openTestDatabase(t, database).Account("olivier")
    history := calendar.Months[january].History
    if (len(history) != 3) || history[2].At.Before(reopenedAt) || !calendar.Months[january].ValidatedAt.Equal(history[2].At) {
        t.Fatalf("a month validated again is dated before its reopening: %+v", history)
    }
}

func TestUserCommandsOnAnUnknownAccountKeepTheSessions(t *testing.T) {
    directory, database := newTestDirectory(t)
    template := "--template=" + filepath.Join("..", "..", "..", "share", "timesheet", "timesheet-database-template.json")
    runCommand(t, "", "init", template, database)
    runCommand(t, "correct horse battery", "user", "add", "root", "--profile=ADMIN", database)
    sessions := filepath.Join(directory, ".ts.json.sessions")
    writeTestFile(t, sessions, "garbage")
    for _, command := range [][]string{{"user", "passwd", "ghost"}, {"user", "profile", "ghost", "USER"}, {"user", "del", "ghost"}} {
        if run := runCommand(t, "", append(command, database)...); (run.code != ExitError) || !strings.Contains(run.stderr, "unknown account") {
            t.Fatalf("%v: %+v", command, run)
        }
    }
    if data, _ := os.ReadFile(sessions); string(data) != "garbage" {
        t.Fatalf("the session file was rewritten: %s", data)
    }
}

func TestImportTemplateIsValid(t *testing.T) {
    _, database := newTestDirectory(t)
    share := filepath.Join("..", "..", "..", "share", "timesheet")
    runCommand(t, "", "init", "--template=" + filepath.Join(share, "timesheet-database-template.json"), database)
    runCommand(t, "correct horse battery", "user", "add", "demo", database)
    sample := filepath.Join(share, "timesheet-import-template.json")
    if run := runCommand(t, "", "import", "demo", sample, "--dry-run", database); (run.code != ExitOK) || !strings.Contains(run.stdout, "Simulation") {
        t.Fatalf("dry run of the import template: %+v", run)
    }
    if run := runCommand(t, "", "import", "demo", sample, "--validate", database); (run.code != ExitOK) || !strings.Contains(run.stdout, "2026-01 validé") || strings.Contains(run.stdout, "non renseigné") {
        t.Fatalf("import of the import template: %+v", run)
    }
    _, calendar, _ := openTestDatabase(t, database).Account("demo")
    if (len(calendar.Days) != 31) || (len(calendar.Adjustments) != 1) || !calendar.IsValidated(activity.NewMonth(2026, time.January)) {
        t.Fatalf("import template read as %d days, %d adjustments", len(calendar.Days), len(calendar.Adjustments))
    }
}

func TestConfigSetAcceptsSeveralTrustedProxies(t *testing.T) {
    directory, database := newTestDirectory(t)
    config := "--config=" + filepath.Join(directory, "timesheet.json")
    template := "--template=" + filepath.Join("..", "..", "..", "share", "timesheet", "timesheet-database-template.json")
    runCommand(t, "", "init", template, database)
    if run := runCommand(t, "", "config", "set", "server.trusted_proxies", "10.0.0.0/8, 172.16.0.0/12,192.168.0.0/16", database, config); run.code != ExitOK {
        t.Fatalf("list of prefixes: %+v", run)
    }
    if run := runCommand(t, "", "config", "show", database, config); (run.code != ExitOK) || !strings.Contains(strings.Join(strings.Fields(run.stdout), ""), `"trusted_proxies":["10.0.0.0/8","172.16.0.0/12","192.168.0.0/16"]`) {
        t.Fatalf("stored list: %+v", run)
    }
    for _, value := range []string{"10.0.0.0/33", "proxy.example", "10.0.0.1,fe80::1%eth0"} {
        if run := runCommand(t, "", "config", "set", "server.trusted_proxies", value, database, config); run.code != ExitError {
            t.Errorf("%s: %+v", value, run)
        }
    }
    if run := runCommand(t, "", "run", "--trusted-proxy=10.0.0.0/8", "--trusted-proxy=nope", database); run.code != ExitUsage {
        t.Fatalf("invalid option: %+v", run)
    }
    if run := runCommand(t, "", "config", "set", "server.trusted_proxy", "10.0.0.1", database, config); run.code != ExitUsage {
        t.Fatalf("former key: %+v", run)
    }
    if run := runCommand(t, "", "config", "set", "server.trusted_proxies", "", database, config); run.code != ExitOK {
        t.Fatalf("empty list: %+v", run)
    }
    if run := runCommand(t, "", "config", "show", database, config); !strings.Contains(strings.Join(strings.Fields(run.stdout), ""), `"trusted_proxies":[]`) {
        t.Fatalf("emptied list: %+v", run)
    }
}

func TestServerConfigurationLivesInItsOwnFile(t *testing.T) {
    directory, database := newTestDirectory(t)
    template := "--template=" + filepath.Join("..", "..", "..", "share", "timesheet", "timesheet-database-template.json")
    file := filepath.Join(directory, "etc", "timesheet.json")
    config := "--config=" + file
    if run := runCommand(t, "", "config", "show", config); (run.code != ExitOK) || !strings.Contains(run.stdout, `"listen": "127.0.0.1:8080"`) || !strings.Contains(run.stdout, `"timezone": "Europe/Paris"`) {
        t.Fatalf("defaults without file: %+v", run)
    }
    if _, err := os.Stat(file); err == nil {
        t.Fatal("config show created the file")
    }
    if run := runCommand(t, "", "config", "set", "server.timezone", "Asia/Tokyo", config); (run.code != ExitOK) || !strings.Contains(run.stdout, file) {
        t.Fatalf("config set: %+v", run)
    }
    data, err := os.ReadFile(file)
    if (err != nil) || !strings.Contains(string(data), `"timezone": "Asia/Tokyo"`) || !strings.Contains(string(data), `"server": {`) {
        t.Fatalf("written file: %s %v", data, err)
    }
    runCommand(t, "", "init", template, database)
    if data, _ := os.ReadFile(strings.TrimPrefix(database, "--database=")); strings.Contains(string(data), "timezone") || strings.Contains(string(data), `"server"`) {
        t.Fatal("the database still holds server settings")
    }
    if run := runCommand(t, "correct horse battery", "user", "add", "tokyo", database, config); run.code != ExitOK {
        t.Fatalf("user add: %+v", run)
    }
    if _, calendar, _ := openTestDatabase(t, database).Account("tokyo"); calendar.Settings.Start != activity.NewDate(2026, time.November, 1) {
        t.Fatalf("tracking start %s", calendar.Settings.Start)
    }
    for content, message := range map[string]string{`{"server": {"listen": "127.0.0.1:8080"}}`: "server.origins", `{"server": {"addr": "127.0.0.1:8080"}}`: "unknown field", `{"server": {}, "extra": 1}`: "extra", `nonsense`: "invalid character"} {
        writeTestFile(t, file, content)
        for _, command := range [][]string{{"config", "show"}, {"run"}, {"user", "add", "other"}} {
            if run := runCommand(t, "correct horse battery", append(command, database, config)...); (run.code != ExitError) || !strings.Contains(run.stderr, message) {
                t.Errorf("%v with %s: %+v", command, content, run)
            }
        }
    }
    if run := runCommand(t, "", "config", "show", "--config="); run.code != ExitUsage {
        t.Fatalf("empty --config: %+v", run)
    }
}

func TestConfigShowAndSetHandleEveryKey(t *testing.T) {
    directory, _ := newTestDirectory(t)
    file := filepath.Join(directory, "timesheet.json")
    config := "--config=" + file
    if run := runCommand(t, "", "config", "show", "database.path", config); (run.code != ExitOK) || (run.stdout != activity.DefaultDatabasePath + "\n") {
        t.Fatalf("raw value: %+v", run)
    }
    for _, args := range [][]string{{"config", "show", "server.addr"}, {"config", "show", "a", "b"}, {"config", "set", "server.tls_cert", "a.pem", "server.tls_key"}, {"config", "set", "server.addr", "x"}} {
        if run := runCommand(t, "", append(args, config)...); run.code != ExitUsage {
            t.Errorf("%v: %+v", args, run)
        }
    }
    for _, args := range [][]string{{"server.tls_cert", "a.pem"}, {"sessions.idle_timeout", "1m"}, {"sessions.idle_timeout", "soon"}, {"limits.accounts", "many"}} {
        if run := runCommand(t, "", append([]string{"config", "set"}, append(args, config)...)...); (run.code != ExitError) || !strings.Contains(run.stderr, args[0] + ":") {
            t.Errorf("%v: %+v", args, run)
        }
    }
    if run := runCommand(t, "", "config", "set", "server.tls_cert", "a.pem", "server.tls_key", "b.pem", "sessions.idle_timeout", "30m", config); (run.code != ExitOK) || !strings.Contains(run.stdout, "server.tls_cert, server.tls_key, sessions.idle_timeout mis à jour") {
        t.Fatalf("linked keys set together: %+v", run)
    }
    for key, want := range map[string]string{"server.tls_key": "b.pem", "sessions.idle_timeout": "30m", "server.origins": "http://127.0.0.1:8080,http://localhost:8080"} {
        if run := runCommand(t, "", "config", "show", key, config); (run.code != ExitOK) || (run.stdout != want + "\n") {
            t.Errorf("%s: %+v", key, run)
        }
    }
    data, err := os.ReadFile(file)
    if (err != nil) || !strings.Contains(string(data), "\"origins\": [\n      \"http://127.0.0.1:8080\",\n") || !strings.Contains(string(data), "\"tls_cert\": \"a.pem\"") || !strings.Contains(string(data), "\"idle_timeout\": \"30m\"") {
        t.Fatalf("the file keeps one entry per line: %s %v", data, err)
    }
    if run := runCommand(t, "", "config", "set", "server.tls_cert", "null", "server.tls_key", "null", config); run.code != ExitOK {
        t.Fatalf("null removes the pair: %+v", run)
    }
}

func TestDatabasePathComesFromTheConfiguration(t *testing.T) {
    directory, _ := newTestDirectory(t)
    config := "--config=" + filepath.Join(directory, "timesheet.json")
    template := "--template=" + filepath.Join("..", "..", "..", "share", "timesheet", "timesheet-database-template.json")
    path := filepath.Join(directory, "configured.json")
    if run := runCommand(t, "", "config", "set", "database.path", path, config); run.code != ExitOK {
        t.Fatalf("config set: %+v", run)
    }
    if run := runCommand(t, "", "init", template, config); (run.code != ExitOK) || !strings.Contains(run.stdout, path) {
        t.Fatalf("init without --database: %+v", run)
    }
    runCommand(t, "correct horse battery", "user", "add", "root", "--profile=ADMIN", config)
    if run := runCommand(t, "", "user", "list", config); (run.code != ExitOK) || !strings.Contains(run.stdout, "root") {
        t.Fatalf("user list without --database: %+v", run)
    }
    explicit := filepath.Join(directory, "explicit.json")
    if run := runCommand(t, "", "init", template, "--database=" + explicit, config); (run.code != ExitOK) || !strings.Contains(run.stdout, explicit) {
        t.Fatalf("--database comes first: %+v", run)
    }
    if run := runCommand(t, "", "config", "set", "database.path", filepath.Join(directory, ".hidden.json"), config); (run.code != ExitError) || !strings.Contains(run.stderr, "database.path:") {
        t.Fatalf("hidden database name: %+v", run)
    }
    settings, err := NewSettings([]string{"user", "list"})
    if err != nil {
        t.Fatal(err)
    }
    absolute, _ := filepath.Abs(filepath.Join("var", "base.json"))
    if err := settings.ResolveDatabase(filepath.Join("var", "base.json")); (err != nil) || (settings.Database() != absolute) {
        t.Fatalf("a relative path is made absolute: %s %v", settings.Database(), err)
    }
}

func TestLimitsComeFromTheConfiguration(t *testing.T) {
    directory, database := newTestDirectory(t)
    config := "--config=" + filepath.Join(directory, "timesheet.json")
    template := "--template=" + filepath.Join("..", "..", "..", "share", "timesheet", "timesheet-database-template.json")
    runCommand(t, "", "init", template, database, config)
    runCommand(t, "correct horse battery", "user", "add", "root", "--profile=ADMIN", database, config)
    runCommand(t, "correct horse battery", "user", "add", "second", database, config)
    if run := runCommand(t, "", "config", "set", "limits.accounts", "1", config); run.code != ExitOK {
        t.Fatalf("config set: %+v", run)
    }
    if run := runCommand(t, "", "user", "list", database, config); (run.code != ExitOK) || !strings.Contains(run.stdout, "second") {
        t.Fatalf("a lowered limit never makes the database unreadable: %+v", run)
    }
    if run := runCommand(t, "correct horse battery", "user", "add", "third", database, config); (run.code != ExitError) || !strings.Contains(run.stderr, "too many accounts") {
        t.Fatalf("a lowered limit refuses a new account: %+v", run)
    }
}

func TestRunTakesTheTLSPairFromTheFile(t *testing.T) {
    directory, database := newTestDirectory(t)
    config := "--config=" + filepath.Join(directory, "timesheet.json")
    template := "--template=" + filepath.Join("..", "..", "..", "share", "timesheet", "timesheet-database-template.json")
    runCommand(t, "", "init", template, database, config)
    runCommand(t, "correct horse battery", "user", "add", "root", "--profile=ADMIN", database, config)
    runCommand(t, "", "config", "set", "server.tls_cert", filepath.Join(directory, "a.pem"), "server.tls_key", filepath.Join(directory, "b.pem"), config)
    if run := runCommand(t, "", "run", "--listen=127.0.0.1:0", "--origin=http://127.0.0.1:1", database, config); (run.code != ExitError) || !strings.Contains(run.stderr, "TLS needs https origins") {
        t.Fatalf("a pair of the file with http origins: %+v", run)
    }
    if run := runCommand(t, "", "run", "--listen=127.0.0.1:0", "--origin=https://127.0.0.1:1", database, config); (run.code != ExitError) || !strings.Contains(run.stderr, "TLS setup failed") {
        t.Fatalf("the https origins of the options pass the check before the pair is loaded: %+v", run)
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
