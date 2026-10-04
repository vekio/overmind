package app_test

import "uuid"

// testIDGenerator is shared by creation tests; repository fakes live with their use cases.
type testIDGenerator struct{}

func (testIDGenerator) Generate() uuid.UUID { return uuid.New() }
