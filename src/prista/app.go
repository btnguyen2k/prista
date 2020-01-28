package prista

import (
	"fmt"
	"github.com/btnguyen2k/singu"
	"github.com/btnguyen2k/singu/leveldb"
	"github.com/go-akka/configuration"
	"log"
	"main/src/logger"
	"main/src/utils"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	AppConfig  *configuration.Config
	LogConfig  *configuration.Config
	LogWriters map[string]logger.ILogWriter
	Buffer     singu.IQueue

	ConcurrentWrite int64 = 0
)

/*
Start bootstraps the application.
*/
func Start() {
	var err error
	AppConfig = initAppConfig()
	utils.Location, err = time.LoadLocation(AppConfig.GetString("timezone"))
	if err != nil {
		panic(err)
	}

	LogConfig = AppConfig.GetConfig("log")

	Buffer = initBuffer(LogConfig)

	LogWriters = initLogWriters(LogConfig)
	if LogWriters == nil {
		panic("no valid log writer configured")
	}
	if _, ok := LogWriters["default"]; !ok {
		panic("no valid log writer for 'default' category")
	}

	go goWriteLogs(Buffer)
	go goProcessOrphanLogs(Buffer)

	var wg sync.WaitGroup
	if initHttpServer(&wg) {
		wg.Add(1)
	}
	if initUdpServer(&wg) {
		wg.Add(1)
	}
	if initGrpcServer(&wg) {
		wg.Add(1)
	}
	wg.Wait()

	fmt.Printf("Application exists.")
}

const defaultConfigFile = "./config/application.conf"

func initAppConfig() *configuration.Config {
	configFile := os.Getenv("APP_CONFIG")
	if configFile == "" {
		log.Printf("No environment APP_CONFIG found, fallback to [%s]", defaultConfigFile)
		configFile = defaultConfigFile
	}
	return loadAppConfig(configFile)
}

func initBuffer(config *configuration.Config) singu.IQueue {
	tempDir := config.GetString("temp_dir", "./temp")
	return leveldb.NewLeveldbQueue("buffer", tempDir, 0, false, 0)
}

// Go routine to requeue orphan messages
func goProcessOrphanLogs(buffer singu.IQueue) {
	for {
		time.Sleep(11 * time.Second)
		if msgList, err := buffer.OrphanMessages(10, 1000); err != nil {
			log.Printf(fmt.Sprintf("ERROR: error fetchig orphan messages: %e", err))
		} else if len(msgList) > 0 {
			log.Printf(fmt.Sprintf("INFO: processing %d orphan messages...", len(msgList)))
			for _, msg := range msgList {
				if _, err := buffer.Requeue(msg.Id, false); err != nil {
					log.Printf(fmt.Sprintf("ERROR: error requeueing orphan message %s/%s: %e", msg.Id, string(msg.Payload), err))
				}
			}
		}
	}
}

// Go routine to fetch messages from buffer and send to log writer
func goWriteLogs(buffer singu.IQueue) {
	for {
		time.Sleep(1 * time.Second)
		var counter int64 = 0
		t1 := time.Now()
		for msg, err := buffer.Take(); err == nil && msg != nil; msg, err = buffer.Take() {
			counter++
			tokens := strings.Split(string(msg.Payload), "\t")
			var finish = true
			if len(tokens) == 2 {
				logWriter := LogWriters[tokens[0]]
				if logWriter == nil {
					logWriter = LogWriters["default"]
				}
				if logWriter == nil {
					log.Printf(fmt.Sprintf("WARM: no log writer found for category [%s]", tokens[0]))
				} else if err := logWriter.Write(tokens[0], tokens[1]); err != nil {
					log.Printf(fmt.Sprintf("ERROR: error writing log to [%s]: %e", tokens[0], err))
					finish = false
				}
			}
			if finish {
				if err := buffer.Finish(msg.Id); err != nil {
					log.Printf(fmt.Sprintf("ERROR: error finishing message %s/%s: %e", msg.Id, string(msg.Payload), err))
				}
			} else if _, err := buffer.Requeue(msg.Id, false); err != nil {
				log.Printf(fmt.Sprintf("ERROR: error requeueing message %s/%s: %e", msg.Id, string(msg.Payload), err))
			}
			if ConcurrentWrite > 0 && counter >= 100/(ConcurrentWrite+1) || time.Now().Unix()-t1.Unix() >= 10 {
				// throttle [buffer->log-writer] rate
				break
			}
		}
		if counter > 0 {
			log.Printf(fmt.Sprintf("INFO: %d log(s) written", counter))
		}
	}
}

func initLogWriters(config *configuration.Config) map[string]logger.ILogWriter {
	if config != nil && config.Root().IsObject() {
		result := make(map[string]logger.ILogWriter)
		for cat, conf := range config.Root().GetObject().Items() {
			if conf != nil && conf.IsObject() {
				cat = strings.ToLower(cat)
				if writer, err := logger.NewLogWriter(cat, conf.GetObject().Unwrapped()); err != nil {
					panic(err)
				} else {
					result[cat] = writer
				}
			} else {
				panic(fmt.Sprintf("invalid config for log writer [%s]: %v", cat, conf))
			}
		}
		return result
	}
	return nil
}

// initialize and start UDP server
func initUdpServer(wg *sync.WaitGroup) bool {
	fmt.Println("HEHEHEHE")
	listenPort := AppConfig.GetInt32("server.udp.listen_port", 0)
	if listenPort <= 0 {
		log.Println("No valid [server.udp.listen_port] configured, UDP Server is disabled.")
		return false
	}
	listenAddr := AppConfig.GetString("server.udp.listen_addr", "127.0.0.1")

	fmt.Println("UDP", listenAddr, listenPort)

	pc, err := net.ListenPacket("udp", fmt.Sprintf("%s:%d", listenAddr, listenPort))
	if err != nil {
		panic(err)
	}
	fmt.Println(pc)

	return true
}

// convenient function to handle incoming message
func handleIncomingMessage(payload []byte) error {
	// increase concurrency count to throttle [buffer->log-writer] rate
	atomic.AddInt64(&ConcurrentWrite, 1)
	defer atomic.AddInt64(&ConcurrentWrite, -1)

	_, err := Buffer.Queue(singu.NewQueueMessage(payload))
	return err
}
