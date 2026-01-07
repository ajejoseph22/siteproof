-- Migration: 000017_ticket_line_item_totals
-- Description: Add line item total aggregation to tickets
-- Author: SiteProof
-- Date: 2025-01-01

ALTER TABLE tickets
    ADD COLUMN line_items_total DECIMAL(10, 2) NOT NULL DEFAULT 0;

UPDATE tickets
SET line_items_total = COALESCE(
    (SELECT SUM(total) FROM ticket_line_items WHERE ticket_id = tickets.id),
    0
);

COMMENT ON COLUMN tickets.line_items_total IS 'Sum of line item totals for the ticket.';
