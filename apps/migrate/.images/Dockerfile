# Database Migration Image
# Uses Atlas CLI to run migrations against PostgreSQL
#
# Environment Variables (required):
#   DB_HOST     - Database host
#   DB_PORT     - Database port (default: 5432)
#   DB_DBNAME   - Database name
#   DB_USERNAME - Database username
#   DB_PASSWORD - Database password or IAM auth token

FROM arigaio/atlas:latest AS atlas

FROM debian:bookworm-slim

# Install CA certificates for HTTPS and PostgreSQL client for health checks
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    postgresql-client \
    && rm -rf /var/lib/apt/lists/*

# Copy Atlas binary
COPY --from=atlas /atlas /usr/local/bin/atlas

# Create non-root user for security
RUN useradd -m -s /bin/bash migrate
USER migrate
WORKDIR /home/migrate

# Copy migration files
COPY --chown=migrate:migrate apps/pkgs/db/migrations ./migrations
COPY --chown=migrate:migrate apps/pkgs/db/atlas.hcl ./atlas.hcl

# Default environment variables
ENV DB_PORT=5432

# Healthcheck to verify database connectivity
HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=3 \
    CMD pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USERNAME" || exit 1

# Run migrations
# Note: The entrypoint script handles IAM auth token if needed
COPY --chown=migrate:migrate apps/migrate/.images/entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]
