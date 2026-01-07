-- Migration: 000017_ticket_line_item_totals
-- Description: Remove line item total aggregation from tickets
-- Author: SiteProof
-- Date: 2025-01-01

ALTER TABLE tickets
    DROP COLUMN line_items_total;
