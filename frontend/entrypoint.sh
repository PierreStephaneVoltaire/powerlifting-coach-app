#!/bin/sh

cat > /app/build/config.json <<EOF
{
  "apiUrl": "${REACT_APP_API_URL:-}",
  "authUrl": "${REACT_APP_AUTH_URL:-}"
}
EOF

exec "$@"
