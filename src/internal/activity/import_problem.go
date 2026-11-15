/*
 * import_problem.go - Copyright (c) 2023-2026 - Olivier Poncet
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

package activity

// ---------------------------------------------------------------------------
// ImportProblem declaration
// ---------------------------------------------------------------------------

type ImportProblem struct {
    Date     Date
    Previous Date
    Kind     importProblemKind
    Detail   string
}

type importProblemKind int

const (
    importBeyondHorizon importProblemKind = iota
    importMonthValidated
    importEntryExists
    importGridRefusal
    importStartBeforeValidated
    importAdjustmentOutside
    importAdjustmentCounter
    importAdjustmentQuantity
    importAdjustmentReason
    importStartLocked
    importStartAfterAdjustment
    importCounterUnknown
    importCounterAmount
    importCounterInitial
    importInitialLocked
    importNothing
    importStartMoved
    importAdjustmentPresent
    importPublicHolidaysUnknown
    importAdjustmentDate
)

// ---------------------------------------------------------------------------
// ImportProblem implementation
// ---------------------------------------------------------------------------

func (p ImportProblem) Text() string {
    date := p.Date.String()
    switch p.Kind {
        case importBeyondHorizon:
            return date + ": beyond the planning horizon"
        case importMonthValidated:
            return date + ": the month is validated"
        case importEntryExists:
            return date + ": an entry already exists (use --replace)"
        case importGridRefusal:
            return date + ": " + p.Detail
        case importStartBeforeValidated:
            return "the file starts before the tracking start, already validated"
        case importAdjustmentOutside:
            return "adjustment of " + date + ": the date must fall in an open month, from the tracking start"
        case importAdjustmentCounter:
            return "adjustment of " + date + ": counter " + p.Detail + " unknown or archived"
        case importAdjustmentQuantity:
            return "adjustment of " + date + ": a non-zero quantity between -366 and 366 is expected"
        case importAdjustmentReason:
            return "adjustment of " + date + ": a reason is required, one line of 200 characters at most"
        case importStartLocked:
            return "tracking start: a month is already validated, it can no longer change"
        case importStartAfterAdjustment:
            return "tracking start: adjustments are dated before " + date
        case importCounterUnknown:
            return "counter " + p.Detail + " unknown or archived"
        case importCounterAmount:
            return p.Detail + ": invalid accrual amount (between 0 and 366)"
        case importCounterInitial:
            return p.Detail + ": invalid initial balance (between -366 and 366)"
        case importInitialLocked:
            return p.Detail + ": a month is already validated, the initial balances can no longer change"
        case importStartMoved:
            return "tracking start moved from " + p.Previous.String() + " to " + date
        case importAdjustmentPresent:
            return "adjustment of " + date + " already present, ignored"
        case importPublicHolidaysUnknown:
            return "public holidays: zone " + p.Detail + " unknown"
        case importAdjustmentDate:
            return "adjustment without a date"
    }
    return "no day to import"
}

func (p ImportProblem) Label() string {
    date := p.Date.Numeric()
    switch p.Kind {
        case importBeyondHorizon:
            return date + " : au-delà de l'horizon de planification"
        case importMonthValidated:
            return date + " : le mois est validé"
        case importEntryExists:
            return date + " : une saisie existe déjà (cochez « Remplacer les jours déjà saisis »)"
        case importGridRefusal:
            return date + " : " + p.Detail
        case importStartBeforeValidated:
            return "Le fichier commence avant la date de début, déjà validée"
        case importAdjustmentOutside:
            return "Ajustement du " + date + " : la date doit tomber dans un mois ouvert, à partir de la date de début"
        case importAdjustmentCounter:
            return "Ajustement du " + date + " : compteur " + p.Detail + " inconnu ou archivé"
        case importAdjustmentQuantity:
            return "Ajustement du " + date + " : une quantité non nulle entre -366 et 366 est attendue"
        case importAdjustmentReason:
            return "Ajustement du " + date + " : un motif est requis, une ligne de 200 caractères au plus"
        case importStartLocked:
            return "Date de début : un mois est déjà validé, elle ne peut plus changer"
        case importStartAfterAdjustment:
            return "Date de début : des ajustements sont datés avant le " + date
        case importCounterUnknown:
            return "Compteur " + p.Detail + " inconnu ou archivé"
        case importCounterAmount:
            return p.Detail + " : montant d'acquisition invalide (entre 0 et 366)"
        case importCounterInitial:
            return p.Detail + " : solde initial invalide (entre -366 et 366)"
        case importInitialLocked:
            return p.Detail + " : un mois est déjà validé, les soldes initiaux ne peuvent plus changer"
        case importStartMoved:
            return "Date de début avancée du " + p.Previous.Numeric() + " au " + date
        case importAdjustmentPresent:
            return "Ajustement du " + date + " déjà présent, ignoré"
        case importPublicHolidaysUnknown:
            return "Calendrier des jours fériés " + p.Detail + " inconnu"
        case importAdjustmentDate:
            return "Ajustement sans date"
    }
    return "Le fichier ne contient aucun jour à importer"
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
