package duration

import (
	"testing"
	"time"
)

func TestCustomFormat(t *testing.T) {
	d := 389*24*time.Hour +
		12*time.Hour +
		31*time.Minute +
		54*time.Second +
		346*time.Millisecond

	f := `{{.Years}} - {{.Days}} - {{.Hours}} - {{.Minutes}} - {{.Seconds}}`
	e := `1 - 24 - 12 - 31 - 54`

	if s, _ := CustomHumanizeDuration(d, f); s != e {
		t.Errorf("Got unexpected result: expected=%q result=%q", e, s)
	}
}

func TestDefaultFormat(t *testing.T) {
	d := 389*24*time.Hour +
		12*time.Hour +
		31*time.Minute +
		54*time.Second +
		346*time.Millisecond

	e := `1 year, 24 days, 12 hours, 31 minutes, 54 seconds`

	if s := HumanizeDuration(d); s != e {
		t.Errorf("Got unexpected result: expected=%q result=%q", e, s)
	}
}
