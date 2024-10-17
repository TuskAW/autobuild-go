package colors

import "fmt"

const (
	Reset     = "\033[0m"  // Reset the color
	Red       = "\033[31m" // Red
	Green     = "\033[32m" // Green
	Yellow    = "\033[33m" // Yellow
	Blue      = "\033[34m" // Blue
	Purple    = "\033[35m" // Purple
	Cyan      = "\033[36m" // Cyan
	White     = "\033[37m" // White
	Bold      = "\033[1m"  // Bold text
	Underline = "\033[4m"  // Underline text
)

func InfoLog(strfmt string, args ...interface{}) {
	fmt.Printf("    "+Green+"II\t"+Reset+strfmt+"\n", args...)
}

func ErrLog(strfmt string, args ...interface{}) {
	fmt.Printf("    "+Red+"\u274c\t"+Reset+strfmt+"\n", args...)
}

func WarnLog(strfmt string, args ...interface{}) {
	fmt.Printf("    "+Yellow+"WA\t"+Reset+strfmt+"\n", args...)
}

func Success(strfmt string, args ...interface{}) {
	fmt.Printf("    "+Green+"\u2714\t"+Reset+strfmt+"\n", args...)
}

func Icon(color string, icon string, strfmt string, args ...interface{}) {
	fmt.Printf(fmt.Sprintf("    %s%s\t%s\n", color, icon, Reset+strfmt), args...)
}

func HorizontalLine(strfmt string) {
	fmt.Printf(Yellow + "============= " + strfmt + " =============\n" + Reset)
}
