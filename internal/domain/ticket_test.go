package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestCreateLineItemInputValidate(t *testing.T) {
	unitRate := 95.0

	valid := &CreateLineItemInput{
		TicketID:    uuid.New(),
		ItemType:    LineItemTypeLabor,
		Description: "Overtime labor",
		Quantity:    2.5,
		Unit:        "hours",
		UnitRate:    &unitRate,
		SortOrder:   0,
	}

	if err := valid.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	invalid := &CreateLineItemInput{
		TicketID:    uuid.Nil,
		ItemType:    LineItemType("bad"),
		Description: "",
		Quantity:    0,
		Unit:        "",
		UnitRate:    floatPointer(-1),
		SortOrder:   -1,
	}

	if err := invalid.Validate(); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestUpdateLineItemInputValidate(t *testing.T) {
	if err := (&UpdateLineItemInput{}).Validate(); err == nil {
		t.Fatal("expected validation error for empty update")
	}

	invalidType := LineItemType("bad")
	invalidQuantity := -1.0
	invalid := &UpdateLineItemInput{
		ItemType: &invalidType,
		Quantity: &invalidQuantity,
	}
	if err := invalid.Validate(); err == nil {
		t.Fatal("expected validation error for invalid fields")
	}

	quantity := 1.25
	unit := "hours"
	valid := &UpdateLineItemInput{
		Quantity: &quantity,
		Unit:     &unit,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func floatPointer(value float64) *float64 {
	return &value
}
