package cli

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

type selectableItem struct {
	Label       string
	Description string
	Selected    bool
}

// promptMultiSelect shows an interactive multi-select prompt in the terminal.
// All items start as selected. Returns indices of selected items.
// Returns error on Ctrl+C or terminal issues.
// Falls back to returning all items if stdin is not a terminal.
func promptMultiSelect(items []selectableItem) ([]int, error) {
	if len(items) == 0 {
		return nil, nil
	}

	fd := int(os.Stdin.Fd())

	// Non-interactive fallback: return all items
	if !term.IsTerminal(fd) {
		return allIndices(len(items)), nil
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, fmt.Errorf("failed to set raw terminal: %w", err)
	}
	defer func() { _ = term.Restore(fd, oldState) }()

	return runMultiSelect(items, os.Stdin, os.Stdout)
}

// allIndices returns [0, 1, ..., n-1].
func allIndices(n int) []int {
	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}
	return indices
}

// runMultiSelect is the core interactive multi-select loop, separated from
// terminal setup so it can be tested with piped io.Reader/io.Writer.
func runMultiSelect(items []selectableItem, r io.Reader, w io.Writer) ([]int, error) {
	cursor := 0
	buf := make([]byte, 3)
	totalLines := len(items) + 2 // header + blank line + items

	write := func(s string) {
		_, _ = io.WriteString(w, s)
	}

	render := func(first bool) {
		if !first {
			write(fmt.Sprintf("\033[%dA", totalLines))
		}
		write("\033[J") // clear to end of screen
		write("  Use \xe2\x86\x91\xe2\x86\x93 to navigate, Space to toggle, 'a' to select/deselect all, Enter to confirm\r\n")
		write("\r\n")
		for i, item := range items {
			check := " "
			if item.Selected {
				check = "x"
			}
			prefix := "  "
			if i == cursor {
				prefix = "> "
			}
			desc := item.Description
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}
			write(fmt.Sprintf("%s[%s] %s - %s\r\n", prefix, check, item.Label, desc))
		}
	}

	render(true)

	for {
		n, err := r.Read(buf)
		if err != nil {
			return nil, err
		}

		switch {
		case n == 1 && buf[0] == '\r': // Enter
			write("\r\n")
			var selected []int
			for i, item := range items {
				if item.Selected {
					selected = append(selected, i)
				}
			}
			return selected, nil

		case n == 1 && buf[0] == 3: // Ctrl+C
			write("\r\n")
			return nil, fmt.Errorf("installation canceled")

		case n == 1 && buf[0] == ' ': // Space - toggle current
			items[cursor].Selected = !items[cursor].Selected

		case n == 1 && (buf[0] == 'a' || buf[0] == 'A'): // Toggle all
			allSelected := true
			for _, item := range items {
				if !item.Selected {
					allSelected = false
					break
				}
			}
			for i := range items {
				items[i].Selected = !allSelected
			}

		case n == 3 && buf[0] == 27 && buf[1] == 91 && buf[2] == 65: // Up arrow
			if cursor > 0 {
				cursor--
			}

		case n == 3 && buf[0] == 27 && buf[1] == 91 && buf[2] == 66: // Down arrow
			if cursor < len(items)-1 {
				cursor++
			}
		}

		render(false)
	}
}
