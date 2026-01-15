BEGIN;

-- Drop triggers first (depend on function)
DROP TRIGGER IF EXISTS accounts_audit_trigger ON accounts;
DROP TRIGGER IF EXISTS positions_audit_trigger ON positions;
DROP TRIGGER IF EXISTS orders_audit_trigger ON orders;

-- Drop audit function
DROP FUNCTION IF EXISTS audit.if_modified_func();

-- Drop audit table
DROP TABLE IF EXISTS audit.logged_actions;

-- Drop audit schema (only if empty)
DROP SCHEMA IF EXISTS audit CASCADE;

COMMIT;
