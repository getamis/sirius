// Copyright 2017 AMIS Technologies
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/Sirupsen/logrus"
)

const (
	nocolor = 0
	red     = 31
	green   = 32
	yellow  = 33
	blue    = 34
	cyan    = 36
	gray    = 37

	termMsgBefore = 90
	termMsgJust   = 50

	defaultStackDepth = 5
)

var (
	baseTimestamp time.Time
	isTerminal    bool
)

func init() {
	baseTimestamp = time.Now()
	isTerminal = true
}

type defaultFormatter struct {
	// Set to true to bypass checking for a TTY before outputting colors.
	ForceColors bool

	// Force enabling colors.
	EnableColors bool

	// Disable timestamp logging. useful when output is redirected to logging
	// system that already adds timestamps.
	DisableTimestamp bool
}

func (f *defaultFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := &bytes.Buffer{}

	isColorTerminal := isTerminal && (runtime.GOOS != "windows")
	isColored := f.ForceColors || (f.EnableColors && isColorTerminal)

	var levelColor int
	switch entry.Level {
	case logrus.InfoLevel:
		levelColor = green
	case logrus.DebugLevel:
		levelColor = cyan
	case logrus.WarnLevel:
		levelColor = yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = red
	default:
		levelColor = blue
	}

	var tag, timestamp, caller string
	if isColored {
		tag = f.getColoredString(f.levelTag(entry.Level), levelColor)
		if f.DisableTimestamp {
			timestamp = ""
		} else {
			timestamp = f.getColoredString(f.timeFormat(), gray)
		}

		caller = f.getColoredString(f.caller(defaultStackDepth), gray)
	} else {
		tag = f.levelTag(entry.Level)
		if f.DisableTimestamp {
			timestamp = ""
		} else {
			timestamp = f.timeFormat()
		}

		caller = f.caller(defaultStackDepth)
	}

	f.appendValue(b, tag, false)
	b.WriteString("[")
	f.appendValue(b, timestamp, false)
	f.appendValue(b, caller, false)
	b.WriteString("]")

	if b.Len() < termMsgBefore {
		b.Write(bytes.Repeat([]byte{' '}, termMsgBefore-b.Len()))
	}

	f.appendValue(b, entry.Message, false)

	// try to justify the log output for short messages
	if len(entry.Data) > 0 && len(entry.Message) < termMsgJust {
		b.Write(bytes.Repeat([]byte{' '}, termMsgJust-len(entry.Message)))
	}

	f.writeFields(b, entry.Data)

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func needsQuoting(text string) bool {
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.') {
			return false
		}
	}
	return true
}

func (f *defaultFormatter) writeFields(b *bytes.Buffer, fields logrus.Fields) {
	keys := make([]string, 0)
	for k := range fields {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(fmt.Sprintf("%v ", fields[k]))
	}
}

func (f *defaultFormatter) appendValue(b *bytes.Buffer, value interface{}, space bool) {
	switch value := value.(type) {
	case string:
		if needsQuoting(value) {
			b.WriteString(value)
		} else {
			fmt.Fprintf(b, "%s", value)
		}
	case error:
		errmsg := value.Error()
		if needsQuoting(errmsg) {
			b.WriteString(errmsg)
		} else {
			fmt.Fprintf(b, "%s", value)
		}
	default:
		fmt.Fprint(b, value)
	}

	if space {
		b.WriteByte(' ')
	}
}

func (f *defaultFormatter) timeFormat() string {
	return time.Now().Format("20060102|15:04:05.000")
}

func (f *defaultFormatter) levelTag(level logrus.Level) string {
	switch level {
	case logrus.DebugLevel:
		return "DEBUG"
	case logrus.InfoLevel:
		return "INFO "
	case logrus.WarnLevel:
		return "WARN "
	case logrus.ErrorLevel:
		return "ERROR"
	case logrus.FatalLevel:
		return "CRIT "
	case logrus.PanicLevel:
		return "CRIT "
	}

	return " "
}

func (f *defaultFormatter) caller(depth int) string {
	_, file, line, _ := runtime.Caller(depth)
	file = path.Join(filepath.Base(filepath.Dir(file)), filepath.Base(file))
	return fmt.Sprintf("|%s:%d", file, line)
}

func (f *defaultFormatter) getColoredString(s string, color int) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", color, s)
}
