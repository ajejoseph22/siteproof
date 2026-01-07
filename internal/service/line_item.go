package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ajejoseph22/siteproof/internal/domain"
)

type lineItemScanner interface {
	Scan(dest ...any) error
}

// ListLineItems retrieves line items for a ticket.
func (s *TicketService) ListLineItems(ctx context.Context, companyID, ticketID uuid.UUID) ([]*domain.TicketLineItem, error) {
	if err := s.ensureTicketExists(ctx, companyID, ticketID); err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, ticket_id, item_type, description,
		       quantity::float8, unit, unit_rate::float8, total::float8,
		       sort_order, created_at
		FROM ticket_line_items
		WHERE ticket_id = $1
		ORDER BY sort_order, created_at
	`, ticketID)
	if err != nil {
		return nil, fmt.Errorf("list line items: %w", err)
	}
	defer rows.Close()

	items := []*domain.TicketLineItem{}
	for rows.Next() {
		item, err := scanLineItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan line item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// CreateLineItem adds a line item to a ticket.
func (s *TicketService) CreateLineItem(ctx context.Context, companyID, ticketID, actorID uuid.UUID, input *domain.CreateLineItemInput) (*domain.TicketLineItem, error) {
	input.TicketID = ticketID
	if err := input.Validate(); err != nil {
		return nil, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin line item create: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	status, previousTotal, err := s.lockTicketForLineItems(ctx, tx, companyID, ticketID)
	if err != nil {
		return nil, err
	}
	if status != domain.TicketStatusDraft {
		return nil, domain.ErrTicketLocked
	}

	row := tx.QueryRow(ctx, `
		INSERT INTO ticket_line_items (
			ticket_id, item_type, description, quantity, unit, unit_rate, total, sort_order
		)
		VALUES ($1, $2, $3, $4, $5, $6,
			CASE WHEN $6 IS NULL THEN NULL ELSE $4 * $6 END,
			$7
		)
		RETURNING id, ticket_id, item_type, description,
		          quantity::float8, unit, unit_rate::float8, total::float8,
		          sort_order, created_at
	`, ticketID, string(input.ItemType), input.Description, input.Quantity, input.Unit, input.UnitRate, input.SortOrder)

	item, err := scanLineItem(row)
	if err != nil {
		return nil, fmt.Errorf("create line item: %w", err)
	}

	currentTotal, err := s.recalculateLineItemsTotal(ctx, tx, ticketID)
	if err != nil {
		return nil, err
	}

	changes := lineItemChanges(nil, item)
	if previousTotal != currentTotal {
		changes["line_items_total"] = domain.ChangeValue{Old: previousTotal, New: currentTotal}
	}

	if err := s.insertAuditLog(ctx, tx, domain.CreateAuditLogInput{
		EntityType: "ticket",
		EntityID:   ticketID,
		Action:     domain.AuditActionLineItemAdded,
		ActorID:    &actorID,
		Changes:    changes,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit line item create: %w", err)
	}

	return item, nil
}

// UpdateLineItem updates a line item on a ticket.
func (s *TicketService) UpdateLineItem(ctx context.Context, companyID, ticketID, lineItemID, actorID uuid.UUID, input *domain.UpdateLineItemInput) (*domain.TicketLineItem, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin line item update: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	status, previousTotal, err := s.lockTicketForLineItems(ctx, tx, companyID, ticketID)
	if err != nil {
		return nil, err
	}
	if status != domain.TicketStatusDraft {
		return nil, domain.ErrTicketLocked
	}

	existing, err := s.getLineItemForUpdate(ctx, tx, ticketID, lineItemID)
	if err != nil {
		return nil, err
	}

	itemType := lineItemTypePointer(input.ItemType)
	row := tx.QueryRow(ctx, `
		UPDATE ticket_line_items
		SET item_type = COALESCE($3, item_type),
		    description = COALESCE($4, description),
		    quantity = COALESCE($5, quantity),
		    unit = COALESCE($6, unit),
		    unit_rate = COALESCE($7, unit_rate),
		    total = CASE
		        WHEN COALESCE($7, unit_rate) IS NULL THEN NULL
		        ELSE COALESCE($5, quantity) * COALESCE($7, unit_rate)
		    END,
		    sort_order = COALESCE($8, sort_order)
		WHERE id = $1 AND ticket_id = $2
		RETURNING id, ticket_id, item_type, description,
		          quantity::float8, unit, unit_rate::float8, total::float8,
		          sort_order, created_at
	`, lineItemID, ticketID, itemType, input.Description, input.Quantity, input.Unit, input.UnitRate, input.SortOrder)

	updated, err := scanLineItem(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("update line item: %w", err)
	}

	currentTotal, err := s.recalculateLineItemsTotal(ctx, tx, ticketID)
	if err != nil {
		return nil, err
	}

	changes := lineItemChanges(existing, updated)
	if previousTotal != currentTotal {
		changes["line_items_total"] = domain.ChangeValue{Old: previousTotal, New: currentTotal}
	}

	if err := s.insertAuditLog(ctx, tx, domain.CreateAuditLogInput{
		EntityType: "ticket",
		EntityID:   ticketID,
		Action:     domain.AuditActionLineItemUpdated,
		ActorID:    &actorID,
		Changes:    changes,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit line item update: %w", err)
	}

	return updated, nil
}

// DeleteLineItem removes a line item from a ticket.
func (s *TicketService) DeleteLineItem(ctx context.Context, companyID, ticketID, lineItemID, actorID uuid.UUID) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin line item delete: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	status, previousTotal, err := s.lockTicketForLineItems(ctx, tx, companyID, ticketID)
	if err != nil {
		return err
	}
	if status != domain.TicketStatusDraft {
		return domain.ErrTicketLocked
	}

	existing, err := s.getLineItemForUpdate(ctx, tx, ticketID, lineItemID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `DELETE FROM ticket_line_items WHERE id = $1 AND ticket_id = $2`, lineItemID, ticketID)
	if err != nil {
		return fmt.Errorf("delete line item: %w", err)
	}

	currentTotal, err := s.recalculateLineItemsTotal(ctx, tx, ticketID)
	if err != nil {
		return err
	}

	changes := lineItemChanges(existing, nil)
	if previousTotal != currentTotal {
		changes["line_items_total"] = domain.ChangeValue{Old: previousTotal, New: currentTotal}
	}

	if err := s.insertAuditLog(ctx, tx, domain.CreateAuditLogInput{
		EntityType: "ticket",
		EntityID:   ticketID,
		Action:     domain.AuditActionLineItemDeleted,
		ActorID:    &actorID,
		Changes:    changes,
	}); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit line item delete: %w", err)
	}

	return nil
}

func (s *TicketService) ensureTicketExists(ctx context.Context, companyID, ticketID uuid.UUID) error {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM tickets WHERE id = $1 AND company_id = $2
		)
	`, ticketID, companyID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check ticket exists: %w", err)
	}
	if !exists {
		return domain.ErrNotFound
	}
	return nil
}

