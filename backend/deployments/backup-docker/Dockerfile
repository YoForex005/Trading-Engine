# Dockerfile for Backup Container

FROM debian:bookworm-slim

# Install required packages
RUN apt-get update && apt-get install -y \
    postgresql-client \
    redis-tools \
    gzip \
    tar \
    gnupg \
    rsync \
    awscli \
    curl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create directories
RUN mkdir -p /var/backups/trading-engine \
    /var/log/trading-engine/backup \
    /scripts

# Copy backup scripts
COPY scripts/backup/*.sh /scripts/
COPY scripts/backup/backup.config /config/backup.config

# Make scripts executable
RUN chmod +x /scripts/*.sh

# Set working directory
WORKDIR /scripts

# Default command
CMD ["/bin/bash"]

# Labels
LABEL maintainer="DevOps Team <devops@trading-engine.com>"
LABEL description="Trading Engine Backup Container"
LABEL version="1.0.0"
