/*
 * month_view.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "strconv"
    "strings"

    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// MonthView declaration
// ---------------------------------------------------------------------------

type MonthView struct {
    statements *StatementView
    years      *YearView
    summaries  *SummaryView
}

type monthPageRecord struct {
    Owner       string
    ReadOnly    bool
    Month       monthRecord
    Days        []int
    Calendar    yearMonthRecord
    Totals      []yearTotalRecord
    Total       string
    Summary     summaryRecord
    Legend      []legendRecord
    Codes       map[string]string
    Adjustments adjustmentsRecord
    Statements  []statementRecord
}

type adjustmentsRecord struct {
    Editable bool
    Action   string
    Min      string
    Max      string
    Counters []OptionRecord
    Items    []adjustmentRecord
}

type adjustmentRecord struct {
    ID      string
    Date    string
    Label   string
    Counter string
    Amount  string
    Reason  string
}

type monthRecord struct {
    Label          string
    Status         string
    Validated      bool
    BeforeStart    bool
    Editable       bool
    Previous       string
    Next           string
    Current        string
    FormAction     string
    ValidateAction string
    ReopenAction   string
    Revision       int64
    Rows           []dayRecord
    CanValidate    bool
    ValidateReason string
    CanReopen      bool
    Anomalies      []string
    Missing        string
    History        []string
}

type dayRecord struct {
    Editable bool
    Date     string
    Label    string
    Kind     string
    States   string
    Halves   []halfRecord
    Day      halfRecord
    Comment  commentRecord
    Problems []string
}

type halfRecord struct {
    Name      string
    AriaLabel string
    Class     string
    Blank     string
    Type      string
    TypeLabel string
    Value     string
    Text      string
    Options   []OptionRecord
}

type commentRecord struct {
    Name      string
    AriaLabel string
    Value     string
}

// ---------------------------------------------------------------------------
// MonthView implementation
// ---------------------------------------------------------------------------

func NewMonthView() *MonthView {
    return &MonthView{statements: NewStatementView(), years: NewYearView(), summaries: NewSummaryView()}
}

func (v *MonthView) Build(ws *Workspace, page string, m activity.Month, submitted map[activity.Date]activity.DayEntry, problems map[activity.Date][]string) monthPageRecord {
    book := ws.Book
    journal := book.Journal(book.Horizon(ws.Today, m))
    record := monthPageRecord{ReadOnly: ws.ReadOnly}
    if ws.ReadOnly {
        record.Owner = string(ws.Owner.Login)
    }
    record.Month = v.month(ws, page, m, submitted, problems)
    record.Days = v.years.Days()
    record.Calendar = v.years.Month(ws, m)
    record.Totals, record.Total = v.years.Totals(ws, []activity.Month{m})
    record.Summary = v.summaries.Build(ws, journal, m, m, "Synthèse de " + record.Month.Label, "Report du mois précédent")
    record.Codes = map[string]string{}
    for _, code := range ws.Config.Activities {
        record.Codes[code.ID] = "code-" + code.ID
    }
    record.Adjustments = v.adjustments(ws, m, record.Month.Editable)
    record.Legend = NewLegendView().Build(ws.Config)
    for _, rule := range ws.Config.Counters {
        statement := v.statements.Build(journal, rule, m)
        if book.Enabled(rule) || (len(statement.Lines) > 0) || (len(statement.Cancelled) > 0) {
            record.Statements = append(record.Statements, statement)
        }
    }
    return record
}

func (v *MonthView) adjustments(ws *Workspace, m activity.Month, editable bool) adjustmentsRecord {
    record := adjustmentsRecord{Editable: editable, Action: "/dashboard/adjustments", Min: m.First().String(), Max: m.Last().String()}
    if start := ws.Calendar.Settings.Start; start.After(m.First()) {
        record.Min = start.String()
    }
    for _, rule := range ws.Config.Counters {
        if !rule.Archived {
            record.Counters = append(record.Counters, OptionRecord{Value: rule.ID, Label: rule.Label})
        }
    }
    for _, adjustment := range ws.Calendar.Adjustments {
        if adjustment.Date.Month() != m {
            continue
        }
        item := adjustmentRecord{ID: adjustment.ID, Date: adjustment.Date.String(), Label: adjustment.Date.Numeric(), Counter: adjustment.Counter, Amount: adjustment.Amount.Format(), Reason: adjustment.Reason}
        if rule, known := ws.Config.Counter(adjustment.Counter); known {
            item.Counter = rule.Label
        }
        record.Items = append(record.Items, item)
    }
    return record
}

func (v *MonthView) month(ws *Workspace, page string, m activity.Month, submitted map[activity.Date]activity.DayEntry, problems map[activity.Date][]string) monthRecord {
    calendar := ws.Calendar
    record := monthRecord{Label: m.Title() + " " + strconv.Itoa(m.Year()), Validated: calendar.IsValidated(m), BeforeStart: m.Before(calendar.Settings.Start.Month()), Current: ws.PagePath(page, ws.Today.Month()), FormAction: ws.FormPath(m), ValidateAction: ws.FormPath(m) + "/validate", ReopenAction: ws.FormPath(m) + "/reopen", Revision: calendar.Revision}
    if m.Previous().Year() >= activity.MinYear {
        record.Previous = ws.PagePath(page, m.Previous())
    }
    if m.Next().Year() <= activity.MaxYear {
        record.Next = ws.PagePath(page, m.Next())
    }
    record.Editable = !ws.ReadOnly && !record.Validated && !record.BeforeStart && !m.After(ws.Horizon())
    switch {
        case record.BeforeStart:
            record.Status = "avant le début du suivi"
        case m.After(ws.Horizon()):
            record.Status = "au-delà de l'horizon de planification"
        case record.Validated:
            record.Status = "validé le " + activity.NewDateOf(calendar.Months[m].ValidatedAt.In(ws.Location)).Numeric()
        default:
            record.Status = "ouvert"
    }
    if !ws.ReadOnly && !record.BeforeStart {
        closing := activity.NewMonthClosing(ws.Book, nil)
        if record.Validated {
            record.CanReopen = closing.CheckReopen(m) == nil
        } else if err := closing.Check(m, ws.Today); err != nil {
            record.ValidateReason = v.reason(err)
        } else {
            record.CanValidate = true
        }
    }
    for _, anomaly := range ws.Book.Anomalies(m) {
        record.Anomalies = append(record.Anomalies, anomaly.Date.Label() + ", " + anomaly.Message)
    }
    indicators := ws.Book.Indicators(m, ws.Today)
    record.Missing = indicators.Missing.FormatDays()
    for _, event := range calendar.Months[m].History {
        day := activity.NewDateOf(event.At.In(ws.Location)).Numeric()
        text := "Validé le " + day
        if event.Action == activity.ActionReopened {
            text = "Rouvert le " + day + " : " + event.Reason
        }
        record.History = append(record.History, text)
    }
    for d := m.First(); !d.After(m.Last()); d = d.AddDays(1) {
        entry, edited := submitted[d]
        if !edited {
            entry = calendar.Days[d]
        }
        row := v.day(ws, d, calendar.Days[d], entry, problems[d])
        row.Editable = record.Editable && calendar.Tracks(d)
        record.Rows = append(record.Rows, row)
    }
    return record
}

func (v *MonthView) day(ws *Workspace, d activity.Date, stored activity.DayEntry, entry activity.DayEntry, problems []string) dayRecord {
    holidays := ws.Book.PublicHolidays()
    record := dayRecord{Date: d.String(), Label: d.Label(), Problems: problems}
    states := []string{}
    record.Kind = "Semaine"
    if name, holiday := holidays.Name(d); holiday {
        record.Kind = name
        states = append(states, "is-holiday")
    } else if d.IsWeekend() {
        record.Kind = "Week-end"
        states = append(states, "is-weekend")
    }
    if d == ws.Today {
        states = append(states, "is-today")
    }
    if len(problems) > 0 {
        states = append(states, "is-invalid")
    }
    missing := false
    for _, half := range entry.Halves() {
        dayType := ws.Book.DayType(d, half, ws.Today)
        missing = missing || (dayType == activity.DayMissing)
        record.Halves = append(record.Halves, v.half(ws, half.Label() + " du " + d.Short(), string(half) + ":" + d.String(), entry.Code(half), stored.Code(half), dayType, v.blank(ws, d)))
    }
    if missing {
        states = append(states, "is-missing")
    }
    record.States = strings.Join(states, " ")
    record.Day = v.half(ws, "Journée du " + d.Short(), "day:" + d.String(), "", "", activity.DayEmpty, v.blank(ws, d))
    record.Comment = commentRecord{Name: "comment:" + d.String(), AriaLabel: "Commentaire du " + d.Short(), Value: entry.Comment}
    return record
}

func (v *MonthView) blank(ws *Workspace, d activity.Date) string {
    switch {
        case !ws.Calendar.Tracks(d):
            return "day-" + activity.DayUntracked.Slug()
        case d.Before(ws.Today):
            return "day-" + activity.DayMissing.Slug()
    }
    return "day-" + activity.DayEmpty.Slug()
}

func (v *MonthView) half(ws *Workspace, aria string, name string, value string, stored string, dayType activity.DayType, blank string) halfRecord {
    record := halfRecord{Name: name, AriaLabel: aria, Type: string(dayType), TypeLabel: dayType.Label(), Value: value, Blank: blank, Class: "day-" + dayType.Slug()}
    if code, known := ws.Config.Activity(value); known && (dayType != activity.DayUntracked) {
        record.Text = code.Label
        record.Type = string(activity.DayCode)
        record.TypeLabel = code.Label
        record.Class = "code-" + code.ID
    }
    record.Options = append(record.Options, OptionRecord{Value: "", Label: "Aucun", Selected: value == ""})
    for _, code := range ws.Book.Choices(stored) {
        record.Options = append(record.Options, OptionRecord{Value: code.ID, Label: code.Label, Selected: code.ID == value})
    }
    return record
}

func (v *MonthView) reason(err error) string {
    switch {
        case errors.Is(err, activity.ErrMonthNotOver):
            return "Le mois n'est pas terminé."
        case errors.Is(err, activity.ErrPreviousMonthOpen):
            return "Le mois précédent doit être validé d'abord."
        case errors.Is(err, activity.ErrMonthMissingDays):
            return "Des demi-journées ne sont pas renseignées."
        case errors.Is(err, activity.ErrMonthAnomalies):
            return "Des saisies sont invalides."
        case errors.Is(err, activity.ErrMonthBeforeStart):
            return "Le mois précède le début du suivi."
    }
    return "Le mois ne peut pas être validé."
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
