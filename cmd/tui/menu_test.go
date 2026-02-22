package main

import "testing"

func TestResolveMainMenuChoice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		isAdmin bool
		choice  string
		ok      bool
	}{
		{name: "genres allowed", isAdmin: false, choice: "1", ok: true},
		{name: "logout allowed", isAdmin: false, choice: "0", ok: true},
		{name: "admin users denied for non admin", isAdmin: false, choice: "10", ok: false},
		{name: "admin users allowed for admin", isAdmin: true, choice: "10", ok: true},
		{name: "unknown denied", isAdmin: true, choice: "999", ok: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			app := &TUIApp{isAdmin: tt.isAdmin}
			_, ok := app.resolveMainMenuChoice(tt.choice)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
		})
	}
}

func TestMainMenuActionsContainsExpectedKeys(t *testing.T) {
	t.Parallel()

	app := &TUIApp{}
	actions := app.mainMenuActions()
	for _, key := range []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10"} {
		if actions[key] == nil {
			t.Fatalf("expected action for key %s", key)
		}
	}
}
