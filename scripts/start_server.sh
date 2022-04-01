#!/usr/bin/env sh

tmpfile=$(mktemp /tmp/docker-compose.XXXXXX.yml)
curl -s "https://raw.githubusercontent.com/SphericalPotatoInVacuum/soa-message-queues/main/deploy/docker-compose.yml" > "$tmpfile"

docker-compose -f "$tmpfile" pull
docker-compose -f "$tmpfile" up || true
docker-compose -f "$tmpfile" down

rm "$tmpfile"
