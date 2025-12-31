-- Migration: 000006_ticket_line_items
-- Description: Create line items for labor, materials, and equipment on tickets
-- Author: SiteProof
-- Date: 2024-12-31

CREATE TABLE ticket_line_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id       UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    
    -- Item details
    item_type       TEXT NOT NULL,
    description     TEXT NOT NULL,
    quantity        DECIMAL(10, 2) NOT NULL,
    unit            TEXT NOT NULL,
    unit_rate       DECIMAL(10, 2),
    total           DECIMAL(10, 2),
    
    -- Ordering
    sort_order      INTEGER NOT NULL DEFAULT 0,
    
    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    -- Constraints
    CONSTRAINT line_items_type_valid CHECK (item_type IN ('labor', 'material', 'equipment')),
    CONSTRAINT line_items_description_not_empty CHECK (length(trim(description)) > 0),
    CONSTRAINT line_items_quantity_positive CHECK (quantity > 0),
    CONSTRAINT line_items_unit_not_empty CHECK (length(trim(unit)) > 0),
    CONSTRAINT line_items_rate_non_negative CHECK (unit_rate IS NULL OR unit_rate >= 0),
    CONSTRAINT line_items_total_non_negative CHECK (total IS NULL OR total >= 0)
);

-- Indexes
CREATE INDEX idx_line_items_ticket ON ticket_line_items(ticket_id);
CREATE INDEX idx_line_items_sort ON ticket_line_items(ticket_id, sort_order);

-- Comments
COMMENT ON TABLE ticket_line_items IS 'Line items on tickets capturing extra labor, materials, and equipment';
COMMENT ON COLUMN ticket_line_items.item_type IS 'Type of line item: labor, material, or equipment';
COMMENT ON COLUMN ticket_line_items.unit IS 'Unit of measure (e.g., hours, each, linear_ft, day)';
COMMENT ON COLUMN ticket_line_items.unit_rate IS 'Optional price per unit';
COMMENT ON COLUMN ticket_line_items.total IS 'Calculated or manually entered total (quantity × unit_rate)';

