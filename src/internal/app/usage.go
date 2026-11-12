/*
 * usage.go - Copyright (c) 2023-2026 - Olivier Poncet
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
    "fmt"
    "io"
    "strings"
)

// ---------------------------------------------------------------------------
// Usage declaration
// ---------------------------------------------------------------------------

type Usage struct {
    program string
    groups  []usageGroup
}

type usageGroup struct {
    title   string
    options [][2]string
}

// ---------------------------------------------------------------------------
// Usage implementation
// ---------------------------------------------------------------------------

func NewUsage(program string) *Usage {
    return &Usage{
        program: program,
        groups: []usageGroup{
            {title: "Commandes", options: [][2]string{
                {"init", "créer la base à partir du modèle de configuration"},
                {"run", "lancer le serveur web"},
                {"config show", "afficher la configuration du serveur"},
                {"config set KEY VALUE", "changer un réglage du fichier de configuration (server.addr, server.origins, server.trusted_proxies, server.timezone)"},
                {"user add LOGIN", "créer un compte (mot de passe demandé)"},
                {"user profile LOGIN PROFILE", "changer le profil d'un compte (USER ou ADMIN)"},
                {"user passwd LOGIN", "changer le mot de passe d'un compte"},
                {"user del LOGIN", "supprimer un compte et ses données"},
                {"user list", "lister les comptes"},
                {"session clear", "fermer toutes les sessions (le serveur arrêté)"},
                {"import LOGIN FILE", "importer un fichier JSON (format timesheet-import)"},
                {"help", "afficher cette aide"},
                {"version", "afficher la version"},
            }},
            {title: "Options communes", options: [][2]string{
                {"--database=<file>", "base unique (défaut " + DefaultDatabase + ")"},
                {"--config=<file>", "configuration du serveur (défaut " + DefaultConfig + ")"},
                {"--log-level=<level>", "debug, info, warn ou error (défaut info)"},
                {"-h, --help", "afficher cette aide"},
            }},
            {title: "Options de init", options: [][2]string{
                {"--template=<file>", "modèle de configuration (défaut " + DefaultTemplate + ")"},
            }},
            {title: "Options de run", options: [][2]string{
                {"--addr=<address>", "adresse d'écoute (défaut celle de la configuration)"},
                {"--origin=<origin>", "origine publique, répétable (défaut celles de la configuration)"},
                {"--trusted-proxy=<address>", "proxy dont X-Forwarded-For est accepté, adresse ou préfixe, répétable (défaut ceux de la configuration)"},
                {"--tls-cert=<file>", "certificat d'une paire fournie (HTTPS)"},
                {"--tls-key=<file>", "clé privée d'une paire fournie (HTTPS)"},
            }},
            {title: "Options de user add", options: [][2]string{
                {"--profile=<profile>", "USER ou ADMIN (défaut USER)"},
            }},
            {title: "Options de import", options: [][2]string{
                {"--from=<YYYY-MM>", "premier mois importé"},
                {"--to=<YYYY-MM>", "dernier mois importé"},
                {"--dry-run", "contrôler sans rien écrire"},
                {"--replace", "remplacer les saisies existantes aux dates présentes dans le fichier"},
                {"--validate", "valider ensuite les mois importés, dans l'ordre"},
            }},
        },
    }
}

func (u *Usage) Write(w io.Writer) {
    fmt.Fprintf(w, "Usage: %s COMMANDE [ARGUMENT...] [OPTION...]\n\n", u.program)
    width := 0
    for _, group := range u.groups {
        for _, option := range group.options {
            width = max(width, len(option[0]))
        }
    }
    for _, group := range u.groups {
        fmt.Fprintf(w, "%s:\n", group.title)
        for _, option := range group.options {
            fmt.Fprintf(w, "  %s%s  %s\n", option[0], strings.Repeat(" ", width - len(option[0])), option[1])
        }
        fmt.Fprintln(w)
    }
    fmt.Fprintf(w, "Le mot de passe de user add et user passwd est lu sur le terminal sans écho, sinon sur une ligne de l'entrée standard, ou dans la variable TIMESHEET_PASSWORD.\nLes options s'écrivent --nom=valeur, avant ou après les arguments ; un argument ne peut pas commencer par un tiret.\n")
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
