#!/bin/sh
#
# make-install.sh - Copyright (c) 2023-2026 - Olivier Poncet
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
# debug
# ----------------------------------------------------------------------------

set -x

# ----------------------------------------------------------------------------
# install directories
# ----------------------------------------------------------------------------

arg_srcdir="$(pwd)"
arg_dstdir="${DESTDIR:-}"
arg_bindir="${arg_dstdir}/usr/bin"
arg_confdir="${arg_dstdir}/etc/timesheet"
arg_datadir="${arg_dstdir}/usr/share/timesheet"

# ----------------------------------------------------------------------------
# install files
# ----------------------------------------------------------------------------

arg_src_binary="${arg_srcdir}/bin/timesheet"
arg_dst_binary="${arg_bindir}/timesheet"

arg_src_config="${arg_srcdir}/etc/timesheet/timesheet.json"
arg_dst_config="${arg_confdir}/timesheet.json"

arg_src_database="${arg_srcdir}/share/timesheet/database-template.json"
arg_dst_database="${arg_datadir}/database-template.json"

arg_src_import="${arg_srcdir}/share/timesheet/import-template.json"
arg_dst_import="${arg_datadir}/import-template.json"

# ----------------------------------------------------------------------------
# install the binary
# ----------------------------------------------------------------------------

install -D -m 755 "${arg_src_binary}" "${arg_dst_binary}"            || exit 1

# ----------------------------------------------------------------------------
# install the configuration
# ----------------------------------------------------------------------------

install -D -m 644 "${arg_src_config}" "${arg_dst_config}"            || exit 1

# ----------------------------------------------------------------------------
# install the database template
# ----------------------------------------------------------------------------

install -D -m 644 "${arg_src_database}" "${arg_dst_database}"        || exit 1

# ----------------------------------------------------------------------------
# install the import template
# ----------------------------------------------------------------------------

install -D -m 644 "${arg_src_import}" "${arg_dst_import}"            || exit 1

# ----------------------------------------------------------------------------
# End-Of-File
# ----------------------------------------------------------------------------
