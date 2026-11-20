#!/bin/sh
#
# ci-setup-fedora.sh - Copyright (c) 2023-2026 - Olivier Poncet
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
# no debug
# ----------------------------------------------------------------------------

set +x

# ----------------------------------------------------------------------------
# some useful variables
# ----------------------------------------------------------------------------

dnf_upgrade='yes'
dnf_install_packages='yes'
dnf_clean='yes'

dnf_packages="
rpm-build
gcc
make
golang
systemd-rpm-macros
tzdata
tar
gzip
findutils
"

# ----------------------------------------------------------------------------
# debug
# ----------------------------------------------------------------------------

set -x

# ----------------------------------------------------------------------------
# upgrade the whole system
# ----------------------------------------------------------------------------

if [ "${dnf_upgrade:-no}" = 'yes' ]
then
    dnf upgrade -y                                                   || exit 1
fi

# ----------------------------------------------------------------------------
# install the dependencies
# ----------------------------------------------------------------------------

if [ "${dnf_install_packages:-no}" = 'yes' ]
then
    dnf install -y ${dnf_packages}                                   || exit 1
fi

# ----------------------------------------------------------------------------
# clean the local repository
# ----------------------------------------------------------------------------

if [ "${dnf_clean:-no}" = 'yes' ]
then
    dnf clean all                                                    || exit 1
fi

# ----------------------------------------------------------------------------
# End-Of-File
# ----------------------------------------------------------------------------
