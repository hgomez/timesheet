/*
 * public_holiday_zone.go - Copyright (c) 2023-2026 - Olivier Poncet
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
// PublicHolidayZone declaration
// ---------------------------------------------------------------------------

type PublicHolidayZone string

const (
    ZoneMetropole             PublicHolidayZone = "fr-metropole"
    ZoneAlsaceMoselle         PublicHolidayZone = "fr-alsace-moselle"
    ZoneGuadeloupe            PublicHolidayZone = "fr-guadeloupe"
    ZoneGuyane                PublicHolidayZone = "fr-guyane"
    ZoneMartinique            PublicHolidayZone = "fr-martinique"
    ZoneMayotte               PublicHolidayZone = "fr-mayotte"
    ZoneNouvelleCaledonie     PublicHolidayZone = "fr-nouvelle-caledonie"
    ZoneLaReunion             PublicHolidayZone = "fr-la-reunion"
    ZonePolynesieFrancaise    PublicHolidayZone = "fr-polynesie-francaise"
    ZoneSaintBarthelemy       PublicHolidayZone = "fr-saint-barthelemy"
    ZoneSaintMartin           PublicHolidayZone = "fr-saint-martin"
    ZoneWallisEtFutuna        PublicHolidayZone = "fr-wallis-et-futuna"
    ZoneSaintPierreEtMiquelon PublicHolidayZone = "fr-saint-pierre-et-miquelon"
)

// ---------------------------------------------------------------------------
// PublicHolidayZone implementation
// ---------------------------------------------------------------------------

func PublicHolidayZones() []PublicHolidayZone {
    return []PublicHolidayZone{ZoneMetropole, ZoneAlsaceMoselle, ZoneGuadeloupe, ZoneGuyane, ZoneMartinique, ZoneMayotte, ZoneNouvelleCaledonie, ZoneLaReunion, ZonePolynesieFrancaise, ZoneSaintBarthelemy, ZoneSaintMartin, ZoneWallisEtFutuna, ZoneSaintPierreEtMiquelon}
}

func (z PublicHolidayZone) Valid() bool {
    return z.Label() != ""
}

func (z PublicHolidayZone) Label() string {
    switch z {
        case ZoneMetropole:
            return "Métropole"
        case ZoneAlsaceMoselle:
            return "Alsace-Moselle"
        case ZoneGuadeloupe:
            return "Guadeloupe"
        case ZoneGuyane:
            return "Guyane"
        case ZoneMartinique:
            return "Martinique"
        case ZoneMayotte:
            return "Mayotte"
        case ZoneNouvelleCaledonie:
            return "Nouvelle-Calédonie"
        case ZoneLaReunion:
            return "La Réunion"
        case ZonePolynesieFrancaise:
            return "Polynésie française"
        case ZoneSaintBarthelemy:
            return "Saint-Barthélemy"
        case ZoneSaintMartin:
            return "Saint-Martin"
        case ZoneWallisEtFutuna:
            return "Wallis-et-Futuna"
        case ZoneSaintPierreEtMiquelon:
            return "Saint-Pierre-et-Miquelon"
    }
    return ""
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
