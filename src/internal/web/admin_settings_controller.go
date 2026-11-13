/*
 * admin_settings_controller.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "slices"
    "strconv"
    "strings"

    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// AdminSettingsController declaration
// ---------------------------------------------------------------------------

type AdminSettingsController struct {
    repository Repository
    server     activity.ServerConfiguration
    renderer   *Renderer
    errors     *ErrorPages
    notices    *Notices
    logger     *slog.Logger
}

type configurationRecord struct {
    Revision   int64
    Groups     []serverGroupRecord
    Activities []activityCodeRecord
    Counters   []counterRuleRecord
    Effects    []OptionRecord
    Counter    []OptionRecord
    Every      []OptionRecord
}

type serverGroupRecord struct {
    Title string
    Facts []serverFactRecord
}

type serverFactRecord struct {
    Label string
    Value string
}

type activityCodeRecord struct {
    ID       string
    Label    string
    Effects  []OptionRecord
    Counters []OptionRecord
    Color    string
    Archived bool
    Used     bool
}

type counterRuleRecord struct {
    ID         string
    Label      string
    ShortLabel string
    Enabled    bool
    Amount     string
    Every      []OptionRecord
    Period     string
    CarryOver  bool
    Archived   bool
    Used       bool
}

const (
    effectFeeds    = "feeds"
    effectConsumes = "consumes"
)

// ---------------------------------------------------------------------------
// AdminSettingsController implementation
// ---------------------------------------------------------------------------

func NewAdminSettingsController(repository Repository, server activity.ServerConfiguration, renderer *Renderer, errors *ErrorPages, logger *slog.Logger) *AdminSettingsController {
    return &AdminSettingsController{repository: repository, server: server, renderer: renderer, errors: errors, notices: NewNotices(), logger: logger}
}

func (c *AdminSettingsController) Show(w http.ResponseWriter, r *http.Request) {
    c.render(w, r, http.StatusOK, "")
}

func (c *AdminSettingsController) render(w http.ResponseWriter, r *http.Request, status int, alert string) {
    config := c.repository.Configuration()
    record := configurationRecord{Revision: config.Revision, Groups: c.groups()}
    record.Effects = c.effects(true)
    record.Counter = c.counters(config, "")
    record.Every = c.every("")
    for _, code := range config.Activities {
        record.Activities = append(record.Activities, activityCodeRecord{ID: code.ID, Label: code.Label, Effects: c.effects(code.Feeding()), Counters: c.counters(config, code.Counter()), Color: string(code.Color), Archived: code.Archived, Used: c.repository.ActivityUsed(code.ID)})
    }
    for _, rule := range config.Counters {
        record.Counters = append(record.Counters, counterRuleRecord{ID: rule.ID, Label: rule.Label, ShortLabel: rule.ShortLabel, Enabled: rule.EnabledByDefault, Amount: rule.Accrual.Amount.String(), Every: c.every(rule.Accrual.Every), Period: rule.PeriodStart.String(), CarryOver: rule.CarryOver, Archived: rule.Archived, Used: c.repository.CounterUsed(rule.ID)})
    }
    c.renderer.Render(w, r, status, "admin_settings", PageModel{Section: "settings-admin", Notice: c.notices.Message(r.URL.Query().Get("notice")), Alert: alert, Data: record})
}

func (c *AdminSettingsController) groups() []serverGroupRecord {
    server, database, sessions, login, limits := c.server.Server, c.server.Database, c.server.Sessions, c.server.Login, c.server.Limits
    return []serverGroupRecord{
        {Title: "Réseau", Facts: []serverFactRecord{
            {"Adresse d'écoute", server.Listen},
            {"Origines publiques", strings.Join(server.Origins, ", ")},
            {"Proxys de confiance", c.optional(strings.Join(server.TrustedProxies, ", "))},
            {"Fuseau horaire", server.Timezone},
            {"Niveau du journal", server.LogLevel},
            {"Certificat TLS", c.optional(string(server.TLSCert))},
            {"Clé TLS", c.optional(string(server.TLSKey))},
            {"Version TLS minimale", server.TLSMinVersion},
            {"Durée HSTS", server.HSTSMaxAge.Label()},
        }},
        {Title: "Requêtes", Facts: []serverFactRecord{
            {"Lecture des en-têtes", server.ReadHeaderTimeout.Label()},
            {"Lecture de la requête", server.ReadTimeout.Label()},
            {"Écriture de la réponse", server.WriteTimeout.Label()},
            {"Connexion inactive", server.KeepAliveTimeout.Label()},
            {"Arrêt du serveur", server.ShutdownTimeout.Label()},
            {"Taille des en-têtes", c.size(server.MaxHeaderBytes)},
            {"Taille du corps", c.size(server.MaxBodyBytes)},
            {"Longueur de l'adresse", strconv.FormatInt(server.MaxURLBytes, 10) + " octets"},
        }},
        {Title: "Base de données", Facts: []serverFactRecord{
            {"Fichier", database.Path},
            {"Taille maximale", c.size(database.MaxBytes)},
        }},
        {Title: "Sessions", Facts: []serverFactRecord{
            {"Expiration sans activité", sessions.IdleTimeout.Label()},
            {"Durée maximale", sessions.MaxLifetime.Label()},
            {"Sessions par compte", strconv.FormatInt(sessions.PerAccount, 10)},
            {"Enregistrement des sessions", sessions.SaveInterval.Label()},
        }},
        {Title: "Connexion", Facts: []serverFactRecord{
            {"Validité du formulaire", login.FormLifetime.Label()},
            {"Premier délai", login.InitialDelay.Label()},
            {"Délai maximal", login.MaxDelay.Label()},
            {"Fenêtre du budget", login.BudgetWindow.Label()},
            {"Nouvelles adresses par fenêtre", strconv.FormatInt(login.NewAddressesPerWindow, 10)},
            {"Adresses suivies", strconv.FormatInt(login.TrackedAddresses, 10)},
            {"Oubli d'une adresse", login.ForgetAfter.Label()},
        }},
        {Title: "Mots de passe", Facts: []serverFactRecord{
            {"Itérations PBKDF2", strconv.FormatInt(c.server.Passwords.Iterations, 10)},
        }},
        {Title: "Limites", Facts: []serverFactRecord{
            {"Comptes", strconv.FormatInt(limits.Accounts, 10)},
            {"Taille d'un compte", c.size(limits.AccountBytes)},
            {"Ajustements par compte", strconv.FormatInt(limits.Adjustments, 10)},
        }},
    }
}

func (c *AdminSettingsController) optional(value string) string {
    if value == "" {
        return "aucun"
    }
    return value
}

func (c *AdminSettingsController) size(bytes int64) string {
    switch {
        case (bytes % (1 << 20)) == 0:
            return strconv.FormatInt(bytes >> 20, 10) + " Mio"
        case (bytes % (1 << 10)) == 0:
            return strconv.FormatInt(bytes >> 10, 10) + " Kio"
    }
    return strconv.FormatInt(bytes, 10) + " octets"
}

func (c *AdminSettingsController) counters(config activity.Configuration, selected string) []OptionRecord {
    options := []OptionRecord{}
    for _, rule := range config.Counters {
        options = append(options, OptionRecord{Value: rule.ID, Label: rule.Label, Selected: rule.ID == selected})
    }
    return options
}

func (c *AdminSettingsController) effects(feeding bool) []OptionRecord {
    return []OptionRecord{{Value: effectFeeds, Label: "Alimente", Selected: feeding}, {Value: effectConsumes, Label: "Décompte", Selected: !feeding}}
}

func (c *AdminSettingsController) every(selected activity.AccrualPeriod) []OptionRecord {
    return []OptionRecord{{Value: string(activity.EveryMonth), Label: activity.EveryMonth.Label(), Selected: selected == activity.EveryMonth}, {Value: string(activity.EveryYear), Label: activity.EveryYear.Label(), Selected: selected == activity.EveryYear}}
}

func (c *AdminSettingsController) Submit(w http.ResponseWriter, r *http.Request) {
    revision, err := strconv.ParseInt(r.PostFormValue("revision"), 10, 64)
    if err != nil {
        c.errors.BadRequest(w, r)
        return
    }
    config := c.repository.Configuration().Clone()
    alert := ""
    switch r.PostFormValue("action") {
        case "activity-add", "activity-update":
            alert = c.saveActivity(&config, r)
        case "activity-delete":
            alert = c.deleteActivity(&config, r.PostFormValue("id"))
        case "counter-add", "counter-update":
            alert = c.saveCounter(&config, r)
        case "counter-delete":
            alert = c.deleteCounter(&config, r.PostFormValue("id"))
        default:
            c.errors.BadRequest(w, r)
            return
    }
    if alert != "" {
        c.render(w, r, http.StatusUnprocessableEntity, alert)
        return
    }
    err = activity.ErrStaleRevision
    if revision == config.Revision {
        err = c.repository.SaveConfiguration(config.Revision, config)
    }
    if err != nil {
        switch {
            case errors.Is(err, activity.ErrStaleRevision):
                c.render(w, r, http.StatusConflict, "La configuration a changé entre-temps. Vérifiez puis recommencez.")
            case errors.Is(err, activity.ErrInvalidConfiguration):
                c.render(w, r, http.StatusUnprocessableEntity, "Configuration refusée : " + c.refusal(err))
            case errors.Is(err, activity.ErrConfigurationInUse):
                c.render(w, r, http.StatusConflict, "Configuration refusée : elle rendrait invalides des données existantes.")
            default:
                c.logger.Error("cannot save the configuration", "error", err.Error())
                c.errors.Status(w, r, http.StatusInternalServerError)
        }
        return
    }
    c.logger.Info("configuration changed", "action", r.PostFormValue("action"), "by", string((&Viewer{}).From(r).Login))
    http.Redirect(w, r, "/admin/settings?notice=configuration", http.StatusSeeOther)
}

func (c *AdminSettingsController) refusal(err error) string {
    field, _, _ := strings.Cut(strings.TrimPrefix(err.Error(), activity.ErrInvalidConfiguration.Error() + ": "), ": ")
    switch {
        case field == "activities":
            return "entre 1 et " + strconv.Itoa(activity.MaxActivities) + " codes d'activité."
        case field == "counters":
            return "trop de compteurs (" + strconv.Itoa(activity.MaxCounters) + " au plus)."
        case strings.HasSuffix(field, ".id"):
            return "identifiant invalide ou déjà utilisé (une majuscule, puis jusqu'à 15 majuscules, chiffres ou _)."
        case strings.HasSuffix(field, ".short_label"):
            return "libellé court invalide (1 à 6 caractères)."
        case strings.HasSuffix(field, ".label"):
            return "libellé invalide (1 à 60 caractères sur une ligne)."
        case strings.HasSuffix(field, ".accrual.every"):
            return "rythme d'acquisition invalide (mois ou année)."
        case strings.HasSuffix(field, ".accrual.amount"):
            return "montant d'acquisition invalide (entre 0 et 366)."
        case strings.HasSuffix(field, ".period_start"):
            return "début de période invalide (MM-JJ)."
        case strings.HasSuffix(field, ".consumes"), strings.HasSuffix(field, ".feeds"):
            return "un code doit alimenter ou décompter un compteur existant."
        case strings.HasSuffix(field, ".color"):
            return "couleur invalide (#rrggbb)."
    }
    return "valeur invalide."
}

func (c *AdminSettingsController) saveActivity(config *activity.Configuration, r *http.Request) string {
    id := strings.TrimSpace(r.PostFormValue("id"))
    code := activity.ActivityCode{ID: id, Label: strings.TrimSpace(r.PostFormValue("label")), Color: activity.Color(strings.ToLower(r.PostFormValue("color"))), Archived: r.PostFormValue("archived") == "on"}
    switch r.PostFormValue("effect") {
        case effectFeeds:
            code.Feeds = r.PostFormValue("counter")
        case effectConsumes:
            code.Consumes = r.PostFormValue("counter")
        default:
            return "Indiquez si le code alimente ou décompte son compteur."
    }
    index := slices.IndexFunc(config.Activities, code.Matches)
    if r.PostFormValue("action") == "activity-add" {
        if index >= 0 {
            return "Cet identifiant de code existe déjà."
        }
        config.Activities = append(config.Activities, code)
        return ""
    }
    if index < 0 {
        return "Code inconnu."
    }
    previous := config.Activities[index]
    if ((previous.Consumes != code.Consumes) || (previous.Feeds != code.Feeds)) && c.repository.ActivityUsed(id) {
        return "L'effet et le compteur d'un code déjà utilisé ne sont plus modifiables."
    }
    config.Activities[index] = code
    return ""
}

func (c *AdminSettingsController) deleteActivity(config *activity.Configuration, id string) string {
    if c.repository.ActivityUsed(id) {
        return "Ce code est utilisé : archivez-le plutôt que de le supprimer."
    }
    for i, code := range config.Activities {
        if code.ID == id {
            config.Activities = slices.Delete(config.Activities, i, i + 1)
            return ""
        }
    }
    return "Code inconnu."
}

func (c *AdminSettingsController) saveCounter(config *activity.Configuration, r *http.Request) string {
    amount, err := activity.NewQuantityFromText(strings.Replace(strings.TrimSpace(r.PostFormValue("amount")), ",", ".", 1))
    if err != nil {
        return "Montant d'acquisition invalide."
    }
    period := activity.PeriodStart{}
    if period.UnmarshalText([]byte(strings.TrimSpace(r.PostFormValue("period")))) != nil {
        return "Début de période invalide (MM-JJ)."
    }
    rule := activity.CounterRule{ID: strings.TrimSpace(r.PostFormValue("id")), Label: strings.TrimSpace(r.PostFormValue("label")), ShortLabel: strings.TrimSpace(r.PostFormValue("short")), EnabledByDefault: r.PostFormValue("enabled") == "on", Accrual: activity.Accrual{Every: activity.AccrualPeriod(r.PostFormValue("every")), Amount: amount}, PeriodStart: period, CarryOver: r.PostFormValue("carry") == "on", Archived: r.PostFormValue("archived") == "on"}
    for i, existing := range config.Counters {
        if existing.ID != rule.ID {
            continue
        }
        if r.PostFormValue("action") == "counter-add" {
            return "Cet identifiant de compteur existe déjà."
        }
        config.Counters[i] = rule
        return ""
    }
    if r.PostFormValue("action") != "counter-add" {
        return "Compteur inconnu."
    }
    config.Counters = append(config.Counters, rule)
    return ""
}

func (c *AdminSettingsController) deleteCounter(config *activity.Configuration, id string) string {
    if c.repository.CounterUsed(id) {
        return "Ce compteur est utilisé : archivez-le plutôt que de le supprimer."
    }
    for _, code := range config.Activities {
        if code.Counter() == id {
            return "Un code d'activité alimente ou décompte encore ce compteur."
        }
    }
    for i, rule := range config.Counters {
        if rule.ID == id {
            config.Counters = slices.Delete(config.Counters, i, i + 1)
            return ""
        }
    }
    return "Compteur inconnu."
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
