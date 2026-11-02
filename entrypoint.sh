#!/bin/sh
set -e

# PUID / PGID flexível para integração perfeita com NAS (Unraid, Synology, TrueNAS)
PUID=${PUID:-1000}
PGID=${PGID:-1000}

# Se a correr como root (UID 0), ajusta o utilizador e permissões nos diretórios de escrita
if [ "$(id -u)" = "0" ]; then
    # Cria ou ajusta grupo cwa
    if ! getent group cwa >/dev/null 2>&1; then
        addgroup -g "$PGID" cwa >/dev/null 2>&1 || true
    fi

    # Cria ou ajusta utilizador cwa
    if ! getent passwd cwa >/dev/null 2>&1; then
        adduser -u "$PUID" -G cwa -s /bin/sh -D cwa >/dev/null 2>&1 || true
    fi

    # Garante pastas essenciais
    mkdir -p /config /ingest /cwa-book-ingest /cwa-quarantine /library

    # Corrige permissões recursivamente apenas nos diretórios de escrita essenciais
    for dir in /config /ingest /cwa-book-ingest /cwa-quarantine; do
        if [ -d "$dir" ]; then
            chown -R "$PUID:$PGID" "$dir" 2>/dev/null || true
        fi
    done

    # Garante permissões na pasta library se for pasta nova
    if [ -d "/library" ] && [ ! -f "/library/metadata.db" ]; then
        chown "$PUID:$PGID" /library 2>/dev/null || true
    fi

    exec su-exec "$PUID:$PGID" /app/calibre-web-go "$@"
fi

# Se já não for root, executa diretamente
exec /app/calibre-web-go "$@"
