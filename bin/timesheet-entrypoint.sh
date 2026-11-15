#!/bin/sh
#
# timesheet-entrypoint.sh - Copyright (c) 2023-2026 - Olivier Poncet
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 2 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.
#

# ----------------------------------------------------------------------------
# global environment
# ----------------------------------------------------------------------------

CONFIG='/etc/timesheet/timesheet.json'
TEMPLATE='/usr/share/timesheet/database-template.json'
ADMIN_LOGIN="${TIMESHEET_USERNAME:-admin}"
ADMIN_PASSWORD="${TIMESHEET_PASSWORD:-Tms@Admin%26}"

# ----------------------------------------------------------------------------
# create the database and its administrator at the first start
# ----------------------------------------------------------------------------

if [ "${1:-}" = 'run' ]
then
    DATABASE="$(timesheet config show database.path --config="${CONFIG}")" || exit 1
    if [ "${DATABASE}" != 'null' ] && [ ! -e "${DATABASE}" ]
    then
        timesheet init --config="${CONFIG}" --template="${TEMPLATE}" || exit 1
        TIMESHEET_PASSWORD="${ADMIN_PASSWORD}" timesheet user add "${ADMIN_LOGIN}" --profile=ADMIN --provisional --config="${CONFIG}" || exit 1
    fi
fi

# ----------------------------------------------------------------------------
# run timesheet
# ----------------------------------------------------------------------------

exec timesheet "$@" --config="${CONFIG}"

# ----------------------------------------------------------------------------
# End-Of-File
# ----------------------------------------------------------------------------
