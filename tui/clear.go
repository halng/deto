package tui

import (
	"fmt"
)

// Clear only clear from pointer
func Clear() {
	fmt.Print("\033[H\033[2J")
}
