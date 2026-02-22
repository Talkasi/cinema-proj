package validation

import "testing"

func TestValidateDTORequired(t *testing.T) {
	t.Parallel()

	type req struct {
		Name string `validate:"required"`
	}

	err := ValidateDTO(req{})
	if err == nil {
		t.Fatal("expected error")
	}

	verr, ok := err.(*Error)
	if !ok {
		t.Fatalf("unexpected error type: %T", err)
	}

	if verr.Msg != "Invalid input" {
		t.Fatalf("msg = %q", verr.Msg)
	}
	if len(verr.Fields) != 1 {
		t.Fatalf("fields len = %d, want 1", len(verr.Fields))
	}
	if verr.Fields[0].Field != "name" {
		t.Fatalf("field = %q, want name", verr.Fields[0].Field)
	}
	if verr.Fields[0].Rule != "required" {
		t.Fatalf("rule = %q, want required", verr.Fields[0].Rule)
	}
}

func TestValidateDTOEmail(t *testing.T) {
	t.Parallel()

	type req struct {
		Email string `validate:"required,email"`
	}

	err := ValidateDTO(req{Email: "not-an-email"})
	if err == nil {
		t.Fatal("expected error")
	}

	verr, ok := err.(*Error)
	if !ok {
		t.Fatalf("unexpected error type: %T", err)
	}

	if len(verr.Fields) == 0 {
		t.Fatal("expected field errors")
	}
	if verr.Fields[0].Rule != "email" {
		t.Fatalf("rule = %q, want email", verr.Fields[0].Rule)
	}
}

func TestValidateDTOValid(t *testing.T) {
	t.Parallel()

	type req struct {
		Name  string `validate:"required,min=2"`
		Email string `validate:"required,email"`
	}

	err := ValidateDTO(req{Name: "Ivan", Email: "ivan@example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
