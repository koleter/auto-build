package model

import (
	"github.com/subchen/go-log"
	xl "xorm.io/xorm/log"
)

type SqlLogger struct {
	show  bool
	level int
}

func (s *SqlLogger) Debugf(format string, v ...interface{}) {
	log.Debugf(format, v...)
}

func (s *SqlLogger) Errorf(format string, v ...interface{}) {
	log.Errorf(format, v...)
}

func (s *SqlLogger) Infof(format string, v ...interface{}) {
	log.Infof(format, v...)
}

func (s *SqlLogger) Warnf(format string, v ...interface{}) {
	log.Warnf(format, v...)
}

func (s *SqlLogger) Level() xl.LogLevel {
	return xl.LogLevel(s.level)
}

func (s *SqlLogger) SetLevel(lv xl.LogLevel) {
	s.level = int(lv)
}

func (s *SqlLogger) ShowSQL(show ...bool) {
	if len(show) > 0 {
		s.show = show[0]
	}
}

func (s *SqlLogger) IsShowSQL() bool {
	return s.show
}

func (s *SqlLogger) BeforeSQL(context xl.LogContext) {

}

func (s *SqlLogger) AfterSQL(context xl.LogContext) {

}
