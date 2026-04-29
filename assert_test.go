package id

func assertCmpEqual[T comparable](t interface {
	Errorf(format string, args ...any)
}, got, want T,
) {
	if got != want {
		t.Errorf("expected %v, got %v", want, got)
	}
}
