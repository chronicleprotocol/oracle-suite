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

APP_USER="chronicle"
APP_GROUP="chronicle"

APP_WD="/opt/chronicle"
APP_CONFIG="${APP_WD}/config.hcl"
APP_LOG_FORMAT="text"
APP_LOG_VERBOSITY="info"

for APP_NAME in ghost spire; do
	APP_PATH="$(command -v ${APP_NAME})"

	tee <<-EOF /etc/systemd/system/${APP_NAME}.service >&2
		[Unit]
		Description=${APP_NAME}-service
		After=network.target

		[Service]
		Type=simple
		Restart=always
		RestartSec=5
		User=${APP_USER}
		Group=${APP_GROUP}
		WorkingDirectory=${APP_WD}
		ExecStart=${APP_PATH} --config ${APP_CONFIG} run --log.format ${APP_LOG_FORMAT} --log.verbosity ${APP_LOG_VERBOSITY}

		[Install]
		WantedBy=multi-user.target

	EOF
done
