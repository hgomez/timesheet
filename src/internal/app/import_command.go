/*
 * import_command.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "errors"
    "fmt"
    "io"
    "log/slog"
    "slices"
    "strconv"
    "strings"
    "time"

    "timesheet/internal/account"
    "timesheet/internal/activity"
    "timesheet/internal/store"
    "timesheet/internal/web"
)

// ---------------------------------------------------------------------------
// ImportCommand declaration
// ---------------------------------------------------------------------------

type ImportCommand struct {
    settings *Settings
    clock    activity.Clock
    random   io.Reader
    logger   *slog.Logger
}

const maxImportedDays = 366

var ErrImportRefused = errors.New("app: import refused")

// ---------------------------------------------------------------------------
// ImportCommand implementation
// ---------------------------------------------------------------------------

func NewImportCommand(settings *Settings, clock activity.Clock, random io.Reader, logger *slog.Logger) *ImportCommand {
    return &ImportCommand{settings: settings, clock: clock, random: random, logger: logger}
}

func (c *ImportCommand) Run(out io.Writer) error {
    from, to, err := c.period()
    if err != nil {
        return err
    }
    database := store.NewDatabase(c.settings.Database(), c.logger)
    err = database.Lock()
    if errors.Is(err, store.ErrLocked) {
        return fmt.Errorf("%w (stop the server first)", err)
    }
    if err != nil {
        return err
    }
    defer database.Unlock()
    if err := database.Load(); err != nil {
        return err
    }
    login, err := account.NewLogin(c.settings.Argument(0))
    if err != nil {
        return err
    }
    _, calendar, found := database.Account(login)
    if !found {
        return account.ErrUnknownAccount
    }
    config := database.Configuration()
    imported, err := NewImportFileFromFile(c.settings.Argument(1))
    if err != nil {
        return err
    }
    server, err := NewServerFile(c.settings.Config()).Load()
    if err != nil {
        return err
    }
    today := activity.NewDateOf(c.clock.Now().In(server.Location()))
    holidays := activity.NewHolidayCalendar()
    next := calendar.Clone()
    problems := c.apply(config, &next, imported.Settings)
    warnings := []string{}
    dates := []activity.Date{}
    for d, entry := range imported.Days {
        entry.Comment = strings.TrimSpace(entry.Comment)
        imported.Days[d] = entry
        if !entry.Empty() && !d.Month().Before(from) && !d.Month().After(to) {
            dates = append(dates, d)
        }
    }
    slices.SortFunc(dates, activity.Date.Compare)
    adjustments := []ImportAdjustment{}
    for _, adjustment := range imported.Adjustments {
        adjustment.Reason = strings.TrimSpace(adjustment.Reason)
        if !adjustment.Date.Month().Before(from) && !adjustment.Date.Month().After(to) {
            adjustments = append(adjustments, adjustment)
        }
    }
    if (len(dates) == 0) && (len(adjustments) == 0) && (imported.Settings == nil) {
        problems = append(problems, "aucune journée à importer")
    }
    first, last := to, from
    if len(dates) > 0 {
        first, last = dates[0].Month(), dates[len(dates) - 1].Month()
        if !next.Tracks(dates[0]) {
            if next.HasHistory() {
                problems = append(problems, "le fichier commence avant le début du suivi, déjà validé")
            } else {
                warnings = append(warnings, "début du suivi avancé du " + next.Settings.Start.String() + " au " + dates[0].String())
                next.Settings.Start = dates[0]
            }
        }
    }
    horizon := today.Month()
    for range web.PlanningMonths {
        horizon = horizon.Next()
    }
    book := activity.NewBookkeeper(config, next, holidays)
    for _, d := range dates {
        entry := imported.Days[d]
        existing := calendar.Days[d]
        switch {
            case d.Month().After(horizon):
                problems = append(problems, d.String() + " : au-delà de l'horizon de planification")
                continue
            case next.IsValidated(d.Month()):
                problems = append(problems, d.String() + " : le mois est validé")
                continue
            case !existing.Empty() && !c.settings.Replace():
                problems = append(problems, d.String() + " : une saisie existe déjà (utilisez --replace)")
                continue
        }
        for _, problem := range book.CheckEntry(d, existing, entry) {
            problems = append(problems, d.String() + " : " + problem)
        }
        next.Days[d] = entry
    }
    identifiers := activity.NewIdentifierGenerator(c.random)
    for _, adjustment := range adjustments {
        found := adjustment.Problems(config, next, horizon)
        problems = append(problems, found...)
        if len(found) > 0 {
            continue
        }
        rule, _ := config.Counter(adjustment.Counter)
        if adjustment.Exists(next, rule) {
            warnings = append(warnings, "ajustement du " + adjustment.Date.String() + " déjà présent, ignoré")
            continue
        }
        id, err := identifiers.Next()
        if err != nil {
            return err
        }
        next.Adjustments = append(next.Adjustments, adjustment.Adjustment(id, rule))
    }
    for _, warning := range warnings {
        fmt.Fprintln(out, "avertissement : " + warning)
    }
    if len(problems) > 0 {
        for _, problem := range problems {
            fmt.Fprintln(out, "erreur : " + problem)
        }
        return fmt.Errorf("%w: %d error(s)", ErrImportRefused, len(problems))
    }
    if imported.Settings != nil {
        fmt.Fprintf(out, "Réglages lus : début du suivi %s, %d compteur(s).\n", next.Settings.Start, len(imported.Settings.Counters))
    }
    if len(dates) > 0 {
        fmt.Fprintf(out, "%d journée(s) lue(s), de %s à %s.\n", len(dates), first, last)
        c.missing(out, config, next, holidays, first, last, today)
    }
    if len(adjustments) > 0 {
        fmt.Fprintf(out, "%d ajustement(s) lu(s).\n", len(adjustments))
    }
    if c.settings.Validate() && (len(dates) > 0) {
        next = c.validate(out, config, server.Location(), next, holidays, dates, today)
    }
    if c.settings.DryRun() {
        if err := database.CheckCalendar(login, calendar.Revision, next); err != nil {
            return err
        }
        fmt.Fprintln(out, "Simulation : rien n'a été écrit.")
        return nil
    }
    if err := database.SaveCalendar(login, calendar.Revision, next); err != nil {
        return err
    }
    fmt.Fprintln(out, "Import enregistré.")
    return nil
}

func (c *ImportCommand) apply(config activity.Configuration, calendar *activity.Calendar, settings *ImportSettings) []string {
    if settings == nil {
        return nil
    }
    problems := []string{}
    locked := calendar.HasHistory()
    if (settings.Start != nil) && (*settings.Start != calendar.Settings.Start) {
        switch {
            case locked:
                problems = append(problems, "début du suivi : un mois est déjà validé, il n'est plus modifiable")
            case calendar.AdjustedBefore(*settings.Start):
                problems = append(problems, "début du suivi : des ajustements sont datés avant " + settings.Start.String())
            default:
                calendar.Settings.Start = *settings.Start
        }
    }
    if calendar.Settings.Counters == nil {
        calendar.Settings.Counters = map[string]activity.CounterSettings{}
    }
    limit := activity.NewQuantity(maxImportedDays, 1)
    ids := []string{}
    for id := range settings.Counters {
        ids = append(ids, id)
    }
    slices.Sort(ids)
    for _, id := range ids {
        given := settings.Counters[id]
        rule, known := config.Counter(id)
        if !known || rule.Archived {
            problems = append(problems, "compteur " + strconv.QuoteToASCII(id) + " inconnu ou archivé")
            continue
        }
        current := calendar.Settings.Counters[id]
        if given.Enabled != nil {
            current.Enabled = given.Enabled
        }
        if (given.Amount != nil) && ((given.Amount.Sign() < 0) || (given.Amount.Cmp(limit) > 0)) {
            problems = append(problems, rule.Label + " : montant d'acquisition invalide (entre 0 et 366)")
        }
        current.Amount = given.Amount
        initial := current.Initial
        if given.Initial != nil {
            initial = *given.Initial
        }
        if initial.Abs().Cmp(limit) > 0 {
            problems = append(problems, rule.Label + " : solde initial invalide (entre -366 et 366)")
        }
        if !initial.Equal(current.Initial) {
            if locked {
                problems = append(problems, rule.Label + " : un mois est déjà validé, les soldes initiaux ne sont plus modifiables")
            } else {
                current.Initial = initial
            }
        }
        calendar.Settings.Counters[id] = current
    }
    return problems
}

func (c *ImportCommand) period() (activity.Month, activity.Month, error) {
    from, to := activity.NewMonth(activity.MinYear, time.January), activity.NewMonth(activity.MaxYear, time.December)
    fromText, toText := c.settings.Period()
    var err error
    if fromText != "" {
        if from, err = activity.NewMonthFromText(fromText); err != nil {
            return from, to, NewUsageError("invalid --from")
        }
    }
    if toText != "" {
        if to, err = activity.NewMonthFromText(toText); err != nil {
            return from, to, NewUsageError("invalid --to")
        }
    }
    if from.After(to) {
        return from, to, NewUsageError("--from after --to")
    }
    return from, to, nil
}

func (c *ImportCommand) missing(out io.Writer, config activity.Configuration, calendar activity.Calendar, holidays *activity.HolidayCalendar, first activity.Month, last activity.Month, today activity.Date) {
    book := activity.NewBookkeeper(config, calendar, holidays)
    for m := first; !m.After(last); m = m.Next() {
        dates := []string{}
        for d := m.First(); !d.After(m.Last()) && d.Before(today); d = d.AddDays(1) {
            if (book.DayType(d, activity.Morning, today) == activity.DayMissing) || (book.DayType(d, activity.Afternoon, today) == activity.DayMissing) {
                dates = append(dates, d.String())
            }
        }
        if len(dates) > 0 {
            fmt.Fprintf(out, "non renseigné(s) en %s : %s\n", m, strings.Join(dates, ", "))
        }
    }
}

func (c *ImportCommand) validate(out io.Writer, config activity.Configuration, location *time.Location, calendar activity.Calendar, holidays *activity.HolidayCalendar, dates []activity.Date, today activity.Date) activity.Calendar {
    identifiers := activity.NewIdentifierGenerator(c.random)
    imported := map[activity.Month]bool{}
    for _, d := range dates {
        imported[d.Month()] = true
    }
    last := dates[len(dates) - 1].Month()
    for m := calendar.FirstOpen(); m.Before(today.Month()) && !m.After(last); m = m.Next() {
        if !imported[m] {
            fmt.Fprintf(out, "validation arrêtée à %s : ce mois n'a pas été importé\n", m)
            break
        }
        closing := activity.NewMonthClosing(activity.NewBookkeeper(config, calendar, holidays), identifiers)
        end := m.Last().Time()
        at := time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, location).UTC()
        if history := calendar.Months[m].History; (len(history) > 0) && history[len(history) - 1].At.After(at) {
            at = c.clock.Now().UTC()
        }
        next, err := closing.Validate(m, at, today)
        if err != nil {
            fmt.Fprintf(out, "validation arrêtée à %s : %v\n", m, err)
            break
        }
        calendar = next
        fmt.Fprintf(out, "%s validé.\n", m)
    }
    return calendar
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
