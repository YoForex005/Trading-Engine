BEGIN;

-- Remove indexes
DROP INDEX IF EXISTS idx_orders_expiry;
DROP INDEX IF EXISTS idx_orders_position;
DROP INDEX IF EXISTS idx_orders_pending_triggers;

-- Remove columns (in reverse order of dependencies)
ALTER TABLE orders
DROP COLUMN IF EXISTS oco_link_id,
DROP COLUMN IF EXISTS expiry_time,
DROP COLUMN IF EXISTS trailing_delta,
DROP COLUMN IF EXISTS parent_position_id;

COMMIT;
