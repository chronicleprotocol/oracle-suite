#!/usr/bin/env bash
#  Copyright (C) 2021-2023 Chronicle Labs, Inc.
#
#  This program is free software: you can redistribute it and/or modify
#  it under the terms of the GNU Affero General Public License as
#  published by the Free Software Foundation, either version 3 of the
#  License, or (at your option) any later version.
#
#  This program is distributed in the hope that it will be useful,
#  but WITHOUT ANY WARRANTY; without even the implied warranty of
#  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#  GNU Affero General Public License for more details.
#
#  You should have received a copy of the GNU Affero General Public License
#  along with this program.  If not, see <http://www.gnu.org/licenses/>.
set -xeuo pipefail

cleanInstall() {
	printf "\033[32m Reload the service unit from disk\033[0m\n"
	systemctl daemon-reload

	printf "\033[32m Unmask the service\033[0m\n"
	systemctl unmask ghost spire

	printf "\033[32m Set the preset flag for the service unit\033[0m\n"
	systemctl preset ghost spire

	printf "\033[32m Set the enabled flag for the service unit\033[0m\n"
	systemctl enable ghost spire

	printf "\033[32m Restart the service unit\033[0m\n"
	systemctl restart ghost spire
}

upgrade() {
	printf "\033[32m Reload the service unit from disk\033[0m\n"
	systemctl daemon-reload

	printf "\033[32m Restart the service unit\033[0m\n"
	systemctl restart ghost spire
}

# Step 2, check if this is a clean install or an upgrade
action="$1"
if  [ "$1" = "configure" ] && [ -z "$2" ]; then
	action="install"
elif [ "$1" = "configure" ] && [ -n "$2" ]; then
	action="upgrade"
fi

case "$action" in
	"1" | "install")
		printf "\033[32m Post Install of an clean install\033[0m\n"
		cleanInstall
		;;
	"2" | "upgrade")
		printf "\033[32m Post Install of an upgrade\033[0m\n"
		upgrade
		;;
esac
