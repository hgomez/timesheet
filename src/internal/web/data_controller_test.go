/*
 * data_controller_test.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "encoding/json"
    "net/http"
    "net/url"
    "strings"
    "testing"
    "time"

    "timesheet/internal/account"
    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// DataController tests
// ---------------------------------------------------------------------------

func septemberImport(t *testing.T) string {
    t.Helper()
    days := map[string]activity.DayEntry{}
    for d := activity.NewDate(2026, time.September, 1); d.Month().Number() == 9; d = d.AddDays(1) {
        code := "TT"
        if d.IsWeekend() {
            code = "WE"
        }
        days[d.String()] = activity.DayEntry{AM: code, PM: code}
    }
    days["2026-09-15"] = activity.DayEntry{AM: "CP", PM: "CP", Comment: "Vacances"}
    data, err := json.Marshal(map[string]any{"format": activity.ImportFormat, "version": activity.ImportVersion, "settings": map[string]any{"start": "2026-09-01", "counters": map[string]any{"CP": map[string]any{"initial": "3"}}}, "days": days, "adjustments": []any{map[string]any{"date": "2026-10-30", "counter": "CP", "amount": "0.5", "reason": "Alignement sur le bulletin de paie"}}})
    if err != nil {
        t.Fatal(err)
    }
    return string(data)
}

func TestDataImportChecksThenSavesAndValidates(t *testing.T) {
    site := newTestSite(t)
    user := site.client().login("demo")
    page := user.page("/dashboard/data")
    if !strings.Contains(page, `<a href="/dashboard/data" aria-current="page">Mes données</a>`) || strings.Contains(user.page("/dashboard/profile"), "Exporter mes données") {
        t.Fatal("the data must have their own page, entry « Mes données » of the user menu")
    }
    if !strings.Contains(page, "<h1>Données</h1>") || !strings.Contains(page, `<form method="get" action="/dashboard/data/export">`) || !strings.Contains(page, "<h2>Exporter mes données</h2>") || !strings.Contains(page, `<button type="submit">Exporter</button>`) || !strings.Contains(page, `enctype="multipart/form-data"`) {
        t.Fatal("the profile page must offer the section Données")
    }
    if strings.Index(page, "<h2>Importer des données</h2>") > strings.Index(page, "<h2>Exporter mes données</h2>") {
        t.Fatal("the import must come before the export")
    }
    form := url.Values{"csrf": {user.token(page)}, "action": {"check"}, "validate": {"on"}}
    status, _, body := user.upload("/dashboard/data/import", form, "septembre.json", septemberImport(t))
    notice := `<p role="status">Vérification réussie : le fichier peut être importé. Rien n&#39;a été enregistré.</p>`
    if (status != http.StatusOK) || !strings.Contains(body, notice) || !strings.Contains(body, `<ul class="notice-details">`) || !strings.Contains(body, "<li>30 jour(s) lu(s), de septembre 2026 à septembre 2026.</li>") || !strings.Contains(body, "<li>Mois validés : de septembre 2026 à septembre 2026 (1).</li>") || strings.Contains(body, "Validation arrêtée") || (strings.Index(body, notice) > strings.Index(body, "<h1>Données</h1>")) {
        t.Fatalf("check: %d %s", status, body)
    }
    if _, calendar, _ := site.database.Account("demo"); len(calendar.Days) != 0 {
        t.Fatal("a check must not write")
    }
    form.Set("action", "import")
    if status, location, _ := user.upload("/dashboard/data/import", form, "septembre.json", septemberImport(t)); (status != http.StatusSeeOther) || (location != "/dashboard/data") {
        t.Fatalf("import: %d %s", status, location)
    }
    if page := user.page("/dashboard/data"); !strings.Contains(page, "L&#39;import a été enregistré.") {
        t.Fatal("the import must be confirmed")
    }
    _, calendar, _ := site.database.Account("demo")
    if (len(calendar.Days) != 30) || (len(calendar.Adjustments) != 1) || !calendar.IsValidated(activity.NewMonth(2026, time.September)) || calendar.IsValidated(activity.NewMonth(2026, time.October)) {
        t.Fatalf("imported: %d days, %d adjustments", len(calendar.Days), len(calendar.Adjustments))
    }
    if journal := site.logs.waitFor("import saved"); !strings.Contains(journal, `msg="import saved" login=demo days=30 adjustments=1 validated=1`) {
        t.Fatalf("log: %s", journal)
    }
}

func TestDataImportRefusesBadRequests(t *testing.T) {
    site := newTestSite(t)
    user := site.client().login("demo")
    token := user.token(user.page("/dashboard/data"))
    form := url.Values{"csrf": {token}, "action": {"import"}}
    if status, _, body := user.upload("/dashboard/data/import", form, "", ""); (status != http.StatusUnprocessableEntity) || !strings.Contains(body, "Choisissez un fichier à importer.") {
        t.Errorf("no file: %d", status)
    }
    if status, _, body := user.upload("/dashboard/data/import", form, "x.json", `{"format": "timesheet-import", "version": 1, "days": {}, "extra": 1}`); (status != http.StatusUnprocessableEntity) || !strings.Contains(body, "n&#39;est pas un import valide") {
        t.Errorf("invalid file: %d", status)
    }
    if status, _, body := user.upload("/dashboard/data/import", form, "x.json", `{"format": "timesheet-import", "version": 1, "days": {"2026-09-05": {"am": "CP", "pm": "CP"}, "2029-01-03": {"am": "TT", "pm": "TT"}}}`); (status != http.StatusUnprocessableEntity) || !strings.Contains(body, `<p role="alert">L&#39;import est refusé`) || !strings.Contains(body, "<li>03/01/2029 : au-delà de l&#39;horizon de planification</li>") || !strings.Contains(body, "<li>05/09/2026 : Matin : « Congés payés » est décompté") || (strings.Index(body, `<ul class="alert-details">`) > strings.Index(body, "<h1>Données</h1>")) {
        t.Errorf("refused import: %d %s", status, body)
    }
    if _, calendar, _ := site.database.Account("demo"); len(calendar.Days) != 0 {
        t.Fatal("a refused import must not write")
    }
    user.upload("/dashboard/data/import", form, "septembre.json", septemberImport(t))
    if status, _, body := user.upload("/dashboard/data/import", form, "septembre.json", septemberImport(t)); (status != http.StatusUnprocessableEntity) || !strings.Contains(body, "01/09/2026 : une saisie existe déjà") {
        t.Errorf("existing entries: %d", status)
    }
    form.Set("replace", "on")
    if status, _, _ := user.upload("/dashboard/data/import", form, "septembre.json", strings.Replace(septemberImport(t), "Vacances", "Congé d'été", 1)); status != http.StatusSeeOther {
        t.Errorf("replace: %d", status)
    }
    if _, calendar, _ := site.database.Account("demo"); calendar.Days[activity.NewDate(2026, time.September, 15)].Comment != "Congé d'été" {
        t.Error("the entries must be replaced when asked")
    }
    if status, _, _ := user.upload("/dashboard/data/import", url.Values{"action": {"check"}}, "septembre.json", septemberImport(t)); status != http.StatusForbidden {
        t.Errorf("no token: %d", status)
    }
    if status, _, _ := user.upload("/dashboard/data/import", url.Values{"csrf": {token}, "action": {"other"}}, "septembre.json", septemberImport(t)); status != http.StatusBadRequest {
        t.Errorf("unknown action: %d", status)
    }
    if status, _, _ := user.upload("/dashboard/data/import", url.Values{"csrf": {token}, "action": {"check"}}, "large.json", strings.Repeat("x", 300 << 10)); status != http.StatusBadRequest {
        t.Errorf("body over the limit: %d", status)
    }
}

func TestDataImportWarnsInYellow(t *testing.T) {
    site := newTestSite(t)
    user := site.client().login("demo")
    form := url.Values{"csrf": {user.token(user.page("/dashboard/data"))}, "action": {"check"}}
    status, _, body := user.upload("/dashboard/data/import", form, "octobre.json", `{"format": "timesheet-import", "version": 1, "days": {"2026-10-01": {"am": "TT", "pm": "TT"}}, "adjustments": [{"date": "2026-10-30", "counter": "CP", "amount": "1", "reason": "bulletin"}, {"date": "2026-10-30", "counter": "CP", "amount": "1", "reason": "bulletin"}]}`)
    if (status != http.StatusOK) || !strings.Contains(body, `<p role="status" class="is-warning">Points à vérifier :</p>`) || !strings.Contains(body, `<ul class="warning-details">`) || !strings.Contains(body, "<li>Ajustement du 30/10/2026 déjà présent, ignoré</li>") || !strings.Contains(body, "<li>Demi-journées non renseignées en octobre 2026 : 02/10/2026") {
        t.Fatalf("warnings: %d %s", status, body)
    }
}

func TestDataResetErasesTheWholeTracking(t *testing.T) {
    site := newTestSite(t)
    user := site.client().login("demo")
    page := user.page("/dashboard/data")
    if !strings.Contains(page, `<form method="post" action="/dashboard/data/reset">`) || !strings.Contains(page, `<button type="submit" class="danger">Supprimer mes données</button>`) || (strings.Index(page, "<h2>Exporter mes données</h2>") > strings.Index(page, "<h2>Supprimer mes données</h2>")) {
        t.Fatal("the profile page must end with the deletion of the data")
    }
    user.upload("/dashboard/data/import", url.Values{"csrf": {user.token(page)}, "action": {"import"}, "validate": {"on"}}, "septembre.json", septemberImport(t))
    for _, confirm := range []string{"", "Demo", "root"} {
        if status, _, body := user.post("/dashboard/data/reset", url.Values{"csrf": {user.token(page)}, "confirm": {confirm}}); (status != http.StatusUnprocessableEntity) || !strings.Contains(body, "Retapez exactement votre identifiant") {
            t.Errorf("confirmation %q: %d", confirm, status)
        }
    }
    if _, calendar, _ := site.database.Account("demo"); (len(calendar.Days) != 30) || !calendar.IsValidated(activity.NewMonth(2026, time.September)) {
        t.Fatal("a refused deletion must keep the data")
    }
    if status, location, _ := user.post("/dashboard/data/reset", url.Values{"csrf": {user.token(page)}, "confirm": {"demo"}}); (status != http.StatusSeeOther) || (location != "/dashboard/data") {
        t.Fatalf("deletion: %d %s", status, location)
    }
    _, calendar, _ := site.database.Account("demo")
    if (len(calendar.Days) != 0) || (len(calendar.Adjustments) != 0) || (len(calendar.Months) != 0) || (len(calendar.Ledger) != 0) || (len(calendar.Settings.Counters) != 0) || (calendar.Settings.Start != activity.NewDate(2026, time.November, 1)) {
        t.Fatalf("the tracking must start again: %+v", calendar.Settings)
    }
    if page := user.page("/dashboard/data"); !strings.Contains(page, "Vos données ont été supprimées") {
        t.Fatal("the deletion must be confirmed, the session kept")
    }
    if journal := site.logs.waitFor("data erased"); !strings.Contains(journal, `msg="data erased" login=demo`) {
        t.Fatalf("log: %s", journal)
    }
    if status, _, _ := user.post("/dashboard/data/reset", url.Values{"confirm": {"demo"}}); status != http.StatusForbidden {
        t.Errorf("no token: %d", status)
    }
}

func TestDataExportGivesBackTheSameBalances(t *testing.T) {
    site := newTestSite(t)
    user := site.client().login("demo")
    form := url.Values{"csrf": {user.token(user.page("/dashboard/data"))}, "action": {"import"}, "validate": {"on"}}
    user.upload("/dashboard/data/import", form, "septembre.json", septemberImport(t))
    status, header, body := user.download("/dashboard/data/export")
    if (status != http.StatusOK) || (header.Get("Content-Type") != "application/json; charset=utf-8") || (header.Get("Content-Disposition") != `attachment; filename="timesheet-demo-2026-11-08.json"`) || (header.Get("Cache-Control") != "no-store") {
        t.Fatalf("export: %d %v", status, header)
    }
    exported, err := activity.NewImportFile([]byte(body))
    if (err != nil) || (len(exported.Days) != 30) || (len(exported.Adjustments) != 1) || (exported.Settings == nil) || (exported.Settings.Start.String() != "2026-09-01") {
        t.Fatalf("exported file: %v %+v", err, exported.Settings)
    }
    admin := site.client().login("root")
    form = url.Values{"csrf": {admin.token(admin.page("/dashboard/data"))}, "action": {"import"}, "validate": {"on"}}
    if status, _, page := admin.upload("/dashboard/data/import", form, "export.json", body); status != http.StatusSeeOther {
        t.Fatalf("import of the export: %d %s", status, page)
    }
    balances := func(login string) map[string]string {
        _, calendar, _ := site.database.Account(account.Login(login))
        journal := activity.NewBookkeeper(site.database.Configuration(), calendar, activity.NewHolidayCalendar()).Journal(activity.NewMonth(2026, time.December))
        sums := map[string]string{}
        for _, id := range []string{"CP", "TT", "WE"} {
            sums[id] = journal.Balance(id, activity.NewDate(2026, time.December, 31)).String()
        }
        return sums
    }
    if demo, root := balances("demo"), balances("root"); (demo["CP"] != root["CP"]) || (demo["TT"] != root["TT"]) || (demo["WE"] != root["WE"]) {
        t.Fatalf("balances differ: %v %v", demo, root)
    }
    if _, _, page := site.client().get("/dashboard/data/export"); strings.Contains(page, "timesheet-import") {
        t.Fatal("a guest must not download any data")
    }
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
