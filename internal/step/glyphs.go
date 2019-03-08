package step

import (
	"github.com/fatih/color"
)

var (
	successGlyph  = color.GreenString("✓ ")
	failureGlyph  = color.RedString("✕ ")
	warningGlyph  = color.YellowString("! ")
	questionGlyph = "? "
	infoGlyph     = "  "
)
