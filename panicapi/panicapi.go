package panicapi

import (
	"fmt"
	"strconv"
	"strings"
)

/*
SCSubCodeZero and these constants are used as unique subcodes in error responses
*/
const (
	SCSubCodeZero = iota
	SCPathNotFound
	SCFileNotFound
	SCStaticPathNotFound
	SCContentNotFound
	SCContentReadFailed
	SCServerShutDown
	SCInvalidJSONRequest
	SCReadJSONRequest
	SCJSONResponseErr
	SCMissingURLParam
	SCStaticFileInit
	SCTemplateNotFound
	SCTemplateError
	SCRuntimeError
	SCStaticPath
	SCWriteFile
	SCParamValidation
	SCScriptNotFound
	SCScriptError
	SCOpenFileError
	SCUnhandledPanic
	SCMax
)

/*
The status data for a panic
*/
type PanicState struct {
	IsPanicData bool
	Severity    string
	StatusCode  int
	SubCode     int
	ErrorText   string
	LogMessage  string
	TxID        string
	original    string
}

/*
ThrowWarning throw a panic as a special case. It is recovered and logged as warning
*/
func ThrowWarning(statusCode, subCode int, errorText string, logMessage string) {
	panic(fmt.Sprintf("W|%d|%d|%s|%s", statusCode, subCode, errorText, logMessage))
}

/*
ThrowInfo throw a panic as a special case. It is recovered and logged as info
*/
func ThrowInfo(statusCode, subCode int, errorText string, logMessage string) {
	panic(fmt.Sprintf("I|%d|%d|%s|%s", statusCode, subCode, errorText, logMessage))
}

/*
ThrowError throw a panic as a special case. It is recovered and logged as an error
*/
func ThrowError(statusCode, subCode int, errorText string, logMessage string) {
	panic(fmt.Sprintf("E|%d|%d|%s|%s", statusCode, subCode, errorText, logMessage))
}

/*
GetPanicData converts a formatted panic string in to a PanicState struce
*/
func GetPanicData(panic interface{}, txid string) *PanicState {
	panicString := fmt.Sprintf("%s", panic)
	parts := strings.Split(panicString, "|")
	if len(parts) == 1 && parts[0] != "I" && parts[0] != "W" && parts[0] != "E" {
		/*
			Cannot understand the erroe text!
		*/
		return &PanicState{
			TxID:        txid,
			IsPanicData: false,
			Severity:    "E",
			StatusCode:  500,
			SubCode:     SCUnhandledPanic,
			ErrorText:   "Unhandled Panic",
			LogMessage:  panicString,
		}
	}
	/*
		Convert the error text to a PanicState
	*/
	return &PanicState{
		TxID:        txid,
		IsPanicData: true,
		Severity:    getStringValue(parts, 0),
		StatusCode:  getIntValue(parts, 1),
		SubCode:     getIntValue(parts, 2),
		ErrorText:   getStringValue(parts, 3),
		LogMessage:  getStringValue(parts, 4),
	}
}

func (p *PanicState) String() string {
	pd := "UNHANDLED"
	if p.IsPanicData {
		pd = "PANIC"
	}
	return fmt.Sprintf("%s: %s: %d.%d: %s %s", pd, p.Severity, p.StatusCode, p.SubCode, p.ErrorText, p.LogMessage)
}

func (p *PanicState) Error() error {
	return fmt.Errorf(p.String())
}

func getStringValue(parts []string, index int) string {
	if index >= len(parts) {
		return ""
	}
	return parts[index]
}

func getIntValue(parts []string, index int) int {
	if index >= len(parts) {
		return -1
	}
	i, err := strconv.Atoi(parts[index])
	if err != nil {
		return -2
	}
	return i
}
