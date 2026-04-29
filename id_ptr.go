package id

// Ptr returns a pointer to the ID. Useful for optional ID fields in API payloads.
func (id ID[B, V]) Ptr() *ID[B, V] { return &id }

// FromPtr dereferences a pointer-to-ID, returning the zero value if the pointer is nil.
func FromPtr[B any, V comparable](p *ID[B, V]) ID[B, V] {
	if p == nil {
		return ID[B, V]{} //nolint:exhaustruct // intentional zero value
	}

	return *p
}
