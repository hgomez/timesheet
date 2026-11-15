/*
 * settings_controller.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "errors"
    "log/slog"
    "net/http"
    "strconv"
    "strings"

    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// SettingsController declaration
// ---------------------------------------------------------------------------

type SettingsController struct {
    loader     *WorkspaceLoader
    repository Repository
    renderer   *Renderer
    errors     *ErrorPages
    logger     *slog.Logger
}

type settingsRecord struct {
    Revision       int64
    Configuration  int64
    Start          string
    Locked         bool
    PublicHolidays []OptionRecord
    Counters       []counterSettingsRecord
}

type counterSettingsRecord struct {
    ID       string
    Label    string
    Enabled  bool
    Amount   string
    Default  string
    Period   string
    Initial  string
    Colors   []colorRecord
}

type colorRecord struct {
    Name    string
    Label   string
    Value   string
    Default string
}

// ---------------------------------------------------------------------------
// SettingsController implementation
// ---------------------------------------------------------------------------

func NewSettingsController(loader *WorkspaceLoader, repository Repository, renderer *Renderer, errors *ErrorPages, logger *slog.Logger) *SettingsController {
    return &SettingsController{loader: loader, repository: repository, renderer: renderer, errors: errors, logger: logger}
}

func (c *SettingsController) Show(w http.ResponseWriter, r *http.Request) {
    ws, found := c.loader.Own(r)
    if !found {
        c.errors.NotFound(w, r)
        return
    }
    c.render(w, r, ws, http.StatusOK, "")
}

func (c *SettingsController) render(w http.ResponseWriter, r *http.Request, ws *Workspace, status int, alert string, alerts ...string) {
    record := settingsRecord{Revision: ws.Calendar.Revision, Configuration: ws.Config.Revision, Start: ws.Calendar.Settings.Start.String(), Locked: ws.Calendar.HasHistory()}
    own := activity.PublicHolidayZone("")
    if ws.Calendar.Settings.PublicHolidays != nil {
        own = *ws.Calendar.Settings.PublicHolidays
    }
    record.PublicHolidays = append(record.PublicHolidays, OptionRecord{Value: "", Label: "Par défaut (" + ws.Config.PublicHolidayZone().Label() + ")", Selected: own == ""})
    for _, zone := range activity.PublicHolidayZones() {
        record.PublicHolidays = append(record.PublicHolidays, OptionRecord{Value: string(zone), Label: zone.Label(), Selected: zone == own})
    }
    for _, rule := range ws.Config.Counters {
        if rule.Archived {
            continue
        }
        settings := ws.Calendar.Settings.Counters[rule.ID]
        item := counterSettingsRecord{ID: rule.ID, Label: rule.Label, Enabled: ws.Book.Enabled(rule), Default: rule.Accrual.Amount.String(), Period: rule.Accrual.Every.Label(), Initial: settings.Initial.String()}
        if settings.Amount != nil {
            item.Amount = settings.Amount.String()
        }
        for _, code := range ws.Config.Activities {
            if !code.Archived && (code.Counter() == rule.ID) {
                item.Colors = append(item.Colors, colorRecord{Name: code.ID, Label: code.ID + " : " + code.Label, Value: string(ws.Calendar.Settings.Colors.OfCode(code)), Default: string(code.Color)})
            }
        }
        record.Counters = append(record.Counters, item)
    }
    c.renderer.Render(w, r, status, "settings", PageModel{Section: "settings", Alert: alert, Alerts: alerts, Data: record})
}

func (c *SettingsController) Submit(w http.ResponseWriter, r *http.Request) {
    ws, found := c.loader.Own(r)
    if !found {
        c.errors.NotFound(w, r)
        return
    }
    if configuration, err := strconv.ParseInt(r.PostFormValue("configuration"), 10, 64); err != nil {
        c.errors.BadRequest(w, r)
        return
    } else if configuration != ws.Config.Revision {
        c.render(w, r, ws, http.StatusConflict, "La configuration a changé entre-temps. Vérifiez puis recommencez.")
        return
    }
    calendar := ws.Calendar.Clone()
    problems := []string{}
    if !calendar.HasHistory() {
        start, err := activity.NewDateFromText(r.PostFormValue("start"))
        switch {
            case err != nil:
                problems = append(problems, "Date de début de suivi invalide.")
            case calendar.AdjustedBefore(start):
                problems = append(problems, "Des ajustements sont datés avant cette date de début : supprimez-les d'abord.")
            default:
                calendar.Settings.Start = start
        }
    }
    calendar.Settings.PublicHolidays = nil
    if zone := activity.PublicHolidayZone(r.PostFormValue("public_holidays")); zone.Valid() {
        calendar.Settings.PublicHolidays = &zone
    } else if zone != "" {
        problems = append(problems, "Calendrier des jours fériés invalide.")
    }
    counters := map[string]activity.CounterSettings{}
    for _, rule := range ws.Config.Counters {
        settings := calendar.Settings.Counters[rule.ID]
        if rule.Archived {
            if _, set := calendar.Settings.Counters[rule.ID]; set {
                counters[rule.ID] = settings
            }
            continue
        }
        enabled := r.PostFormValue("enabled:" + rule.ID) == "on"
        settings.Enabled = &enabled
        settings.Amount = nil
        if text := strings.TrimSpace(r.PostFormValue("amount:" + rule.ID)); text != "" {
            amount, err := c.quantity(text)
            if (err != nil) || (amount.Sign() < 0) || (amount.Cmp(activity.NewQuantity(MaxDays, 1)) > 0) {
                problems = append(problems, rule.Label + " : montant d'acquisition invalide.")
            } else {
                settings.Amount = &amount
            }
        }
        if !calendar.HasHistory() {
            initial, err := c.quantity(c.orZero(r.PostFormValue("initial:" + rule.ID)))
            if (err != nil) || (initial.Abs().Cmp(activity.NewQuantity(MaxDays, 1)) > 0) {
                problems = append(problems, rule.Label + " : solde initial invalide (entre -366 et 366).")
            } else {
                settings.Initial = initial
            }
        }
        counters[rule.ID] = settings
    }
    calendar.Settings.Counters = counters
    colors := activity.ColorSet{}
    if r.PostFormValue("reset-colors") != "on" {
        for _, code := range ws.Config.Activities {
            color := activity.Color(strings.ToLower(r.PostFormValue("color:" + code.ID)))
            if rule, _ := ws.Config.Counter(code.Counter()); code.Archived || rule.Archived {
                color = ws.Calendar.Settings.Colors.OfCode(code)
            }
            if !color.Valid() {
                problems = append(problems, code.Label + " : couleur invalide.")
                continue
            }
            if color != code.Color {
                colors = colors.WithCode(code.ID, color)
            }
        }
    }
    calendar.Settings.Colors = colors
    if len(problems) > 0 {
        c.render(w, r, ws, http.StatusUnprocessableEntity, "Réglages non enregistrés.", problems...)
        return
    }
    revision, err := strconv.ParseInt(r.PostFormValue("revision"), 10, 64)
    if err != nil {
        c.errors.BadRequest(w, r)
        return
    }
    err = activity.ErrStaleRevision
    if revision == ws.Calendar.Revision {
        err = c.repository.SaveCalendar(ws.Owner.Login, ws.Calendar.Revision, calendar)
    }
    if err != nil {
        if errors.Is(err, activity.ErrStaleRevision) {
            if fresh, found := c.loader.Own(r); found {
                c.render(w, r, fresh, http.StatusConflict, "Les données ont changé entre-temps. Vérifiez puis recommencez.")
                return
            }
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }
        if errors.Is(err, activity.ErrCalendarTooLarge) {
            c.render(w, r, ws, http.StatusRequestEntityTooLarge, "Les données de ce compte atteignent la taille maximale autorisée.")
            return
        }
        c.logger.Error("cannot save the settings", "error", err.Error())
        c.errors.Status(w, r, http.StatusInternalServerError)
        return
    }
    c.logger.Info("settings saved", "login", string(ws.Owner.Login))
    c.renderer.Redirect(w, r, settingsPath, "settings")
}

func (c *SettingsController) quantity(text string) (activity.Quantity, error) {
    return activity.NewQuantityFromText(strings.Replace(strings.TrimSpace(text), ",", ".", 1))
}

func (c *SettingsController) orZero(text string) string {
    if strings.TrimSpace(text) == "" {
        return "0"
    }
    return text
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