func (s *TicketService) lockTicketForLineItems(ctx context.Context, tx pgx.Tx, companyID, ticketID uuid.UUID) (domain.TicketStatus, float64, error) {
	var status domain.TicketStatus
	var total pgtype.Float8
	if err := tx.QueryRow(ctx, `
		SELECT status, line_items_total::float8
		FROM tickets
		WHERE id = $1 AND company_id = $2
		FOR UPDATE
	`, ticketID, companyID).Scan(&status, &total); err != nil {
		if err == pgx.ErrNoRows {
			return "", 0, domain.ErrNotFound
		}
		return "", 0, fmt.Errorf("lock ticket: %w", err)
	}

	return status, floatFromPG(total), nil
}

func (s *TicketService) recalculateLineItemsTotal(ctx context.Context, tx pgx.Tx, ticketID uuid.UUID) (float64, error) {
	var total float64
	if err := tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(total), 0)::float8
		FROM ticket_line_items
		WHERE ticket_id = $1
	`, ticketID).Scan(&total); err != nil {
		return 0, fmt.Errorf("sum line items: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE tickets
		SET line_items_total = $2
		WHERE id = $1
	`, ticketID, total); err != nil {
		return 0, fmt.Errorf("update ticket totals: %w", err)
	}

	return total, nil
}

func (s *TicketService) getLineItemForUpdate(ctx context.Context, tx pgx.Tx, ticketID, lineItemID uuid.UUID) (*domain.TicketLineItem, error) {
	row := tx.QueryRow(ctx, `
		SELECT id, ticket_id, item_type, description,
		       quantity::float8, unit, unit_rate::float8, total::float8,
		       sort_order, created_at
		FROM ticket_line_items
		WHERE id = $1 AND ticket_id = $2
		FOR UPDATE
	`, lineItemID, ticketID)

	item, err := scanLineItem(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get line item: %w", err)
	}

	return item, nil
}

func scanLineItem(scanner lineItemScanner) (*domain.TicketLineItem, error) {
	var item domain.TicketLineItem
	var unitRate pgtype.Float8
	var total pgtype.Float8

	if err := scanner.Scan(
		&item.ID,
		&item.TicketID,
		&item.ItemType,
		&item.Description,
		&item.Quantity,
		&item.Unit,
		&unitRate,
		&total,
		&item.SortOrder,
		&item.CreatedAt,
	); err != nil {
		return nil, err
	}

	item.UnitRate = floatPointer(unitRate)
	item.Total = floatPointer(total)

	return &item, nil
}

