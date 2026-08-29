package syncinit

import "testing"

func TestStaticInitializers(t *testing.T) {
	t.Parallel()

	if CONDITION_VARIABLE_INIT != 0 || INIT_ONCE_STATIC_INIT != 0 || SRWLOCK_INIT != 0 {
		t.Fatal("static synchronization initializers must be zero")
	}
}
