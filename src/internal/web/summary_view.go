/*
 * summary_view.go - Copyright (c) 2023-2026 - Olivier Poncet
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

import "timesheet/internal/activity"

// ---------------------------------------------------------------------------
// SummaryView declaration
// ---------------------------------------------------------------------------

type SummaryView struct{}

type summaryRecord struct {
    Year     int
    Caption  string
    Report   string
    Groups   []summaryGroup
    Headers  []summaryHeader
    Opening  []summaryCell
    Rows     []summaryRowRecord
    Totals   []summaryCell
}

type summaryGroup struct {
    Label string
    Title string
    Span  int
}

type summaryHeader struct {
    Short string
    Title string
}

type summaryRowRecord struct {
    Month     string
    Path      string
    State     string
    StateName string
    Current   bool
    Cells     []summaryCell
}

type summaryCell struct {
    Value    string
    Title    string
    Negative bool
}

// ---------------------------------------------------------------------------
// SummaryView implementation
// ---------------------------------------------------------------------------

func NewSummaryView() *SummaryView {
    return &SummaryView{}
}

func (v *SummaryView) Build(ws *Workspace, journal *activity.Journal, first activity.Month, last activity.Month, caption string, report string) summaryRecord {
    summary := activity.NewSummary(ws.Book, journal, first, last, ws.Today)
    record := summaryRecord{Year: first.Year(), Caption: caption, Report: report}
    for _, rule := range summary.Counters {
        group := summaryGroup{Label: rule.ShortLabel, Title: rule.Label, Span: 3}
        if ws.Config.Consumed(rule.ID) {
            group.Span++
            record.Headers = append(record.Headers, summaryHeader{Short: "Débit", Title: rule.Label + " pris dans le mois"})
            record.Opening = append(record.Opening, summaryCell{})
        }
        record.Headers = append(record.Headers, summaryHeader{Short: "Crédit", Title: "Acquis et crédité dans le mois, " + rule.Label})
        record.Opening = append(record.Opening, summaryCell{})
        record.Headers = append(record.Headers, summaryHeader{Short: "Ajustement", Title: "Ajustements du mois, " + rule.Label}, summaryHeader{Short: "Solde", Title: "Solde en fin de mois, " + rule.Label})
        opening := summary.Opening[rule.ID]
        record.Opening = append(record.Opening, summaryCell{}, summaryCell{Value: opening.Format(), Negative: opening.Sign() < 0})
        record.Groups = append(record.Groups, group)
    }
    record.Groups = append(record.Groups, summaryGroup{Label: "Jours", Title: "Jours du mois", Span: 4})
    record.Headers = append(record.Headers, summaryHeader{Short: "Ouvrés", Title: "Total moins les jours fériés"}, summaryHeader{Short: "Fériés", Title: "Jours fériés légaux du lundi au vendredi"}, summaryHeader{Short: "Non renseignés", Title: "Demi-journées sans saisie, en jours"}, summaryHeader{Short: "Total", Title: "Jours du lundi au vendredi"})
    for len(record.Opening) < len(record.Headers) {
        record.Opening = append(record.Opening, summaryCell{})
    }
    for _, row := range summary.Rows {
        item := summaryRowRecord{Month: row.Month.Title(), Path: ws.MonthPath(row.Month), State: string(row.State), StateName: row.State.Label(), Current: row.Month == ws.Today.Month()}
        if row.State != activity.RowUntracked {
            item.Cells = v.cells(ws.Config, summary.Counters, row)
        }
        record.Rows = append(record.Rows, item)
    }
    record.Totals = v.cells(ws.Config, summary.Counters, summary.Totals)
    return record
}

func (v *SummaryView) cells(config activity.Configuration, counters []activity.CounterRule, row activity.SummaryRow) []summaryCell {
    cells := []summaryCell{}
    for _, rule := range counters {
        if config.Consumed(rule.ID) {
            cells = append(cells, summaryCell{Value: row.Debits[rule.ID].Format()})
        }
        cells = append(cells, summaryCell{Value: row.Credits[rule.ID].Format()})
        cells = append(cells, summaryCell{Value: row.Adjustments[rule.ID].FormatSigned()})
        balance := row.Balances[rule.ID]
        cells = append(cells, summaryCell{Value: balance.Format(), Negative: balance.Sign() < 0})
    }
    i := row.Indicators
    return append(cells, summaryCell{Value: i.WorkingDays.FormatDays()}, summaryCell{Value: i.Holidays.FormatDays()}, summaryCell{Value: i.Missing.FormatDays(), Negative: i.Missing.Sign() > 0}, summaryCell{Value: i.Total.FormatDays()})
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
