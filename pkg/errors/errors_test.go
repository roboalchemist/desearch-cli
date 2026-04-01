package errors

import (
	"errors"
	"testing"
)

func TestSystemError(t *testing.T) {
	err := Wrap("something went wrong")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Error() != "something went wrong" {
		t.Errorf("got %q, want %q", err.Error(), "something went wrong")
	}
	if !IsSystem(err) {
		t.Error("expected IsSystem to return true for wrapped error")
	}
}

func TestIsSystem_NonSystem(t *testing.T) {
	err := errors.New("plain error")
	if IsSystem(err) {
		t.Error("expected IsSystem to return false for plain error")
	}
}

func TestIsSystem_Nil(t *testing.T) {
	if IsSystem(nil) {
		t.Error("expected IsSystem to return false for nil")
	}
}

func TestIsSystem_Wrapped(t *testing.T) {
	inner := Wrap("inner system error")
	outer := errors.New("outer: " + inner.Error())
	if !IsSystem(inner) {
		t.Error("expected IsSystem to return true for inner system error")
	}
	// outer is not a SystemError even though its message mentions a system error
	if IsSystem(outer) {
		t.Error("expected IsSystem to return false for outer non-system error")
	}
}

func TestWrapF(t *testing.T) {
	err := WrapF("value=%d", 42)
	if !IsSystem(err) {
		t.Error("expected WrapF to produce a system error")
	}
	if err.Error() != "value=42" {
		t.Errorf("got %q, want %q", err.Error(), "value=42")
	}
}

func TestUsageError(t *testing.T) {
	inner := errors.New("unknown flag --foo")
	err := WrapUsage(inner)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Error() != "unknown flag --foo" {
		t.Errorf("got %q, want %q", err.Error(), "unknown flag --foo")
	}
	if !IsUsage(err) {
		t.Error("expected IsUsage to return true for wrapped error")
	}
}

func TestIsUsage_NonUsage(t *testing.T) {
	err := errors.New("plain error")
	if IsUsage(err) {
		t.Error("expected IsUsage to return false for plain error")
	}
}

func TestIsUsage_Nil(t *testing.T) {
	if IsUsage(nil) {
		t.Error("expected IsUsage to return false for nil")
	}
}

func TestWrapUsage_Nil(t *testing.T) {
	if WrapUsage(nil) != nil {
		t.Error("expected WrapUsage(nil) to return nil")
	}
}

func TestIsUsage_NotSystem(t *testing.T) {
	sysErr := Wrap("system error")
	if IsUsage(sysErr) {
		t.Error("expected IsUsage to return false for SystemError")
	}
}

func TestIsSystem_NotUsage(t *testing.T) {
	usageErr := WrapUsage(errors.New("bad flag"))
	if IsSystem(usageErr) {
		t.Error("expected IsSystem to return false for UsageError")
	}
}
