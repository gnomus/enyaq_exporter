#!/bin/sh
set -eu

exec /go/enyaq_exporter/enyaq_exporter -username "${USERNAME}" -password "${PASSWORD}" -vin "${VIN}"