package main

import "testing"

func TestStatusColor(t *testing.T) {
	cases := map[string]interface{}{
		"idle":    colGreen,
		"working": colYellow,
		"blocked": colRed,
		"done":    colTeal,
		"unknown": colOverlay,
		"weird":   colOverlay, // anything unmapped → grey
	}
	for status, want := range cases {
		if got := statusColor(status); got != want {
			t.Errorf("statusColor(%q) = %v, want %v", status, got, want)
		}
	}
}
