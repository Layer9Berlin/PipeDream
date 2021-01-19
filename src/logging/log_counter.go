package logging

import "github.com/sirupsen/logrus"

type LogCounterHook struct {
	level logrus.Level
	counter *int
}

func NewLogCounterHook(level logrus.Level, counter *int) *LogCounterHook {
	return &LogCounterHook{
		level: level,
		counter: counter,
	}
}


func (counterHook *LogCounterHook)Fire(*logrus.Entry) error {
	*counterHook.counter = *counterHook.counter + 1
	return nil
}

func (counterHook *LogCounterHook)Levels() []logrus.Level {
	return []logrus.Level{counterHook.level}
}
