package logger

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btnguyen2k/consu/reddo"
	"github.com/btnguyen2k/consu/semita"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// NewFileLogWriter creates a new log writer that writes logs to files on disk, initialized and ready for use.
//	- cat: log category name
//	- conf: log writer configurations
func NewFileLogWriter(cat string, confMap map[string]interface{}) (ILogWriter, error) {
	logWriter := &FileLogWriter{category: cat}
	return logWriter, logWriter.Init(confMap)
}

type FileLogWriter struct {
	category    string // log category
	root        string // root directory to store log files
	filePattern string // log file pattern (accept Go style of datetime format)

	currentFileName string
	currentFile     *os.File
	logType         string
	lock            sync.Mutex
	inited          bool
}

const (
	confRoot        = "root"
	confFilePattern = "file_pattern"
	confLogType     = "log_type"

	logTypeTsv     = "tsv"
	logTypeJson    = "json"
	defaultLogType = logTypeJson
)

// Init implements ILogWriter.Init
func (w *FileLogWriter) Init(confMap map[string]interface{}) error {
	w.lock.Lock()
	defer w.lock.Unlock()
	if !w.inited {
		log.Printf("Intializing FileLogWriter for category [%s]...", w.category)
		conf := semita.NewSemita(confMap)

		// config: root
		if root, err := conf.GetValueOfType(confRoot, reddo.TypeString); err != nil {
			return err
		} else {
			w.root = strings.TrimPrefix(strings.TrimSpace(root.(string)), "/")
		}
		if w.root == "" {
			log.Println("WARN: no root directory defined, default to current directory")
		}

		// config: file pattern
		if filePattern, err := conf.GetValueOfType(confFilePattern, reddo.TypeString); err != nil {
			return err
		} else {
			w.filePattern = strings.TrimSpace(filePattern.(string))
		}
		if w.filePattern == "" {
			return errors.New(fmt.Sprintf("no [%s] configuration defined", confFilePattern))
		}

		// config: log type
		logType, _ := conf.GetValueOfType(confLogType, reddo.TypeString)
		if logType == nil {
			logType = ""
		}
		w.logType = strings.TrimSpace(logType.(string))
		if w.logType != logTypeTsv && w.logType != logTypeJson {
			w.logType = defaultLogType
		}

		if err := os.MkdirAll(w.root, 0755); err != nil {
			return err
		}

		// w.currentFileName = time.Now().Format(w.filePattern)

		w.inited = true
	}
	return nil
}

func (w *FileLogWriter) syncAndClose(f *os.File) error {
	var err error
	if f != nil {
		err = f.Sync()
		if err != nil {
			f.Close()
		} else {
			err = f.Close()
		}
	}
	return err
}

// Destroy implements ILogWriter.Write
func (w *FileLogWriter) Destroy() error {
	err := w.syncAndClose(w.currentFile)
	return err
}

// RefreshConfig implements ILogWriter.RefreshConfig
func (w *FileLogWriter) RefreshConfig(conf map[string]interface{}) error {
	panic("implement me")
}

func (w *FileLogWriter) formatLogMessage(category, message string) []byte {
	switch w.logType {
	case logTypeTsv:
		return []byte(category + "\t" + strings.TrimSpace(message))
	case logTypeJson:
		js, _ := json.Marshal(map[string]string{
			"category": category,
			"message":  strings.TrimSpace(message),
		})
		return js
	}
	return nil
}

// Write implements ILogWriter.Write
func (w *FileLogWriter) Write(category, message string) error {
	if !w.inited {
		return errors.New("this log writer has not been initialized")
	}
	w.lock.Lock()
	defer w.lock.Unlock()

	// rotate file if needed
	fileName := time.Now().Format(w.filePattern)
	if w.currentFileName != fileName && w.currentFile != nil {
		log.Println(fmt.Sprintf("INFO: rotating file %s -> %s", w.currentFileName, fileName))
		if err := w.syncAndClose(w.currentFile); err != nil {
			return err
		}
		w.currentFile = nil
		w.currentFileName = fileName
	}
	if w.currentFileName == "" {
		w.currentFileName = fileName
	}

	if w.currentFile == nil {
		var err error
		w.currentFile, err = os.OpenFile(w.root+"/"+w.currentFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		log.Println(fmt.Sprintf("INFO: opened file %s", w.currentFileName))
	}

	data := w.formatLogMessage(category, message)
	if data == nil {
		return errors.New("cannot format log message for writing")
	}
	_, err := w.currentFile.Write(data)
	if err == nil {
		_, err = w.currentFile.Write([]byte("\n"))
	}
	return err
}
