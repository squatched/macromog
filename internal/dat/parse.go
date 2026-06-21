package dat

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ReadMacroSet parses a single mcr*.dat macro set file.
func ReadMacroSet(data []byte) (MacroSet, error) {
	if len(data) != MacroSetFileSize {
		return MacroSet{}, fmt.Errorf("macro set: expected %d bytes, got %d", MacroSetFileSize, len(data))
	}

	magic := binary.LittleEndian.Uint32(data[0:4])
	if magic != MagicVersion {
		return MacroSet{}, fmt.Errorf("macro set: unsupported version %d", magic)
	}

	var set MacroSet
	offset := HeaderSize
	for i := 0; i < MacrosPerSet; i++ {
		m, n := parseMacro(data[offset:])
		if i < 10 {
			set.Ctrl[i] = m
		} else {
			set.Alt[i-10] = m
		}
		offset += n
	}
	return set, nil
}

func parseMacro(data []byte) (Macro, int) {
	var m Macro
	pos := MacroPrefixSize
	for i := 0; i < LineCount; i++ {
		m.Contents[i] = DecodeText(data[pos : pos+LineSize])
		pos += LineSize
	}
	m.Name = DecodeText(data[pos : pos+NameSize])
	return m, MacroSize
}

// ReadBookTitles loads book names from mcr.ttl (books 1-20) and mcr_2.ttl (21-40).
// Missing title files yield empty names for those books.
func ReadBookTitles(dir string) ([MaxBooks]string, error) {
	var titles [MaxBooks]string

	if err := readTitleFile(filepath.Join(dir, "mcr.ttl"), titles[:20]); err != nil && !os.IsNotExist(err) {
		return titles, err
	}
	if err := readTitleFile(filepath.Join(dir, "mcr_2.ttl"), titles[20:]); err != nil && !os.IsNotExist(err) {
		return titles, err
	}
	return titles, nil
}

func readTitleFile(path string, dest []string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(data) < HeaderSize {
		return fmt.Errorf("%s: file too short", filepath.Base(path))
	}
	magic := binary.LittleEndian.Uint32(data[0:4])
	if magic != MagicVersion {
		return fmt.Errorf("%s: unsupported version %d", filepath.Base(path), magic)
	}

	payload := data[HeaderSize:]
	if len(payload)%BookNameSize != 0 {
		return fmt.Errorf("%s: payload size %d is not a multiple of %d", filepath.Base(path), len(payload), BookNameSize)
	}

	count := len(payload) / BookNameSize
	if count > len(dest) {
		count = len(dest)
	}
	for i := 0; i < count; i++ {
		chunk := payload[i*BookNameSize : (i+1)*BookNameSize]
		dest[i] = strings.TrimRight(DecodeText(chunk), "\x00")
	}
	return nil
}

// DiscoverMacroFiles returns sorted absolute paths to mcr*.dat files in dir.
// nmcr*.dat and other variants are ignored.
func DiscoverMacroFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	type indexedPath struct {
		index int
		path  string
	}
	var items []indexedPath
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		index, ok := ParseMacroFileName(e.Name())
		if !ok {
			continue
		}
		items = append(items, indexedPath{index, filepath.Join(dir, e.Name())})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].index < items[j].index })
	paths := make([]string, len(items))
	for i, item := range items {
		paths[i] = item.path
	}
	return paths, nil
}

// ReadMacroSetFile reads and parses a macro set from path.
func ReadMacroSetFile(path string) (MacroSet, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return MacroSet{}, err
	}
	return ReadMacroSet(data)
}