#!/bin/sh

set -eu

/usr/local/bin/docker-entrypoint.sh postgres &

sleep 6

/opt/bitnami/postgres-exporter/bin/postgres_exporter

