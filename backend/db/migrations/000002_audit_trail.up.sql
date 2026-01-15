BEGIN;

-- Audit schema for all change tracking
CREATE SCHEMA IF NOT EXISTS audit;

-- Audit log table (stores all changes to critical entities)
CREATE TABLE audit.logged_actions (
    event_id BIGSERIAL PRIMARY KEY,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    relid OID NOT NULL,
    session_user_name TEXT,
    action_tstamp_tx TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    action_tstamp_stm TIMESTAMPTZ NOT NULL DEFAULT CLOCK_TIMESTAMP(),
    action_tstamp_clk TIMESTAMPTZ NOT NULL DEFAULT CLOCK_TIMESTAMP(),
    transaction_id BIGINT,
    application_name TEXT,
    client_addr INET,
    client_port INTEGER,
    action TEXT NOT NULL CHECK (action IN ('I', 'U', 'D')),  -- INSERT, UPDATE, DELETE
    row_data JSONB,           -- Full row data before change (for UPDATE/DELETE)
    changed_fields JSONB,     -- Only changed fields (for UPDATE)
    statement_only BOOLEAN NOT NULL
);

-- Indexes for efficient audit queries
CREATE INDEX idx_audit_table_name ON audit.logged_actions(table_name);
CREATE INDEX idx_audit_action_tstamp ON audit.logged_actions(action_tstamp_tx);
CREATE INDEX idx_audit_relid ON audit.logged_actions(relid);
CREATE INDEX idx_audit_row_data ON audit.logged_actions USING gin(row_data);

-- Generic audit trigger function (works for any table)
CREATE OR REPLACE FUNCTION audit.if_modified_func() RETURNS TRIGGER AS $body$
DECLARE
    audit_row audit.logged_actions;
    include_values BOOLEAN;
    log_diffs BOOLEAN;
    h_old JSONB;
    h_new JSONB;
    excluded_cols TEXT[] = ARRAY[]::TEXT[];
BEGIN
    IF TG_WHEN <> 'AFTER' THEN
        RAISE EXCEPTION 'audit.if_modified_func() may only run as an AFTER trigger';
    END IF;

    audit_row = ROW(
        nextval('audit.logged_actions_event_id_seq'),  -- event_id
        TG_TABLE_SCHEMA::TEXT,                         -- schema_name
        TG_TABLE_NAME::TEXT,                           -- table_name
        TG_RELID,                                      -- relid
        session_user::TEXT,                            -- session_user_name
        CURRENT_TIMESTAMP,                             -- action_tstamp_tx
        statement_timestamp(),                         -- action_tstamp_stm
        clock_timestamp(),                             -- action_tstamp_clk
        txid_current(),                                -- transaction_id
        current_setting('application_name'),           -- application_name
        inet_client_addr(),                            -- client_addr
        inet_client_port(),                            -- client_port
        substring(TG_OP, 1, 1),                        -- action
        NULL,                                          -- row_data
        NULL,                                          -- changed_fields
        'f'                                            -- statement_only
    );

    IF NOT TG_ARGV[0]::BOOLEAN IS DISTINCT FROM 'f'::BOOLEAN THEN
        audit_row.client_query = NULL;
    END IF;

    IF TG_ARGV[1] IS NOT NULL THEN
        excluded_cols = TG_ARGV[1]::TEXT[];
    END IF;

    IF (TG_OP = 'UPDATE' AND TG_LEVEL = 'ROW') THEN
        h_old = to_jsonb(OLD.*) - excluded_cols;
        h_new = to_jsonb(NEW.*) - excluded_cols;
        audit_row.row_data = h_old;
        audit_row.changed_fields = h_new - h_old;

        IF audit_row.changed_fields = '{}'::JSONB THEN
            -- No actual changes, skip audit log
            RETURN NULL;
        END IF;
    ELSIF (TG_OP = 'DELETE' AND TG_LEVEL = 'ROW') THEN
        audit_row.row_data = to_jsonb(OLD.*) - excluded_cols;
    ELSIF (TG_OP = 'INSERT' AND TG_LEVEL = 'ROW') THEN
        audit_row.row_data = to_jsonb(NEW.*) - excluded_cols;
    ELSE
        RAISE EXCEPTION '[audit.if_modified_func] - Trigger func added as trigger for unhandled case: %, %', TG_OP, TG_LEVEL;
        RETURN NULL;
    END IF;

    INSERT INTO audit.logged_actions VALUES (audit_row.*);
    RETURN NULL;
END;
$body$
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, public;

COMMENT ON FUNCTION audit.if_modified_func() IS $body$
Track changes to a table at the statement and/or row level.

Optional parameters to trigger in CREATE TRIGGER call:
    param 0: boolean, whether to log the query text. Default 't'.
    param 1: text[], columns to ignore in updates. Default [].

Example:
    CREATE TRIGGER accounts_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON accounts
    FOR EACH ROW EXECUTE FUNCTION audit.if_modified_func('t', '{updated_at}');
$body$;

-- Attach audit triggers to accounts table
CREATE TRIGGER accounts_audit_trigger
AFTER INSERT OR UPDATE OR DELETE ON accounts
FOR EACH ROW EXECUTE FUNCTION audit.if_modified_func('t', '{updated_at}');

-- Attach audit triggers to positions table
CREATE TRIGGER positions_audit_trigger
AFTER INSERT OR UPDATE OR DELETE ON positions
FOR EACH ROW EXECUTE FUNCTION audit.if_modified_func('t', '{updated_at}');

-- Attach audit triggers to orders table
CREATE TRIGGER orders_audit_trigger
AFTER INSERT OR UPDATE OR DELETE ON orders
FOR EACH ROW EXECUTE FUNCTION audit.if_modified_func('t');

-- DO NOT attach trigger to trades table
-- Trades are immutable (insert-only), no updates or deletes allowed
-- Audit trail would duplicate data with no value

COMMIT;
