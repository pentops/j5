package date_j5t

import (
	"encoding"
	"testing"
	"time"
)

// Type Assertions
var _ encoding.TextMarshaler = NewDate(2019, 1, 20)
var _ encoding.TextUnmarshaler = NewDate(2019, 1, 20)

func TestDate(t *testing.T) {

	dv := NewDate(2019, 1, 20)
	dvt := dv.AsTime(time.UTC)
	if y, m, d := dvt.Date(); y != 2019 || m != time.January || d != 20 {
		t.Errorf("Date mismatch: %v", dvt)
	}

	if ds := dv.DateString(); ds != "2019-01-20" {
		t.Errorf("DateString mismatch: %v", ds)
	}

	if dsByte, err := dv.MarshalText(); err != nil {
		t.Errorf("MarshalText error: %v", err)
	} else if string(dsByte) != "2019-01-20" {
		t.Errorf("MarshalText mismatch: %v", string(dsByte))
	}

	if err := dv.UnmarshalText([]byte("2019-01-20")); err != nil {
		t.Errorf("UnmarshalText error: %v", err)
	}

	if dv.Year != 2019 || dv.Month != 1 || dv.Day != 20 {
		t.Errorf("UnmarshalText mismatch: %v", dv)
	}

	added := dv.AddDate(1, 1, 1)
	if str := added.DateString(); str != "2020-02-21" {
		t.Errorf("AddDate mismatch: %v", str)
	}

	// Leap Year
	if str := NewDate(2020, 2, 28).AddDate(0, 0, 1).DateString(); str != "2020-02-29" {
		t.Errorf("AddDate mismatch: %v", str)
	}

	// Not Leap Year
	if str := NewDate(2021, 2, 28).AddDate(0, 0, 1).DateString(); str != "2021-03-01" {
		t.Errorf("AddDate mismatch: %v", str)
	}
}
