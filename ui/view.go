package ui

import "fmt"

func (m model) View() string {
	// The header
	s := "What should we buy at the market?\n\n"

	// Iterate over our choices
	for i, choice := range m.torrents {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = "▶"
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}

		// Render the row
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice.Name)
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}

func ratePretty(rate uint32) string {
	floatRate := float64(rate)
	suffix := "B/s"
	if floatRate >= 1024 {
		floatRate /= 1024
		suffix = "KiB/s"
	}
	if floatRate >= 1024 {
		floatRate /= 1024
		suffix = "MiB/s"
	}
	return fmt.Sprintf("%.2f "+suffix, floatRate)
}

func sizePretty(size uint32) string {
	floatSize := float64(size)
	suffix := "B"
	if floatSize >= 1024 {
		floatSize /= 1024
		suffix = "KiB"
	}
	if floatSize >= 1024 {
		floatSize /= 1024
		suffix = "MiB"
	}
	if floatSize >= 1024 {
		floatSize /= 1024
		suffix = "GiB"
	}
	if floatSize >= 1024 {
		floatSize /= 1024
		suffix = "TiB"
	}
	return fmt.Sprintf("%.1f "+suffix, floatSize)
}
