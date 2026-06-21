package validate_test

import (
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/validate"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantValid   bool
		errContains []string // substrings that must appear in at least one error
	}{
		// ── Valid cases ────────────────────────────────────────────────────────
		{
			name:      "minimal valid file",
			input:     "version: 1\nbooks: {}\n",
			wantValid: true,
		},
		{
			name: "full example from spec",
			input: `version: 1
character: "Hendrimod"
exported_at: "2026-06-20T03:30:00Z"
books:
  1:
    name: "WHM75"
    sets:
      1:
        ctrl:
          1:
            name: "Cure"
            contents:
              - /ma "Cure IV" <me>
              - /wait 1
          2:
            name: "Esuna"
            contents:
              - /ma "Esuna" <me>
        alt:
          1:
            name: "Protect"
            contents:
              - /ma "Protectra V" <me>
  6:
    name: "RDM75NIN"
    sets:
      1:
        ctrl:
          1:
            name: "WS"
            contents:
              - /ws "Savage Blade" <t>
`,
			wantValid: true,
		},
		{
			name:      "max book index 40",
			input:     "version: 1\nbooks:\n  40:\n    name: \"LastBook\"\n",
			wantValid: true,
		},
		{
			name:      "min book index 1",
			input:     "version: 1\nbooks:\n  1:\n    name: \"First\"\n",
			wantValid: true,
		},
		{
			name:      "book name at max length 15 chars",
			input:     "version: 1\nbooks:\n  1:\n    name: \"ABCDE12345ABCDE\"\n",
			wantValid: true,
		},
		{
			name:      "book with no name field",
			input:     "version: 1\nbooks:\n  1:\n    sets: {}\n",
			wantValid: true,
		},
		{
			name:      "book with empty name",
			input:     "version: 1\nbooks:\n  1:\n    name: \"\"\n",
			wantValid: true,
		},
		{
			name:      "max set index 10",
			input:     "version: 1\nbooks:\n  1:\n    sets:\n      10:\n        ctrl: {}\n",
			wantValid: true,
		},
		{
			name:      "min set index 1",
			input:     "version: 1\nbooks:\n  1:\n    sets:\n      1:\n        ctrl: {}\n",
			wantValid: true,
		},
		{
			name:      "macro key 0 is valid",
			input:     "version: 1\nbooks:\n  1:\n    sets:\n      1:\n        ctrl:\n          0:\n            name: \"Zero\"\n",
			wantValid: true,
		},
		{
			name:      "macro name at max length 8 chars",
			input:     "version: 1\nbooks:\n  1:\n    sets:\n      1:\n        ctrl:\n          1:\n            name: \"12345678\"\n",
			wantValid: true,
		},
		{
			name: "macro with exactly 6 content lines",
			input: `version: 1
books:
  1:
    sets:
      1:
        ctrl:
          1:
            name: "Full"
            contents:
              - /line1
              - /line2
              - /line3
              - /line4
              - /line5
              - /line6
`,
			wantValid: true,
		},
		{
			name:      "content line at exactly 60 chars",
			input:     "version: 1\nbooks:\n  1:\n    sets:\n      1:\n        ctrl:\n          1:\n            contents:\n              - " + strings.Repeat("a", 60) + "\n",
			wantValid: true,
		},
		{
			name:      "exported_at with UTC Z suffix",
			input:     "version: 1\nbooks: {}\nexported_at: \"2026-06-20T03:30:00Z\"\n",
			wantValid: true,
		},
		{
			name:      "exported_at with timezone offset",
			input:     "version: 1\nbooks: {}\nexported_at: \"2026-06-20T12:00:00+09:00\"\n",
			wantValid: true,
		},
		{
			name: "all macro keys 0 through 9 are valid",
			input: `version: 1
books:
  1:
    sets:
      1:
        ctrl:
          0:
            name: "Key0"
          1:
            name: "Key1"
          9:
            name: "Key9"
`,
			wantValid: true,
		},
		{
			name:      "empty books map is valid sparse format",
			input:     "version: 1\nbooks: {}\n",
			wantValid: true,
		},
		{
			name:      "macro with no contents field",
			input:     "version: 1\nbooks:\n  1:\n    sets:\n      1:\n        ctrl:\n          1:\n            name: \"Empty\"\n",
			wantValid: true,
		},
		{
			name:      "alt row validated same as ctrl",
			input:     "version: 1\nbooks:\n  1:\n    sets:\n      1:\n        alt:\n          1:\n            name: \"AltKey\"\n",
			wantValid: true,
		},
		{
			name:      "optional character field accepted",
			input:     "version: 1\ncharacter: \"Kupomog\"\nbooks: {}\n",
			wantValid: true,
		},

		// ── Invalid cases ──────────────────────────────────────────────────────
		{
			name:        "invalid YAML syntax",
			input:       "version: 1\nbooks: {\n",
			wantValid:   false,
			errContains: []string{"invalid YAML"},
		},
		{
			name:        "missing version field",
			input:       "books: {}\n",
			wantValid:   false,
			errContains: []string{"version", "required"},
		},
		{
			name:        "missing books field",
			input:       "version: 1\n",
			wantValid:   false,
			errContains: []string{"books", "required"},
		},
		{
			name:        "both required fields missing",
			input:       "character: \"Kupomog\"\n",
			wantValid:   false,
			errContains: []string{"version", "books"},
		},
		{
			name:        "version 0 is rejected",
			input:       "version: 0\nbooks: {}\n",
			wantValid:   false,
			errContains: []string{"version", "must be 1"},
		},
		{
			name:        "version 2 is rejected",
			input:       "version: 2\nbooks: {}\n",
			wantValid:   false,
			errContains: []string{"version", "must be 1"},
		},
		{
			name:        "version negative is rejected",
			input:       "version: -1\nbooks: {}\n",
			wantValid:   false,
			errContains: []string{"version", "must be 1"},
		},
		{
			name:        "book index 0 is out of range",
			input:       "version: 1\nbooks:\n  0:\n    name: \"Bad\"\n",
			wantValid:   false,
			errContains: []string{"books.0", "1–40"},
		},
		{
			name:        "book index 41 is out of range",
			input:       "version: 1\nbooks:\n  41:\n    name: \"Bad\"\n",
			wantValid:   false,
			errContains: []string{"books.41", "1–40"},
		},
		{
			name:        "book name 16 chars exceeds max",
			input:       "version: 1\nbooks:\n  1:\n    name: \"ABCDE12345ABCDEF\"\n",
			wantValid:   false,
			errContains: []string{"books.1.name", "max 15 characters", "ABCDE12345ABCDEF"},
		},
		{
			name:        "book name with space is rejected",
			input:       "version: 1\nbooks:\n  1:\n    name: \"WHM 75\"\n",
			wantValid:   false,
			errContains: []string{"books.1.name", "alphanumeric", "WHM 75"},
		},
		{
			name:        "book name with hyphen is rejected",
			input:       "version: 1\nbooks:\n  1:\n    name: \"WHM-75\"\n",
			wantValid:   false,
			errContains: []string{"books.1.name", "alphanumeric", "WHM-75"},
		},
		{
			name:        "book name with underscore is rejected",
			input:       "version: 1\nbooks:\n  1:\n    name: \"WHM_75\"\n",
			wantValid:   false,
			errContains: []string{"books.1.name", "alphanumeric", "WHM_75"},
		},
		{
			name:        "set index 0 is out of range",
			input:       "version: 1\nbooks:\n  1:\n    sets:\n      0:\n        ctrl: {}\n",
			wantValid:   false,
			errContains: []string{"books.1.sets.0", "1–10"},
		},
		{
			name:        "set index 11 is out of range",
			input:       "version: 1\nbooks:\n  1:\n    sets:\n      11:\n        ctrl: {}\n",
			wantValid:   false,
			errContains: []string{"books.1.sets.11", "1–10"},
		},
		{
			name:        "macro row key 10 is out of range",
			input:       "version: 1\nbooks:\n  1:\n    sets:\n      1:\n        ctrl:\n          10:\n            name: \"Bad\"\n",
			wantValid:   false,
			errContains: []string{"key must be 0–9"},
		},
		{
			name:        "macro name 9 chars exceeds max",
			input:       "version: 1\nbooks:\n  1:\n    sets:\n      1:\n        ctrl:\n          1:\n            name: \"123456789\"\n",
			wantValid:   false,
			errContains: []string{"books.1.sets.1.ctrl.1.name", "max 8 characters", "123456789"},
		},
		{
			name: "macro with 7 content lines exceeds max",
			input: `version: 1
books:
  1:
    sets:
      1:
        ctrl:
          1:
            contents:
              - /line1
              - /line2
              - /line3
              - /line4
              - /line5
              - /line6
              - /line7
`,
			wantValid:   false,
			errContains: []string{"contents", "max 6"},
		},
		{
			name:        "content line 61 chars exceeds max",
			input:       "version: 1\nbooks:\n  1:\n    sets:\n      1:\n        ctrl:\n          1:\n            contents:\n              - " + strings.Repeat("a", 61) + "\n",
			wantValid:   false,
			errContains: []string{"contents[0]", "max 60 characters"},
		},
		{
			name:        "exported_at date-only string is invalid",
			input:       "version: 1\nbooks: {}\nexported_at: \"2026-06-20\"\n",
			wantValid:   false,
			errContains: []string{"exported_at", "RFC 3339"},
		},
		{
			name:        "exported_at free-form string is invalid",
			input:       "version: 1\nbooks: {}\nexported_at: \"not a date\"\n",
			wantValid:   false,
			errContains: []string{"exported_at", "RFC 3339"},
		},
		{
			name:        "empty input reports both required fields missing",
			input:       "",
			wantValid:   false,
			errContains: []string{"version", "books"},
		},
		{
			name: "multiple independent errors all reported",
			input: `version: 2
books:
  0:
    name: "ABCDE12345ABCDEF"
`,
			wantValid:   false,
			errContains: []string{"version", "books.0", "books.0.name"},
		},
		{
			name: "alt row key out of range reported correctly",
			input: `version: 1
books:
  1:
    sets:
      1:
        alt:
          10:
            name: "Bad"
`,
			wantValid:   false,
			errContains: []string{"alt", "key must be 0–9"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validate.Validate([]byte(tt.input))
			isValid := len(errs) == 0

			if tt.wantValid && !isValid {
				t.Errorf("expected valid, got errors: %v", errs)
				return
			}
			if !tt.wantValid && isValid {
				t.Error("expected validation errors, got none")
				return
			}

			for _, want := range tt.errContains {
				found := false
				for _, e := range errs {
					if strings.Contains(e.Error(), want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected an error containing %q\ngot errors: %v", want, errs)
				}
			}
		})
	}
}

func TestErrorString(t *testing.T) {
	tests := []struct {
		e    validate.Error
		want string
	}{
		{validate.Error{Path: "books.1.name", Message: "max 15 characters, got 20"}, "books.1.name: max 15 characters, got 20"},
		{validate.Error{Path: "", Message: "invalid YAML: unexpected EOF"}, "invalid YAML: unexpected EOF"},
	}
	for _, tt := range tests {
		if got := tt.e.Error(); got != tt.want {
			t.Errorf("Error.Error() = %q, want %q", got, tt.want)
		}
	}
}
