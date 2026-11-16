package app

import (
	"fmt"

	"girc/ui"
)

// formatChat formats a chat message. Own messages use a different color.
func formatChat(nick, text string, own bool) string {
	color := "green"
	if own {
		color = "aqua"
	}
	return fmt.Sprintf("[%s]<%s>[-] %s", color, ui.Escape(nick), ui.Escape(text))
}

// formatAction formats a CTCP ACTION (/me).
func formatAction(nick, text string) string {
	return fmt.Sprintf("[fuchsia]* %s[-] %s", ui.Escape(nick), ui.Escape(text))
}

// printInfo prints an informational line. text is escaped.
func (a *App) printInfo(buffer, text string) {
	a.view.Println(buffer, "[gray]-!-[-] "+ui.Escape(text))
}

// printError prints an error line.
func (a *App) printError(buffer string, err error) {
	a.view.Println(buffer, "[red]error:[-] "+ui.Escape(err.Error()))
}

func reason(r string) string {
	if r == "" {
		return ""
	}
	return " (" + r + ")"
}
