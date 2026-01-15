BEGIN;

-- Add advanced order type fields to orders table
-- These fields support SL/TP, trailing stops, OCO, and pending orders

ALTER TABLE orders
ADD COLUMN IF NOT EXISTS parent_position_id BIGINT REFERENCES positions(id) ON DELETE CASCADE,
ADD COLUMN IF NOT EXISTS trailing_delta DECIMAL(20, 8),
ADD COLUMN IF NOT EXISTS expiry_time TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS oco_link_id BIGINT REFERENCES orders(id) ON DELETE SET NULL;

-- Add comments for clarity
COMMENT ON COLUMN orders.parent_position_id IS 'Links SL/TP orders to their parent position (NULL for standalone orders)';
COMMENT ON COLUMN orders.trailing_delta IS 'Distance for trailing stops in price units (NULL for non-trailing)';
COMMENT ON COLUMN orders.expiry_time IS 'Auto-cancel after this time (NULL for GTC - Good Till Cancelled)';
COMMENT ON COLUMN orders.oco_link_id IS 'References another order for OCO (One Cancels Other) pairing (NULL if no link)';

-- Add index for monitoring pending orders with triggers
CREATE INDEX IF NOT EXISTS idx_orders_pending_triggers ON orders(status, trigger_price)
WHERE status = 'PENDING' AND trigger_price IS NOT NULL;

-- Add index for position-linked orders (SL/TP lookups)
CREATE INDEX IF NOT EXISTS idx_orders_position ON orders(parent_position_id)
WHERE parent_position_id IS NOT NULL;

-- Add index for expiring orders
CREATE INDEX IF NOT EXISTS idx_orders_expiry ON orders(expiry_time)
WHERE expiry_time IS NOT NULL AND status = 'PENDING';

COMMIT;
