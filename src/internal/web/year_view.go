/*
 * year_view.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "strconv"
    "time"

    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// YearView declaration
// ---------------------------------------------------------------------------

type YearView struct {
    summaries *SummaryView
    counters  *CountersView
}

type yearRecord struct {
    Year     int
    Owner    string
    Previous string
    Next     string
    Current  string
    Days     []int
    Months   []yearMonthRecord
    Totals   []yearTotalRecord
    Total    string
    Legend   []legendRecord
    Summary  summaryRecord
    Counters []counterBookRecord
}

type yearMonthRecord struct {
    Label  string
    Path   string
    Halves []yearHalfRecord
}

type yearHalfRecord struct {
    Label string
    Cells []yearCellRecord
}

type yearCellRecord struct {
    Class   string
    Letter  string
    Title   string
    Comment bool
}

type yearTotalRecord struct {
    Label string
    Value string
}

// ---------------------------------------------------------------------------
// YearView implementation
// ---------------------------------------------------------------------------

func NewYearView() *YearView {
    return &YearView{summaries: NewSummaryView(), counters: NewCountersView()}
}

func (v *YearView) Build(ws *Workspace, year int) yearRecord {
    record := yearRecord{Year: year, Current: ws.YearPath(ws.Today.Year())}
    if year > activity.MinYear {
        record.Previous = ws.YearPath(year - 1)
    }
    if year < activity.MaxYear {
        record.Next = ws.YearPath(year + 1)
    }
    if ws.ReadOnly {
        record.Owner = string(ws.Owner.Login)
    }
    record.Days = v.Days()
    months := []activity.Month{}
    for number := time.January; number <= time.December; number++ {
        m := activity.NewMonth(year, number)
        record.Months = append(record.Months, v.Month(ws, m))
        months = append(months, m)
    }
    record.Totals, record.Total = v.Totals(ws, months)
    record.Legend = NewLegendView().Build(ws.Config)
    book := ws.Book
    journal := book.Journal(book.Horizon(ws.Today, activity.NewMonth(year, time.December)))
    record.Summary = v.summaries.Build(ws, journal, activity.NewMonth(year, time.January), activity.NewMonth(year, time.December), "Synthèse de l'année " + strconv.Itoa(year), "Report N-1")
    record.Counters = v.counters.Build(ws, journal, year)
    return record
}

func (v *YearView) Totals(ws *Workspace, months []activity.Month) ([]yearTotalRecord, string) {
    totals := map[string]activity.Quantity{}
    missing := activity.NewQuantity(0, 1)
    for _, m := range months {
        if m.Before(ws.Calendar.Settings.Start.Month()) {
            continue
        }
        indicators := ws.Book.Indicators(m, ws.Today)
        for code, value := range indicators.ByCode {
            totals[code] = totals[code].Add(value)
        }
        missing = missing.Add(indicators.Missing)
    }
    records := []yearTotalRecord{}
    for _, code := range ws.Config.Activities {
        if value, counted := totals[code.ID]; counted || !code.Archived {
            records = append(records, yearTotalRecord{Label: code.Label, Value: value.FormatDays()})
        }
    }
    records = append(records, yearTotalRecord{Label: "Non renseignés", Value: missing.FormatDays()})
    total := missing
    for _, value := range totals {
        total = total.Add(value)
    }
    return records, total.FormatDays()
}

func (v *YearView) Days() []int {
    days := []int{}
    for day := 1; day <= 31; day++ {
        days = append(days, day)
    }
    return days
}

func (v *YearView) Month(ws *Workspace, m activity.Month) yearMonthRecord {
    item := yearMonthRecord{Label: m.Title(), Path: ws.MonthPath(m)}
    for _, half := range (activity.DayEntry{}).Halves() {
        row := yearHalfRecord{Label: half.Label()}
        for day := 1; day <= 31; day++ {
            if day > m.Days() {
                row.Cells = append(row.Cells, yearCellRecord{Class: "day-outside"})
                continue
            }
            row.Cells = append(row.Cells, v.cell(ws, m.First().AddDays(day - 1), half))
        }
        item.Halves = append(item.Halves, row)
    }
    return item
}

func (v *YearView) cell(ws *Workspace, d activity.Date, half activity.HalfDay) yearCellRecord {
    entry := ws.Calendar.Days[d]
    dayType := ws.Book.DayType(d, half, ws.Today)
    title := d.Label() + ", " + half.Label() + " : " + dayType.Label()
    if code, known := ws.Config.Activity(entry.Code(half)); known && (dayType != activity.DayUntracked) {
        title = d.Label() + ", " + half.Label() + " : " + code.Label
    } else if name, holiday := ws.Book.PublicHolidays().Name(d); holiday {
        title += " (" + name + ")"
    }
    if entry.Comment != "" {
        title += ", " + entry.Comment
    }
    letter, class := v.letter(dayType), "day-" + dayType.Slug()
    if code, known := ws.Config.Activity(entry.Code(half)); known && (dayType != activity.DayUntracked) {
        letter, class = code.ID, "code-" + code.ID
    }
    return yearCellRecord{Class: class, Letter: letter, Title: title, Comment: (entry.Comment != "") && (half == activity.Morning)}
}

func (v *YearView) letter(t activity.DayType) string {
    if t == activity.DayMissing {
        return "?"
    }
    return ""
}

func (v *YearView) Title(year int) string {
    return "Année " + strconv.Itoa(year)
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
