// Copyright 2015 Ka-Hing Cheung
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

/*
 * JuiceFS, Copyright (C) 2020 Juicedata, Inc.
 *
 * This program is free software: you can use, redistribute, and/or modify
 * it under the terms of the GNU Affero General Public License, version 3
 * or later ("AGPL"), as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package utils

import (
	"fmt"
	glog "log"
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

var mu sync.Mutex
var loggers = make(map[string]*logHandle)

var syslogHook logrus.Hook

type logHandle struct {
	logrus.Logger

	name string
	lvl  *logrus.Level
}

func (l *logHandle) Format(e *logrus.Entry) ([]byte, error) {
	// Mon Jan 2 15:04:05 -0700 MST 2006
	timestamp := ""
	lvl := e.Level
	if l.lvl != nil {
		lvl = *l.lvl
	}

	const timeFormat = "2006/01/02 15:04:05.000000"
	timestamp = e.Time.Format(timeFormat)

	str := fmt.Sprintf("%v %s[%d] <%v>: %v",
		timestamp,
		l.name,
		os.Getpid(),
		strings.ToUpper(lvl.String()),
		e.Message)

	if len(e.Data) != 0 {
		str += " " + fmt.Sprint(e.Data)
	}

	str += "\n"
	return []byte(str), nil
}

// for aws.Logger
func (l *logHandle) Log(args ...interface{}) {
	l.Debugln(args...)
}

func NewLogger(name string) *logHandle {
	l := &logHandle{name: name}
	l.Out = os.Stderr
	l.Formatter = l
	l.Level = logrus.InfoLevel
	l.Hooks = make(logrus.LevelHooks)
	if syslogHook != nil {
		l.Hooks.Add(syslogHook)
	}
	return l
}

func GetLogger(name string) *logHandle {
	mu.Lock()
	defer mu.Unlock()

	if logger, ok := loggers[name]; ok {
		return logger
	}
	logger := NewLogger(name)
	loggers[name] = logger
	return logger
}

func GetStdLogger(l *logHandle, lvl logrus.Level) *glog.Logger {
	mu.Lock()
	defer mu.Unlock()

	w := l.Writer()
	l.Formatter.(*logHandle).lvl = &lvl
	l.Level = lvl
	return glog.New(w, "", 0)
}

func SetLogLevel(lvl logrus.Level) {
	for _, logger := range loggers {
		logger.Level = lvl
	}
}
