/*
 * notices.go - Copyright (c) 2023-2026 - Olivier Poncet
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

// ---------------------------------------------------------------------------
// Notices declaration
// ---------------------------------------------------------------------------

type Notices struct{}

// ---------------------------------------------------------------------------
// Notices implementation
// ---------------------------------------------------------------------------

func NewNotices() *Notices {
    return &Notices{}
}

func (n *Notices) IsWarning(code string) bool {
    return code == "change-password"
}

func (n *Notices) Message(code string) string {
    switch code {
        case "saved":
            return "Le mois a été enregistré."
        case "validated":
            return "Le mois a été validé : ses écritures sont figées."
        case "reopened":
            return "Le mois a été rouvert : ses écritures ont été contre-passées."
        case "adjusted":
            return "L'ajustement a été enregistré."
        case "unadjusted":
            return "L'ajustement a été supprimé."
        case "settings":
            return "Les réglages ont été enregistrés."
        case "password":
            return "Le mot de passe a été changé."
        case "name":
            return "Le nom a été enregistré."
        case "created":
            return "Le compte a été créé."
        case "disabled":
            return "Le compte a été désactivé : il ne peut plus se connecter."
        case "enabled":
            return "Le compte a été réactivé."
        case "sessions":
            return "Les sessions du compte ont été fermées."
        case "deleted":
            return "Le compte a été supprimé."
        case "profile":
            return "Le profil a été changé."
        case "reset":
            return "Le mot de passe provisoire a été enregistré."
        case "configuration":
            return "La configuration a été enregistrée."
        case "change-password":
            return "Choisissez un nouveau mot de passe avant de continuer."
    }
    return ""
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
