# Trading Engine Backup System Makefile

.PHONY: help setup backup-full backup-incr verify restore clean health dr-test

# Configuration
BACKUP_DIR ?= /var/backups/trading-engine
LOG_DIR ?= /var/log/trading-engine/backup
SCRIPT_DIR := $(shell pwd)

help:
	@echo "Trading Engine Backup System"
	@echo "=============================="
	@echo ""
	@echo "Available targets:"
	@echo "  setup         - Initial setup (create dirs, install cron)"
	@echo "  backup-full   - Create full backup"
	@echo "  backup-incr   - Create incremental backup"
	@echo "  verify        - Verify latest backup"
	@echo "  health        - Check backup health"
	@echo "  restore       - Restore from backup (interactive)"
	@echo "  dr-test       - Run disaster recovery test"
	@echo "  clean         - Clean old backups"
	@echo "  status        - Show backup status"
	@echo ""

setup:
	@echo "Setting up backup system..."
	@mkdir -p $(BACKUP_DIR)
	@mkdir -p $(LOG_DIR)
	@chmod +x *.sh
	@echo "✓ Directories created"
	@echo ""
	@echo "Next steps:"
	@echo "1. Edit backup.config with your settings"
	@echo "2. Generate GPG key: gpg --full-generate-key"
	@echo "3. Run: make backup-full"
	@echo "4. Install cron: sudo ./setup-cron.sh"

backup-full:
	@echo "Creating full backup..."
	@./backup-full.sh

backup-incr:
	@echo "Creating incremental backup..."
	@./backup-incremental.sh

verify:
	@echo "Verifying latest backup..."
	@./run-verification.sh

health:
	@echo "Checking backup health..."
	@./backup-health-check.sh

restore:
	@echo "⚠️  WARNING: This will restore from backup"
	@echo "Available backups:"
	@find $(BACKUP_DIR) -name "backup-*.tar.gz*" | sort -r | head -5
	@echo ""
	@read -p "Enter backup file path: " backup_file; \
	read -p "Type 'force' to confirm: " confirm; \
	if [ "$$confirm" = "force" ]; then \
		./restore-full.sh "$$backup_file" force; \
	else \
		echo "Restore cancelled."; \
	fi

dr-test:
	@echo "Running disaster recovery test..."
	@./dr-test.sh

clean:
	@echo "Cleaning old backups..."
	@./cleanup-old-backups.sh

status:
	@echo "Backup System Status"
	@echo "===================="
	@echo ""
	@echo "Last full backup:"
	@find $(BACKUP_DIR) -name "full-*" -type d | sort -r | head -1 || echo "  None"
	@echo ""
	@echo "Last incremental backup:"
	@find $(BACKUP_DIR) -name "incremental-*" -type f | sort -r | head -1 || echo "  None"
	@echo ""
	@echo "Total backups:"
	@find $(BACKUP_DIR) -name "*.tar.gz*" -type f | wc -l
	@echo ""
	@echo "Total size:"
	@du -sh $(BACKUP_DIR) 2>/dev/null || echo "  Unknown"
	@echo ""
	@echo "Disk usage:"
	@df -h $(BACKUP_DIR) | tail -1
	@echo ""

.DEFAULT_GOAL := help
