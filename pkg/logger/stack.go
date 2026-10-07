package logger

import (
	"fmt"
	"runtime"
	"strings"
)

// DefaultSkipPackages are path fragments of frames that should not be reported as
// the log's "source" (the logger's own frames, runtime, test harness).
var DefaultSkipPackages = []string{
	"McQueens_Tea_Cup/pkg/logger",
	"pkg/logger",
	"runtime/",
	"testing/",
}

// getCaller walks up the stack from skip frames and returns the first "file:line"
// that is not part of the logger package itself.
func getCaller(skip int) string {
	var pcs [10]uintptr
	n := runtime.Callers(skip, pcs[:])

	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()

		shouldSkip := false
		for _, skipPkg := range DefaultSkipPackages {
			if strings.Contains(frame.File, skipPkg) {
				shouldSkip = true
				break
			}
		}
		if !shouldSkip {
			return fmt.Sprintf("%s:%d", relativeSource(frame.File), frame.Line)
		}
		if !more {
			break
		}
	}
	return "unknown:0"
}

// relativeSource trims an absolute compile-time file path down to a module-relative
// one by anchoring on the project's top-level package directories. This keeps the
// "source" field stable and readable regardless of where the binary was built
// (local machine vs. CI/container). Example:
//
//	/Users/me/go/src/mcqueens_tea_cup/internal/domain/service/x.go
//	=> internal/domain/service/x.go
func relativeSource(file string) string {
	for _, root := range []string{"/internal/", "/cmd/", "/pkg/"} {
		if i := strings.Index(file, root); i >= 0 {
			return file[i+1:] // drop the leading slash
		}
	}
	return file
}

// getErrorStack returns a formatted stack trace (if the error carries one via a
// "%+v" formatter) and an optional error code. Standard fmt.Errorf errors have no
// stack, so this degrades to the error message.
func getErrorStack(err error) (string, string) {
	if err == nil {
		return "", ""
	}
	return formatStackTrace(err), ""
}

func formatStackTrace(err error) string {
	stackStr := fmt.Sprintf("%+v", err)

	lines := strings.Split(stackStr, "\n")
	var formattedLines []string
	for i, line := range lines {
		if i == 0 {
			formattedLines = append(formattedLines, "Error: "+line)
		} else if strings.HasPrefix(line, "\t") {
			formattedLines = append(formattedLines, "  "+strings.TrimSpace(line))
		} else if line != "" {
			formattedLines = append(formattedLines, line)
		}
	}
	return strings.Join(formattedLines, "\n")
}
