package hyperv

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrInvalidCriticalErrorAction = fmt.Errorf("the provided CriticalErrorAction string is not a valid type")
	ErrInvalidStartAction         = fmt.Errorf("the provided StartAction value is not a valid type")
	ErrInvalidStopAction          = fmt.Errorf("the provided StopAction value is not a valid type")
	ErrInvalidCheckpointType      = fmt.Errorf("the provided CheckpointType value is not a valid type")
	ErrInvalidOnOffState          = fmt.Errorf("the provided OnOffState value is not a valid type")
)

type CriticalErrorAction int

const (
	CriticalErrorActionNone  CriticalErrorAction = 0
	CriticalErrorActionPause CriticalErrorAction = 1
)

const (
	CriticalErrorActionNoneText  = "None"
	CriticalErrorActionPauseText = "Pause"
)

var CriticalErrorActionName = map[CriticalErrorAction]string{
	CriticalErrorActionNone:  CriticalErrorActionNoneText,
	CriticalErrorActionPause: CriticalErrorActionPauseText,
}

var CriticalErrorActionvalue = map[string]CriticalErrorAction{
	strings.ToLower(CriticalErrorActionNoneText):  CriticalErrorActionNone,
	strings.ToLower(CriticalErrorActionPauseText): CriticalErrorActionPause,
}

func (x CriticalErrorAction) String() (string, error) {
	if value, exist := CriticalErrorActionName[x]; exist {
		return value, nil
	}
	return "", ErrInvalidCriticalErrorAction
}

func ToCriticalErrorAction(x string) (CriticalErrorAction, error) {
	if integerValue, err := strconv.Atoi(x); err == nil {
		return CriticalErrorAction(integerValue), nil
	}

	if value, exist := CriticalErrorActionvalue[strings.ToLower(x)]; exist {
		return value, nil
	}

	return CriticalErrorAction(-1), ErrInvalidCriticalErrorAction
}

func (d *CriticalErrorAction) MarshalJSON() ([]byte, error) {
	s, err := d.String()
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(s)
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (d *CriticalErrorAction) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		var i int
		err2 := json.Unmarshal(b, &i)
		if err2 == nil {
			*d = CriticalErrorAction(i)
			return nil
		}

		return err
	}
	*d, err = ToCriticalErrorAction(s)
	return err
}

type StartAction int

const (
	StartActionNothing        StartAction = 2
	StartActionStartIfRunning StartAction = 3
	StartActionStart          StartAction = 4
)

const (
	StartActionNothingText        = "Nothing"
	StartActionStartIfRunningText = "StartIfRunning"
	StartActionStartText          = "Start"
)

var StartActionName = map[StartAction]string{
	StartActionNothing:        StartActionNothingText,
	StartActionStartIfRunning: StartActionStartIfRunningText,
	StartActionStart:          StartActionStartText,
}

var StartActionvalue = map[string]StartAction{
	strings.ToLower(StartActionNothingText):        StartActionNothing,
	strings.ToLower(StartActionStartIfRunningText): StartActionStartIfRunning,
	strings.ToLower(StartActionStartText):          StartActionStart,
}

func (x StartAction) String() (string, error) {
	if value, exist := StartActionName[x]; exist {
		return value, nil
	}

	return "", ErrInvalidStartAction
}

func ToStartAction(x string) (StartAction, error) {
	if integerValue, err := strconv.Atoi(x); err == nil {
		return StartAction(integerValue), nil
	}

	if value, exist := StartActionvalue[strings.ToLower(x)]; exist {
		return value, nil
	}

	return StartAction(-1), ErrInvalidStartAction
}

func (d *StartAction) MarshalJSON() ([]byte, error) {
	s, err := d.String()
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(s)
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (d *StartAction) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		var i int
		err2 := json.Unmarshal(b, &i)
		if err2 == nil {
			*d = StartAction(i)
			return nil
		}

		return err
	}

	*d, err = ToStartAction(s)
	return err
}

type StopAction int

const (
	StopActionTurnOff  StopAction = 2
	StopActionSave     StopAction = 3
	StopActionShutDown StopAction = 4
)

const (
	StopActionTurnOffText  = "TurnOff"
	StopActionSaveText     = "Save"
	StopActionShutDownText = "ShutDown"
)

var StopActionName = map[StopAction]string{
	StopActionTurnOff:  StopActionTurnOffText,
	StopActionSave:     StopActionSaveText,
	StopActionShutDown: StopActionShutDownText,
}

var StopActionvalue = map[string]StopAction{
	strings.ToLower(StopActionTurnOffText):  StopActionTurnOff,
	strings.ToLower(StopActionSaveText):     StopActionSave,
	strings.ToLower(StopActionShutDownText): StopActionShutDown,
}

func (x StopAction) String() (string, error) {
	if value, exist := StopActionName[x]; exist {
		return value, nil
	}

	return "", ErrInvalidStopAction
}

func ToStopAction(x string) (StopAction, error) {
	if integerValue, err := strconv.Atoi(x); err == nil {
		return StopAction(integerValue), nil
	}

	if value, exist := StopActionvalue[strings.ToLower(x)]; exist {
		return value, nil
	}

	return StopAction(-1), ErrInvalidStopAction
}

