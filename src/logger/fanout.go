package logger

import (
	"errors"
	"fmt"
	"github.com/btnguyen2k/consu/reddo"
	"github.com/btnguyen2k/consu/semita"
	"log"
	"regexp"
	"sync"
)

// NewFanoutLogWriter creates a new log writer that fan-outs logs to other log writers, initialized and ready for use.
//	- cat: log category name
//	- conf: log writer configurations
func NewFanoutLogWriter(cat string, confMap map[string]interface{}, enqueueFunc FuncEnqueue) (ILogWriter, error) {
	logWriter := &FanoutLogWriter{category: cat, enqueueFunc: enqueueFunc}
	return logWriter, logWriter.Init(confMap)
}

// FanoutLogWriter fan-outs logs to other log writers
// @available since v0.1.3
type FanoutLogWriter struct {
	category string   // log category
	targets  []string // destination to fan-out log entry to

	enqueueFunc FuncEnqueue // function to enqueue log entry
	lock        sync.Mutex
	inited      bool
}

const (
	confFanoutTargets = "targets"
)

// Info implements ILogWriter.Info
func (w *FanoutLogWriter) Info() map[string]interface{} {
	return map[string]interface{}{
		"name":          "fanout",
		"desc":          "This log writer fan-outs log messages to other log writers",
		"retry_seconds": 0,
	}
}

// Init implements ILogWriter.Init
func (w *FanoutLogWriter) Init(confMap map[string]interface{}) error {
	fmt.Printf("%#v\n", confMap)

	w.lock.Lock()
	defer w.lock.Unlock()
	if !w.inited {
		log.Printf("Intializing FanoutLogWriter for category [%s]...", w.category)
		conf := semita.NewSemita(confMap)

		if w.enqueueFunc == nil {
			return errors.New("enqueue function is not assigned")
		}

		// config: targets
		if targets, err := conf.GetValueOfType(confFanoutTargets, reddo.TypeString); err != nil {
			return err
		} else {
			w.targets = regexp.MustCompile("[,;\\s]+").Split(targets.(string), -1)
		}
		if len(w.targets) == 0 {
			return errors.New("empty target category list")
		}

		w.inited = true
	}
	return nil
}

// Destroy implements ILogWriter.Write
func (w *FanoutLogWriter) Destroy() error {
	return nil
}

// RefreshConfig implements ILogWriter.RefreshConfig
func (w *FanoutLogWriter) RefreshConfig(conf map[string]interface{}) error {
	panic("implement me")
}

// Write implements ILogWriter.Write
func (w *FanoutLogWriter) Write(category, message string) error {
	if !w.inited {
		return errors.New("this log writer has not been initialized")
	}
	w.lock.Lock()
	defer w.lock.Unlock()

	for _, target := range w.targets {
		if err := w.enqueueFunc([]byte(target+SeparatorTsv+message), false); err != nil {
			return err
		}
	}

	return nil
}
