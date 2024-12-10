package decimal_j5t

import (
	"testing"
)

func TestDecimal(t *testing.T) {
	df := FromFloat(1.33)
	di := FromInt(2)

	result, err := df.Mul(FromFloat(2.66))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.ToString() != "3.5378" {
		t.Errorf("expected string to be 3.5378, got %v", result.Value)
	}

	result, err = di.Mul(FromFloat(1.33))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.ToString() != "2.66" {
		t.Errorf("expected string to be 2.66, got %v", result.Value)
	}

	result, err = di.Mul(FromInt(0))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.ToString() != "0" {
		t.Errorf("expected string to be 0, got %v", result.Value)
	}

	result, err = df.Div(di)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.ToString() != "0.665" {
		t.Errorf("expected string to be 0.665, got %v", result.Value)
	}

	result, err = df.Div(FromInt(0))
	if err == nil {
		t.Errorf("expected a divide by zero error")
	}
	if result != nil {
		t.Errorf("expected result to be nil, got %v", result)
	}

	result, err = FromInt(0).Div(FromInt(1))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.ToString() != "0" {
		t.Errorf("expected string to be 0, got %v", result.Value)
	}

	result, err = FromInt(2).Div(FromFloat(.5))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.ToString() != "4" {
		t.Errorf("expected string to be 4, got %v", result.Value)
	}
}