func floatPointer(value pgtype.Float8) *float64 {
	if !value.Valid {
		return nil
	}
	v := value.Float64
	return &v
}

func floatFromPG(value pgtype.Float8) float64 {
	if !value.Valid {
		return 0
	}
	return value.Float64
}

func lineItemTypePointer(input *domain.LineItemType) *string {
	if input == nil {
		return nil
	}
	value := string(*input)
	return &value
}

func lineItemChanges(oldItem, newItem *domain.TicketLineItem) map[string]domain.ChangeValue {
	changes := map[string]domain.ChangeValue{}

	if oldItem == nil || newItem == nil {
		changes["line_item_id"] = domain.ChangeValue{Old: idOrNil(oldItem), New: idOrNil(newItem)}
		changes["ticket_id"] = domain.ChangeValue{Old: ticketIDOrNil(oldItem), New: ticketIDOrNil(newItem)}
		changes["item_type"] = domain.ChangeValue{Old: itemTypeOrNil(oldItem), New: itemTypeOrNil(newItem)}
		changes["description"] = domain.ChangeValue{Old: descriptionOrNil(oldItem), New: descriptionOrNil(newItem)}
		changes["quantity"] = domain.ChangeValue{Old: quantityOrNil(oldItem), New: quantityOrNil(newItem)}
		changes["unit"] = domain.ChangeValue{Old: unitOrNil(oldItem), New: unitOrNil(newItem)}
		changes["unit_rate"] = domain.ChangeValue{Old: unitRateOrNil(oldItem), New: unitRateOrNil(newItem)}
		changes["total"] = domain.ChangeValue{Old: totalOrNil(oldItem), New: totalOrNil(newItem)}
		changes["sort_order"] = domain.ChangeValue{Old: sortOrderOrNil(oldItem), New: sortOrderOrNil(newItem)}
		return changes
	}

	if oldItem.ItemType != newItem.ItemType {
		changes["item_type"] = domain.ChangeValue{Old: oldItem.ItemType, New: newItem.ItemType}
	}
	if oldItem.Description != newItem.Description {
		changes["description"] = domain.ChangeValue{Old: oldItem.Description, New: newItem.Description}
	}
	if oldItem.Quantity != newItem.Quantity {
		changes["quantity"] = domain.ChangeValue{Old: oldItem.Quantity, New: newItem.Quantity}
	}
	if oldItem.Unit != newItem.Unit {
		changes["unit"] = domain.ChangeValue{Old: oldItem.Unit, New: newItem.Unit}
	}
	if !floatPointerEqual(oldItem.UnitRate, newItem.UnitRate) {
		changes["unit_rate"] = domain.ChangeValue{Old: oldItem.UnitRate, New: newItem.UnitRate}
	}
	if !floatPointerEqual(oldItem.Total, newItem.Total) {
		changes["total"] = domain.ChangeValue{Old: oldItem.Total, New: newItem.Total}
	}
	if oldItem.SortOrder != newItem.SortOrder {
		changes["sort_order"] = domain.ChangeValue{Old: oldItem.SortOrder, New: newItem.SortOrder}
	}

	return changes
}

func floatPointerEqual(a, b *float64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func idOrNil(item *domain.TicketLineItem) interface{} {
	if item == nil {
		return nil
	}
	return item.ID
}

func ticketIDOrNil(item *domain.TicketLineItem) interface{} {
	if item == nil {
		return nil
	}
	return item.TicketID
}

func itemTypeOrNil(item *domain.TicketLineItem) interface{} {
	if item == nil {
		return nil
	}
	return item.ItemType
}

func descriptionOrNil(item *domain.TicketLineItem) interface{} {
	if item == nil {
		return nil
	}
	return item.Description
}

func quantityOrNil(item *domain.TicketLineItem) interface{} {
	if item == nil {
		return nil
	}
	return item.Quantity
}

func unitOrNil(item *domain.TicketLineItem) interface{} {
	if item == nil {
		return nil
	}
	return item.Unit
}

func unitRateOrNil(item *domain.TicketLineItem) interface{} {
	if item == nil {
		return nil
	}
	return item.UnitRate
}

func totalOrNil(item *domain.TicketLineItem) interface{} {
	if item == nil {
		return nil
	}
	return item.Total
}

func sortOrderOrNil(item *domain.TicketLineItem) interface{} {
	if item == nil {
		return nil
	}
	return item.SortOrder
}
