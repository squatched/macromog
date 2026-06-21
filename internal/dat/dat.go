package dat

// Macro set (.dat) file layout derived from POLUtils and confirmed against
// live FFXI client files. Each macro is a C++ struct serialized without
// inter-field padding:
//
//	uint32  prefix (always 0)
//	char[6][61] lines (60 chars + NUL, Shift-JIS with FFXI extensions)
//	char[10] name (8 chars + NUL + padding, Shift-JIS)
//
// A macro set file stores 20 macros (10 ctrl + 10 alt) after a 24-byte header.

const (
	MacroSetFileSize = 7624
	HeaderSize       = 24
	MacrosPerSet     = 20
	MacroPrefixSize  = 4
	LineCount        = 6
	LineSize         = 61
	NameSize         = 10
	MacroSize        = MacroPrefixSize + LineCount*LineSize + NameSize

	MagicVersion = 1
	MaxBooks     = 40
	SetsPerBook  = 10
	BookNameSize = 16
)

// Macro is a single macro button entry from a .dat file.
type Macro struct {
	Name     string
	Contents [LineCount]string
}

// MacroSet is one in-game macro set (ctrl bar + alt bar).
type MacroSet struct {
	Ctrl [10]Macro
	Alt  [10]Macro
}

// Empty reports whether the macro has no name and no line content.
func (m Macro) Empty() bool {
	if m.Name != "" {
		return false
	}
	for _, line := range m.Contents {
		if line != "" {
			return false
		}
	}
	return true
}