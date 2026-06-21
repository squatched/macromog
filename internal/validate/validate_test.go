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
character: "squatched"
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

		// ── Japanese client sanity checks ──────────────────────────────────────
		// The underlying DAT format uses Shift-JIS (each CJK char = 2 bytes), so
		// the 8-character macro name limit is effectively 4 CJK chars in-game.
		// Our validator counts Unicode code points, which is permissive for CJK;
		// the import step enforces the byte-level Shift-JIS constraint.
		{
			name: "JP character name in character field is valid",
			input: `version: 1
character: "ヘンドリモード"
books: {}
`,
			wantValid: true,
		},
		{
			// 攻撃開始 = 4 kanji = 4 runes = 8 Shift-JIS bytes — exactly at the
			// game's byte limit and confirmed in screenshots (BaseATK set 1, Ctrl1).
			name: "4-kanji macro name at Shift-JIS byte boundary is valid",
			input: `version: 1
books:
  1:
    sets:
      1:
        ctrl:
          1:
            name: "攻撃開始"
`,
			wantValid: true,
		},
		{
			// ProShel = 7 ASCII = 7 bytes — confirmed in screenshots (BaseMAG set 2,
			// Alt0). Tests near-limit ASCII macro name.
			name: "7-char ASCII macro name near limit is valid",
			input: `version: 1
books:
  1:
    sets:
      2:
        alt:
          0:
            name: "ProShel"
            contents:
              - /ma "Protectra V" <me>
`,
			wantValid: true,
		},
		{
			// From screenshots: 装備FLM (image 2) — mixed kanji+ASCII macro name.
			// Auto-translate characters seen in the screenshot are omitted; their
			// encoding in the DAT format is unknown and out of scope for YAML validation.
			name: "real-world: mixed kanji+ASCII macro name with equipset and echo",
			input: `version: 1
books:
  1:
    sets:
      1:
        ctrl:
          4:
            name: "装備FLM"
            contents:
              - /equipset 67
              - /echo フラマ装備
`,
			wantValid: true,
		},
		{
			// From screenshots: ニビル (image 6) — equipment swap macro.
			// Auto-translate item name brackets omitted (encoding unknown).
			name: "real-world: equipment swap macro with JP item name",
			input: `version: 1
books:
  1:
    sets:
      1:
        ctrl:
          6:
            name: "ニビル"
            contents:
              - /equip main ニビルブレード 1 <wait 1>
              - /equip sub ニビルブレード 2
`,
			wantValid: true,
		},
		{
			// From screenshots: 範囲舞手 (image 5) — 4-kanji name at the Shift-JIS
			// byte boundary, all 6 lines used, JP spell names. The most pathological
			// real-world example from the article. Auto-translate brackets omitted.
			name: "real-world: 4-kanji name at byte boundary, all 6 lines, JP spells",
			input: `version: 1
books:
  1:
    name: "BaseMAG"
    sets:
      1:
        alt:
          7:
            name: "範囲舞手"
            contents:
              - /recast ディフュージョン
              - /equipset 50
              - /ja ディフュージョン <me> <stal> <wait 3>
              - /ja ノートリアスナレッジ <me> <wait 3>
              - /equip feet LLチャルク+2
              - /ma マイティガード <me>
`,
			wantValid: true,
		},
		{
			// From screenshots: BaseATK/BaseMAG palette (images 3-4) — a realistic
			// multi-set book with JP names, ASCII names, and mixed names.
			name: "real-world: multi-set book matching BaseATK/BaseMAG screenshots",
			input: `version: 1
books:
  1:
    name: "BaseATK"
    sets:
      1:
        ctrl:
          1:
            name: "空蝉1"
          2:
            name: "空蝉2"
          3:
            name: "命中上昇"
          4:
            name: "ディフェ"
          5:
            name: "雄叫"
          6:
            name: "戦狂"
          7:
            name: "挑発"
          8:
            name: "攻撃開始"
          9:
            name: "遠隔"
          0:
            name: "戦狂雄叫"
        alt:
          1:
            name: "空蝉1"
          2:
            name: "攻撃開始"
      2:
        ctrl:
          1:
            name: "Hワルツ"
          2:
            name: "Hサンバ"
          9:
            name: "攻撃開始"
          0:
            name: "アシスト"
`,
			wantValid: true,
		},
		{
			name: "JP macro name over 8-character limit is rejected",
			// ケアルケアルケアル = 9 katakana = 9 runes — exceeds our Unicode limit
			input: `version: 1
books:
  1:
    sets:
      1:
        ctrl:
          1:
            name: "ケアルケアルケアル"
`,
			wantValid:   false,
			errContains: []string{"ctrl.1.name", "max 8 characters"},
		},
		{
			name: "JP characters in book name are rejected",
			// book names are alphanumeric ASCII only — confirmed by screenshots
			input: `version: 1
books:
  1:
    name: "白魔道士"
`,
			wantValid:   false,
			errContains: []string{"books.1.name", "alphanumeric", "白魔道士"},
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
			errContains: []string{"exported_at", "2026-06-20T03:30:00Z"},
		},
		{
			name:        "exported_at free-form string is invalid",
			input:       "version: 1\nbooks: {}\nexported_at: \"not a date\"\n",
			wantValid:   false,
			errContains: []string{"exported_at", "2026-06-20T03:30:00Z"},
		},
		{
			// yaml.v3 itself rejects most control characters before our validation runs
			name:        "control character in macro name rejected as invalid YAML",
			input:       "version: 1\nbooks:\n  1:\n    sets:\n      1:\n        ctrl:\n          1:\n            name: \"Cure\x01\"\n",
			wantValid:   false,
			errContains: []string{"invalid YAML"},
		},
		{
			name:        "control character in content line rejected as invalid YAML",
			input:       "version: 1\nbooks:\n  1:\n    sets:\n      1:\n        ctrl:\n          1:\n            contents:\n              - /ma \"Cure\"\x00\n",
			wantValid:   false,
			errContains: []string{"invalid YAML"},
		},
		{
			// YAML \t escape in a double-quoted scalar produces a literal tab that
			// yaml.v3 allows but the game client would not render meaningfully.
			name: "tab in macro name is rejected",
			input: `version: 1
books:
  1:
    sets:
      1:
        ctrl:
          1:
            name: "Cure\t"
`,
			wantValid:   false,
			errContains: []string{"ctrl.1.name", "printable"},
		},
		{
			name: "tab in content line is rejected",
			input: `version: 1
books:
  1:
    sets:
      1:
        ctrl:
          1:
            contents:
              - "/ma \"Cure\"\t<me>"
`,
			wantValid:   false,
			errContains: []string{"contents[0]", "printable"},
		},
		{
			// YAML \n in a double-quoted scalar embeds a literal newline into the string.
			name: "newline in content line is rejected",
			input: `version: 1
books:
  1:
    sets:
      1:
        ctrl:
          1:
            contents:
              - "/ma \"Cure\"\n<me>"
`,
			wantValid:   false,
			errContains: []string{"contents[0]", "printable"},
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
