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
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
)

func newLogger(ctx ...interface{}) Logger {
	l := &logger{
		logger:   log.New(),
		contexts: make(map[string]interface{}),
	}

	l.logger.Formatter = &defaultFormatter{
		EnableColors: true,
	}
	l.logger.Out = os.Stdout
	l.logger.SetLevel(log.DebugLevel)

	return l.New(ctx...)
}

// ----------------------------------------------------------------------------

type logger struct {
	logger   *log.Logger
	contexts map[string]interface{}
}

func (l *logger) New(ctx ...interface{}) Logger {
	_ = l.withFields(ctx...)
	return l
}

func (l *logger) Debug(msg string, ctx ...interface{}) {
	l.withFields(ctx...).Debug(msg)
}

func (l *logger) Info(msg string, ctx ...interface{}) {
	l.withFields(ctx...).Info(msg)
}

func (l *logger) Warn(msg string, ctx ...interface{}) {
	l.withFields(ctx...).Warn(msg)
}

func (l *logger) Error(msg string, ctx ...interface{}) {
	l.withFields(ctx...).Error(msg)
}

func (l *logger) Crit(msg string, ctx ...interface{}) {
	l.withFields(ctx...).Fatal(msg)
}

func (l *logger) withFields(ctx ...interface{}) *log.Entry {
	if len(ctx)%2 != 0 {
		ctx = append(ctx, nil)
	}

	for i := 0; i < len(ctx)/2; i = i + 2 {
		l.contexts[fmt.Sprintf("%s", ctx[i])] = ctx[i+1]
	}

	return l.logger.WithFields(l.contexts)
}
