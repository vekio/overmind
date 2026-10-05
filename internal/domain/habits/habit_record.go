package habits

import (
	"errors"
	"fmt"
	"math"
	"time"
	"uuid"
)

// ErrInvalidHabitRecord identifies invalid occurrence identity, quantity or time.
var ErrInvalidHabitRecord = errors.New("invalid habit record")

// HabitRecord is an entity representing a nonnegative quantity performed at an
// instant, in the past or future. To correct a mistake, delete the record
// and create a new one. Its identity and habit association remain fixed. Unit comes from
// the associated habit's immutable goal; occurrences are not daily totals.
type HabitRecord struct {
	id         uuid.UUID
	habitID    uuid.UUID
	value      float64
	occurredAt time.Time
}

func newHabitRecord(id, habitID uuid.UUID, value float64, at time.Time) (*HabitRecord, error) {
	record := &HabitRecord{id: id, habitID: habitID, value: value, occurredAt: at.Round(0)}
	if err := record.validate(); err != nil {
		return nil, err
	}
	return record, nil
}

func (record *HabitRecord) validate() error {
	if record == nil || record.id == uuid.Nil() || record.habitID == uuid.Nil() {
		return fmt.Errorf("%w: record and habit IDs are required", ErrInvalidHabitRecord)
	}
	if record.value < 0 || math.IsNaN(record.value) || math.IsInf(record.value, 0) {
		return fmt.Errorf("%w: value must be finite and nonnegative", ErrInvalidHabitRecord)
	}
	if record.occurredAt.IsZero() {
		return fmt.Errorf("%w: occurrence time is required", ErrInvalidHabitRecord)
	}
	return nil
}

// ID returns the occurrence identity.
func (record *HabitRecord) ID() uuid.UUID { return record.id }

// HabitID returns the identity of the associated habit.
func (record *HabitRecord) HabitID() uuid.UUID { return record.habitID }

// Value returns the finite nonnegative quantity measured for this occurrence.
func (record *HabitRecord) Value() float64 { return record.value }

// OccurredAt returns the occurrence instant with its original timezone and no monotonic reading.
func (record *HabitRecord) OccurredAt() time.Time { return record.occurredAt }
