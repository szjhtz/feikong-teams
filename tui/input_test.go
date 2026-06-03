package tui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func TestInputCtrlCRequiresConfirmation(t *testing.T) {
	model := newInputModel("> ", nil)

	updated, cmd := model.Update(ctrlCKeyMsg())
	model = updated.(inputModel)
	if model.ctrlC {
		t.Fatal("first Ctrl+C should only request exit confirmation")
	}
	if cmd == nil {
		t.Fatal("first Ctrl+C should start exit confirmation countdown")
	}
	if !strings.Contains(model.View().Content, "再按 ") {
		t.Fatal("exit confirmation warning was not rendered")
	}

	updated, _ = model.Update(ctrlCKeyMsg())
	model = updated.(inputModel)
	if !model.ctrlC {
		t.Fatal("second Ctrl+C during confirmation should request exit")
	}
}

func TestInputCtrlCConfirmationExpires(t *testing.T) {
	model := newInputModel("> ", nil)
	model.exitUntil = time.Now().Add(-time.Second)

	updated, cmd := model.Update(inputExitTickMsg(time.Now()))
	model = updated.(inputModel)
	if cmd != nil {
		t.Fatal("expired exit confirmation should not keep ticking")
	}
	if !model.exitUntil.IsZero() {
		t.Fatal("expired exit confirmation was not cleared")
	}
}

func TestInputTypingClearsExitConfirmation(t *testing.T) {
	model := newInputModel("> ", nil)
	model.exitUntil = time.Now().Add(inputExitConfirmWindow)

	updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Text: "a", Code: 'a'}))
	model = updated.(inputModel)
	if !model.exitUntil.IsZero() {
		t.Fatal("typing should clear exit confirmation")
	}
}

func ctrlCKeyMsg() tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: 'c', Mod: tea.ModCtrl})
}
