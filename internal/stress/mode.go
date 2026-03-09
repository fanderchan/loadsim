package stress

import "fmt"

type Mode string

const (
	ModeFixed Mode = "fixed"
	ModeWave  Mode = "wave"
)

func ParseMode(value string) (Mode, error) {
	switch Mode(value) {
	case ModeFixed:
		return ModeFixed, nil
	case ModeWave:
		return ModeWave, nil
	default:
		return "", fmt.Errorf("invalid mode %q, must be fixed or wave", value)
	}
}
