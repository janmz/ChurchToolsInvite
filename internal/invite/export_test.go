package invite

import (
	"context"
	"time"
)

// SetSleepForTest replaces the delay function and returns a restore func.
func SetSleepForTest(fn func(time.Duration)) func() {
	orig := sleep
	sleep = func(_ context.Context, d time.Duration) { fn(d) }
	return func() { sleep = orig }
}
