package logger

import (
	"errors"
	"fmt"
	"github.com/btnguyen2k/consu/reddo"
	"github.com/btnguyen2k/consu/semita"
	"reflect"
)

// ILogWriter defines API to write log message.
type ILogWriter interface {
	// Init initializes the writer with initial configurations. Writer is considered not ready for use until Init is called successfully.
	Init(conf map[string]interface{}) error

	// Destroy is called to clear up the writer. Writer is no longer usable after Destroy is called.
	Destroy() error

	// RefreshConfig updates writer's configuration live.
	RefreshConfig(conf map[string]interface{}) error

	// Write writes a log message to a category
	Write(category, message string) error
}

// NewLogWriter creates a new log writer instance, initialized and ready for use.
//	- cat: log category name
//	- conf: log writer configurations
func NewLogWriter(cat string, confMap map[string]interface{}) (ILogWriter, error) {
	typeMap := reflect.TypeOf(map[string]interface{}{})
	conf := semita.NewSemita(confMap)
	wrtType, err := conf.GetValueOfType("type", reddo.TypeString)
	if err != nil {
		return nil, err
	}
	switch wrtType.(string) {
	case "file":
		if confFile, err := conf.GetValueOfType("file", typeMap); err != nil {
			return nil, err
		} else {
			return NewFileLogWriter(cat, confFile.(map[string]interface{}))
		}
	default:
		errors.New(fmt.Sprintf("unknown writer type [%s]", wrtType))
	}
	return nil, nil
}