func (d *StopAction) MarshalJSON() ([]byte, error) {
	s, err := d.String()
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(s)
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (d *StopAction) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		var i int
		err2 := json.Unmarshal(b, &i)
		if err2 == nil {
			*d = StopAction(i)
			return nil
		}

		return err
	}

	*d, err = ToStopAction(s)
	return err
}

type CheckpointType int

const (
	CheckpointTypeDisabled       CheckpointType = 2
	CheckpointTypeProduction     CheckpointType = 3
	CheckpointTypeProductionOnly CheckpointType = 4
	CheckpointTypeStandard       CheckpointType = 5
)

const (
	CheckpointTypeDisabledText       = "Disabled"
	CheckpointTypeProductionText     = "Production"
	CheckpointTypeProductionOnlyText = "ProductionOnly"
	CheckpointTypeStandardText       = "Standard"
)

var CheckpointTypename = map[CheckpointType]string{
	CheckpointTypeDisabled:       CheckpointTypeDisabledText,
	CheckpointTypeProduction:     CheckpointTypeProductionText,
	CheckpointTypeProductionOnly: CheckpointTypeProductionOnlyText,
	CheckpointTypeStandard:       CheckpointTypeStandardText,
}

var CheckpointTypevalue = map[string]CheckpointType{
	strings.ToLower(CheckpointTypeDisabledText):       CheckpointTypeDisabled,
	strings.ToLower(CheckpointTypeProductionText):     CheckpointTypeProduction,
	strings.ToLower(CheckpointTypeProductionOnlyText): CheckpointTypeProductionOnly,
	strings.ToLower(CheckpointTypeStandardText):       CheckpointTypeStandard,
}

func (x CheckpointType) String() (string, error) {
	if value, exist := CheckpointTypename[x]; exist {
		return value, nil
	}

	return "", ErrInvalidCheckpointType
}

func ToCheckpointType(x string) (CheckpointType, error) {
	if integerValue, err := strconv.Atoi(x); err == nil {
		return CheckpointType(integerValue), nil
	}

	if value, exist := CheckpointTypevalue[strings.ToLower(x)]; exist {
		return value, nil
	}

	return CheckpointType(-1), ErrInvalidCheckpointType
}

func (d *CheckpointType) MarshalJSON() ([]byte, error) {
	s, err := d.String()
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(s)
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (d *CheckpointType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		var i int
		err2 := json.Unmarshal(b, &i)
		if err2 == nil {
			*d = CheckpointType(i)
			return nil
		}

		return err
	}

	*d, err = ToCheckpointType(s)
	return err
}

type OnOffState int

const (
	OnOffStateOn  OnOffState = 0
	OnOffStateOff OnOffState = 1
)

const (
	OnOffStateOnText  = "On"
	OnOffStateOffText = "Off"
)

var OnOffStatename = map[OnOffState]string{
	OnOffStateOn:  OnOffStateOnText,
	OnOffStateOff: OnOffStateOffText,
}

var OnOffStatevalue = map[string]OnOffState{
	strings.ToLower(OnOffStateOnText):  OnOffStateOn,
	strings.ToLower(OnOffStateOffText): OnOffStateOff,
}

func (x OnOffState) String() (string, error) {
	if value, exist := OnOffStatename[x]; exist {
		return value, nil
	}

	return "", ErrInvalidOnOffState
}

func ToOnOffState(x string) (OnOffState, error) {
	if integerValue, err := strconv.Atoi(x); err == nil {
		return OnOffState(integerValue), nil
	}

	if value, exist := OnOffStatevalue[strings.ToLower(x)]; exist {
		return value, nil
	}

	return OnOffState(-1), ErrInvalidOnOffState
}

func (d *OnOffState) MarshalJSON() ([]byte, error) {
	s, err := d.String()
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(s)
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (d *OnOffState) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		var i int
		err2 := json.Unmarshal(b, &i)
		if err2 == nil {
			*d = OnOffState(i)
			return nil
		}

		return err
	}
	*d, err = ToOnOffState(s)
	return err
}

type VMExists struct {
	Exists bool
}

type VM struct {
	Name                                string
	Path                                string
	Generation                          int
	AutomaticCriticalErrorAction        CriticalErrorAction
	AutomaticCriticalErrorActionTimeout int32
	AutomaticStartAction                StartAction
	AutomaticStartDelay                 int32
	AutomaticStopAction                 StopAction
	CheckpointType                      CheckpointType
	DynamicMemory                       bool
	GuestControlledCacheTypes           bool
	HighMemoryMappedIoSpace             uint64
	LockOnDisconnect                    OnOffState
	LowMemoryMappedIoSpace              uint32
	MemoryMaximumBytes                  int64
	MemoryMinimumBytes                  int64
	MemoryStartupBytes                  int64
	Notes                               string
	ProcessorCount                      int64
	SmartPagingFilePath                 string
	SnapshotFileLocation                string
	StaticMemory                        bool
	// ParentCheckpointName				string  this will allow us to set the checkpoint to use
}

type HyperVVMClient interface {
	GetVMByID(ctx context.Context, id string) (result VM, err error)
}
