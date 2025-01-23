package mode

import "errors"

// ErrInvalidMode Err Invalid Hex
var ErrInvalidMode = errors.New("the provided mode string is not a valid Mode")

const (
	ControllerModeText string = "controller"
	NodeModeText       string = "node"
	AllModeText        string = "all"
)

const (
	// ControllerMode is the mode that only starts the controller service.
	ControllerMode Mode = Mode(ControllerModeText)
	// NodeMode is the mode that only starts the node service.
	NodeMode Mode = Mode(NodeModeText)
	// AllMode is the mode that only starts both the controller and the node service.
	AllMode Mode = Mode(AllModeText)
	// NilMode
	NilMode Mode = ""
)

// Mode is the operating mode of the CSI driver.
type Mode string

// StringToMode String To ObjectID
func StringToMode(mode string) (Mode, error) {
	switch mode {
	case ControllerModeText:
		return ControllerMode, nil
	case NodeModeText:
		return NodeMode, nil
	case AllModeText:
		return AllMode, nil
	default:
		return NilMode, ErrInvalidMode
	}
}

// String String
func (mode Mode) String() string {
	return string(mode)
}
