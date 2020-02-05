package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btnguyen2k/consu/reddo"
	"github.com/btnguyen2k/consu/semita"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	pb "main/src/grpc"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// NewForwardLogWriter creates a new log writer that forwards log entries to another prista instance, initialized and ready for use.
//	- cat: log category name
//	- conf: log writer configurations
func NewForwardLogWriter(cat string, confMap map[string]interface{}) (ILogWriter, error) {
	logWriter := &ForwardLogWriter{category: cat}
	return logWriter, logWriter.Init(confMap)
}

// ForwardLogWriter forwards logs to another prista instance
// @available since v0.1.1
type ForwardLogWriter struct {
	category     string // log category
	destination  string // destination to forward log entry to
	retrySeconds int    // number of seconds to retry writing log entry in case of failure

	destProtocol string                        // udp, grpc or http/https
	udpAddr      *net.UDPAddr                  // for UDP destination
	grpcConn     *grpc.ClientConn              // for gRPC client
	grpcClient   pb.PLogCollectorServiceClient // for gRPC client
	httpBase     string                        // for HTTP client
	httpClient   *http.Client                  // for HTTP client
	lock         sync.Mutex
	inited       bool
}

const (
	confForwardDestination = "destination"
)

// Info implements ILogWriter.Info
func (w *ForwardLogWriter) Info() map[string]interface{} {
	return map[string]interface{}{
		"name":          "forward",
		"desc":          "This log writer forwards log messages to another prista instance",
		"retry_seconds": w.retrySeconds,
	}
}

// Init implements ILogWriter.Init
func (w *ForwardLogWriter) Init(confMap map[string]interface{}) error {
	w.lock.Lock()
	defer w.lock.Unlock()
	if !w.inited {
		log.Printf("Intializing ForwardLogWriter for category [%s]...", w.category)
		conf := semita.NewSemita(confMap)

		// config: destination
		if destination, err := conf.GetValueOfType(confForwardDestination, reddo.TypeString); err != nil {
			return err
		} else {
			w.destination = strings.TrimPrefix(strings.TrimSpace(destination.(string)), "")
		}
		if w.destination == "" {
			return errors.New(fmt.Sprintf("no [%s] configuration defined", confForwardDestination))
		}
		if url, err := url.Parse(w.destination); err != nil {
			return err
		} else if url == nil {
			return errors.New(fmt.Sprintf("cannot parse destination [%s]", w.destination))
		} else if url.Scheme != "udp" && url.Scheme != "grpc" && url.Scheme != "http" && url.Scheme != "https" {
			return errors.New(fmt.Sprintf("unsupported destination [%s]", w.destination))
		} else {
			switch url.Scheme {
			case "udp":
				w.destProtocol = "udp"
				if udpAddr, err := net.ResolveUDPAddr("udp", url.Host); err != nil {
					return err
				} else {
					w.udpAddr = udpAddr
				}
			case "grpc":
				w.destProtocol = "grpc"
				if conn, err := grpc.Dial(url.Host, grpc.WithInsecure()); err != nil {
					return err
				} else {
					w.grpcConn = conn
					w.grpcClient = pb.NewPLogCollectorServiceClient(w.grpcConn)
				}
			case "http", "https":
				w.destProtocol = "http"
				w.httpBase = url.Scheme + "://" + url.Host
				w.httpClient = &http.Client{Timeout: 5 * time.Second}
			}
		}

		if retrySeconds, err := conf.GetValueOfType(ConfRetrySeconds, reddo.TypeInt); err != nil {
			w.retrySeconds = DefaultRetrySeconds
		} else {
			w.retrySeconds = int(retrySeconds.(int64))
		}

		w.inited = true
	}
	return nil
}

// Destroy implements ILogWriter.Write
func (w *ForwardLogWriter) Destroy() error {
	var err error
	if w.grpcConn != nil {
		err = w.grpcConn.Close()
	}
	return err
}

// RefreshConfig implements ILogWriter.RefreshConfig
func (w *ForwardLogWriter) RefreshConfig(conf map[string]interface{}) error {
	panic("implement me")
}

// Write implements ILogWriter.Write
func (w *ForwardLogWriter) Write(category, message string) error {
	if !w.inited {
		return errors.New("this log writer has not been initialized")
	}
	w.lock.Lock()
	defer w.lock.Unlock()

	switch w.destProtocol {
	case "udp":
		if conn, err := net.DialUDP("udp", nil, w.udpAddr); err != nil {
			return err
		} else {
			defer conn.Close()
			buff := []byte(category + SeparatorTsv + message)
			_, err := conn.Write(buff)
			return err
		}
	case "grpc":
		if result, err := w.grpcClient.Log(context.Background(), &pb.PLogMessage{Category: category, Message: message}); err != nil {
			return err
		} else if result.Status != 200 {
			return errors.New(fmt.Sprintf("error while forwarding message via gRPC. Status: %d / Category: %s / Message: %s", result.Status, category, message))
		}
	case "http", "https":
		url := w.httpBase + "/api/log"
		data := map[string]string{"category": category, "message": message}
		js, _ := json.Marshal(data)
		body := bytes.NewBuffer(js)
		if resp, err := w.httpClient.Post(url, "application/json", body); err != nil {
			return err
		} else {
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return errors.New(fmt.Sprintf("error while forwarding message to [%s]. Status: %s / Category: %s / Message: %s", url, resp.Status, category, message))
			}
			if buff, err := ioutil.ReadAll(resp.Body); err != nil {
				return err
			} else {
				var result map[string]interface{}
				if err := json.Unmarshal(buff, &result); err != nil {
					return err
				}
				s := semita.NewSemita(result)
				if status, err := s.GetValueOfType("status", reddo.TypeInt); err != nil {
					return err
				} else if status.(int64) != 200 {
					return errors.New(fmt.Sprintf("error while forwarding message to [%s]. Response: %v", url, result))
				}
			}
		}
	}
	return nil
}
