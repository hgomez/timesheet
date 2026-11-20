#
# timesheet.spec - Copyright (c) 2023-2026 - Olivier Poncet
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
# version
# ----------------------------------------------------------------------------

%{!?ts_version:%{error:the macro ts_version (the version of debian/changelog) is not defined, give it with --define}}

# ----------------------------------------------------------------------------
# package
# ----------------------------------------------------------------------------

Name:           timesheet
Version:        %{ts_version}
Release:        1%{?dist}
Summary:        Half-day activity tracking with leave counters

License:        GPL-2.0-or-later
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  gcc
BuildRequires:  make
BuildRequires:  systemd-rpm-macros

%if 0%{?suse_version}
BuildRequires:  go
BuildRequires:  timezone
Requires:       timezone
Requires(post): glibc
Requires(post): shadow
%else
BuildRequires:  golang
BuildRequires:  tzdata
Requires:       tzdata
Requires(post): glibc-common
Requires(post): shadow-utils
%endif

Requires(post): coreutils
Requires(post): util-linux
%{?systemd_requires}

%define         debug_package %{nil}

%description
Timesheet is a web application tracking a professional activity
half-day by half-day, with leave counters kept like bank accounts,
monthly closing and a yearly summary.

It is designed for employees working in France: French public holidays,
paid leave, RTT and part-time counters, and an interface in French.

It is a single statically linked binary, with a server configuration
file and one JSON database, served by a systemd unit.

# ----------------------------------------------------------------------------
# preparation, build, check, installation
# ----------------------------------------------------------------------------

%prep
%setup -q

%build
make

%check
make check

%install
make install DESTDIR=%{buildroot}
install -D -m 644 debian/timesheet.service %{buildroot}%{_unitdir}/timesheet.service

# ----------------------------------------------------------------------------
# scriptlets
# ----------------------------------------------------------------------------

%pre
%if 0%{?suse_version}
%service_add_pre timesheet.service
%endif

%post
if ! getent passwd timesheet > /dev/null
then
    /usr/sbin/useradd --system --user-group --home-dir /var/lib/timesheet --no-create-home --shell /usr/sbin/nologin timesheet
fi

if [ ! -d /var/lib/timesheet ]
then
    install -d -o timesheet -g timesheet -m 0700 /var/lib/timesheet
fi

arg_database="$(/usr/bin/timesheet config show database.path --config=/etc/timesheet/timesheet.json)" || arg_database=''

if [ -z "${arg_database}" ]
then
    echo "timesheet: /etc/timesheet/timesheet.json cannot be read, no database created (see the migration notes of the specifications)" >&2
elif [ "${arg_database}" = 'null' ]
then
    echo "timesheet: /etc/timesheet/timesheet.json names no database (database.path), no database created" >&2
elif [ ! -e "${arg_database}" ]
then
    case "${arg_database}" in
        /var/lib/timesheet/*/../*|/var/lib/timesheet/../*)
            echo "timesheet: ${arg_database} is not a plain path inside /var/lib/timesheet, create it with timesheet init" >&2
            ;;
        /var/lib/timesheet/*)
            runuser -u timesheet -- /usr/bin/timesheet init --config=/etc/timesheet/timesheet.json --template=/usr/share/timesheet/database-template.json || echo "timesheet: cannot create ${arg_database}, create it with timesheet init" >&2
            ;;
        /*)
            echo "timesheet: ${arg_database} is outside /var/lib/timesheet, create it with timesheet init" >&2
            ;;
        *)
            echo "timesheet: ${arg_database} is a relative path, create it with timesheet init" >&2
            ;;
    esac
fi

%if 0%{?suse_version}
%service_add_post timesheet.service
%else
%systemd_post timesheet.service
%endif

%preun
%if 0%{?suse_version}
%service_del_preun timesheet.service
%else
%systemd_preun timesheet.service
%endif

%postun
%if 0%{?suse_version}
%service_del_postun timesheet.service
%else
%systemd_postun_with_restart timesheet.service
%endif

# ----------------------------------------------------------------------------
# files
# ----------------------------------------------------------------------------

%files
%license COPYING
%{_bindir}/timesheet
%{_unitdir}/timesheet.service
%dir /etc/timesheet
%config(noreplace) /etc/timesheet/timesheet.json
%dir /usr/share/timesheet
/usr/share/timesheet/database-template.json
/usr/share/timesheet/database-example.json
/usr/share/timesheet/import-template.json

# ----------------------------------------------------------------------------
# End-Of-File
# ----------------------------------------------------------------------------
