package colors

import "fmt"

const (
	BlackOnWhite     = "\x1b[3;1m"
	Underline        = "\x1b[4;1m"
	White            = "\x1b[29;1m"
	Grey             = "\x1b[30;1m"
	Red              = "\x1b[31;1m"
	Green            = "\x1b[32;1m"
	Yellow           = "\x1b[33;1m"
	Blue             = "\x1b[34;1m"
	Pink             = "\x1b[35;1m"
	Azure            = "\x1b[36;1m"
	GreyBackground   = "\x1b[40;1m"
	RedBackground    = "\x1b[41;1m"
	GreenBackground  = "\x1b[42;1m"
	YellowBackground = "\x1b[43;1m"
	BlueBackground   = "\x1b[44;1m"
	PinkBackground   = "\x1b[45;1m"
	AzureBackground  = "\x1b[46;1m"
	WhiteBackground  = "\x1b[47;1m"

	Normal = "\x1b[0m"
)

func Color(s string, c string) string {
	return fmt.Sprintf("%s%s%s", c, s, Normal)
}
