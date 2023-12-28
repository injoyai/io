package io

import "testing"

func TestDebug_Debug(t *testing.T) {
	type A struct {
		Debugger
	}

	a := A{}
	a.Debug(true)
	t.Log(a.Debugger)
	if a.Debugger {
		t.Log("true")
	}
}
