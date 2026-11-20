#!/bin/sh
#
# ci-build-rpm.sh - Copyright (c) 2023-2026 - Olivier Poncet
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
# build directories
# ----------------------------------------------------------------------------

arg_prefix="$(pwd)"
arg_topdir="${arg_prefix}/_rpm"
arg_dist_dir="${arg_prefix}/_dist"
arg_basename='timesheet'

# ----------------------------------------------------------------------------
# check the toolchain
# ----------------------------------------------------------------------------

command -v rpmbuild > /dev/null                                      || exit 1
go version                                                           || exit 1

# ----------------------------------------------------------------------------
# package version
# ----------------------------------------------------------------------------

pkg_version="$(sed -n -e 's/^timesheet (\([^)]*\)) .*/\1/p' debian/changelog | head -n 1)" || exit 1

if [ -z "${pkg_version}" ]
then
    echo "ci-build-rpm.sh: cannot read the version from debian/changelog" >&2
    exit 1
fi

# ----------------------------------------------------------------------------
# build directories of rpmbuild
# ----------------------------------------------------------------------------

for arg_dir in BUILD BUILDROOT RPMS SOURCES SRPMS
do
    mkdir -p "${arg_topdir}/${arg_dir}"                              || exit 1
done

# ----------------------------------------------------------------------------
# source tarball
# ----------------------------------------------------------------------------

tar -czf "${arg_topdir}/SOURCES/${arg_basename}-${pkg_version}.tar.gz" \
    --exclude='./.git' \
    --exclude='./.cache' \
    --exclude='./_dist' \
    --exclude='./_rpm' \
    --exclude='./bin' \
    --exclude='./var' \
    --transform "s,^\./,${arg_basename}-${pkg_version}/," \
    -C "${arg_prefix}" .                                             || exit 1

# ----------------------------------------------------------------------------
# build package
# ----------------------------------------------------------------------------

rpmbuild -bb \
    --define "_topdir ${arg_topdir}" \
    --define "ts_version ${pkg_version}" \
    rpm/timesheet.spec                                               || exit 1

# ----------------------------------------------------------------------------
# create the dist directory
# ----------------------------------------------------------------------------

mkdir -p "${arg_dist_dir}"                                           || exit 1

# ----------------------------------------------------------------------------
# move the packages to the dist directory
# ----------------------------------------------------------------------------

mv "${arg_topdir}"/RPMS/*/*.rpm "${arg_dist_dir}/"                   || exit 1
