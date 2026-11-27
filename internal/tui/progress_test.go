package tui

import (
	"strings"
	"testing"
)

func TestDownloadModelKeepsPercentageOnBar(t *testing.T) {
	m := NewDownloadModel()
	next, _ := m.Update(DownloadProgressMsg{Downloaded: 50, Total: 100})
	m = next.(DownloadModel)
	view := m.View()
	status := strings.SplitN(view, "\n", 2)[0]
	if strings.Contains(status, "%") {
		t.Fatalf("status line should not repeat the percentage: %q", status)
	}
	if !strings.Contains(view, "Downloading...") || !strings.Contains(view, "%") {
		t.Fatalf("expected status line and bar percentage: %q", view)
	}
}
