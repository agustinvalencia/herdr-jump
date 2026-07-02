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

func TestStatusRankAttentionFirst(t *testing.T) {
	// blocked and done (need you) must sort above active/idle work.
	if !(statusRank("blocked") < statusRank("done") &&
		statusRank("done") < statusRank("working") &&
		statusRank("working") < statusRank("idle")) {
		t.Fatalf("rank order wrong: blocked=%d done=%d working=%d idle=%d",
			statusRank("blocked"), statusRank("done"), statusRank("working"), statusRank("idle"))
	}
}
