package gqlutil

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors/errbase"
	"github.com/cockroachdb/errors/withstack"
	pkgErr "github.com/pkg/errors"
)

// getOneLineSource extracts the file/line/function information
// of the topmost caller in the innermost recorded stack trace.
// The filename is simplified to remove the path prefix.
// This is used e.g. to populate the "source" field in
// PostgreSQL errors.
func getOneLineSource(err error) (file string, line int, fn string, ok bool) {
	// We want the innermost entry: start by recursing.
	if c := errbase.UnwrapOnce(err); c != nil {
		if file, line, fn, ok = getOneLineSource(c); ok {
			return
		}
	}
	// If we reach this point, we haven't found anything in the cause so
	// far. Look at the current level.

	// If we have a stack trace in the style of github.com/pkg/errors
	// (either from there or our own withStack), use it.
	if st, ok := err.(interface{ StackTrace() pkgErr.StackTrace }); ok {
		return getOneLineSourceFromPkgStack(st.StackTrace())
	}

	// If we have flattened a github.com/pkg/errors-style stack
	// trace to a string, it will happen in the error's safe details
	// and we need to parse it.
	if sd, ok := err.(errbase.SafeDetailer); ok {
		details := sd.SafeDetails()
		if len(details) > 0 {
			switch errbase.GetTypeKey(err) {
			case pkgFundamental, pkgWithStackName, ourWithStackName:
				return getOneLineSourceFromPrintedStack(details[0])
			}
		}
	}

	// No conversion available - no stack trace.
	return "", 0, "", false
}

func getOneLineSourceFromPkgStack(
	st pkgErr.StackTrace,
) (file string, line int, fn string, ok bool) {
	if len(st) > 0 {
		st = st[:1]
		// Note: the stack trace logic changed between go 1.11 and 1.12.
		// Trying to analyze the frame PCs point-wise will cause
		// the output to change between the go versions.
		stS := fmt.Sprintf("%+v", st)
		return getOneLineSourceFromPrintedStack(stS)
	}
	return "", 0, "", false
}

func getOneLineSourceFromPrintedStack(st string) (file string, line int, fn string, ok bool) {
	// We only need 3 lines: the function/file/line info will be on the first two lines.
	// See parsePrintedStack() for details.
	lines := strings.SplitN(strings.TrimSpace(st), "\n", 3)
	if len(lines) > 0 {
		_, file, line, fnName := parsePrintedStackEntry(lines, 0)
		if fnName != "unknown" {
			_, fn = functionName(fnName)
		}
		return file, line, fn, true
	}
	return "", 0, "", false
}

// functionName is an adapted copy of the same function in package raven-go.
func functionName(fnName string) (pack string, name string) {
	name = fnName
	// We get this:
	//	runtime/debug.*T·ptrmethod
	// and want this:
	//  pack = runtime/debug
	//	name = *T.ptrmethod
	if idx := strings.LastIndex(name, "."); idx != -1 {
		pack = name[:idx]
		name = name[idx+1:]
	}
	name = strings.Replace(name, "·", ".", -1)
	return
}

// parsePrintedStackEntry extracts the stack entry information
// in lines at position i. It returns the new value of i if more than
// one line was read.
func parsePrintedStackEntry(
	lines []string, i int,
) (newI int, file string, line int, fnName string) {
	// The function name is on the first line.
	fnName = lines[i]

	// The file:line pair may be on the line after that.
	if i < len(lines)-1 && strings.HasPrefix(lines[i+1], "\t") {
		fileLine := strings.TrimSpace(lines[i+1])
		// Separate file path and line number.
		lineSep := strings.LastIndexByte(fileLine, ':')
		if lineSep == -1 {
			file = fileLine
		} else {
			file = fileLine[:lineSep]
			lineStr := fileLine[lineSep+1:]
			line, _ = strconv.Atoi(lineStr)
		}
		i++
	}
	return i, file, line, fnName
}

var pkgFundamental errbase.TypeKey
var pkgWithStackName errbase.TypeKey
var ourWithStackName errbase.TypeKey

func init() {
	err := errors.New("")
	pkgFundamental = errbase.GetTypeKey(pkgErr.New(""))
	pkgWithStackName = errbase.GetTypeKey(pkgErr.WithStack(err))
	ourWithStackName = errbase.GetTypeKey(withstack.WithStack(err))
}
