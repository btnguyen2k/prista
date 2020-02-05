package prista

import (
	"fmt"
	"log"
	"math/big"
	"net"
	"sync"
)

// initialize and start UDP server
func initUdpServer(wg *sync.WaitGroup, numServers int) bool {
	listenPort := AppConfig.GetInt32("server.udp.listen_port", 0)
	if listenPort <= 0 {
		log.Println("No valid [server.udp.listen_port] configured, UDP server is disabled.")
		return false
	}
	listenAddr := AppConfig.GetString("server.udp.listen_addr", "127.0.0.1")

	pc, err := net.ListenPacket("udp", fmt.Sprintf("%s:%d", listenAddr, listenPort))
	if err != nil {
		panic(err)
	}
	log.Printf("Starting [%s] UDP server on [%s:%d]...\n", AppConfig.GetString("app.name")+" v"+AppConfig.GetString("app.version"), listenAddr, listenPort)
	bodyLimit := AppConfig.GetByteSize("server.max_request_size")
	if bodyLimit == nil || bodyLimit.Int64() <= 0 {
		bodyLimit = big.NewInt(4086)
	}
	buffer := make([]byte, bodyLimit.Int64())
	for i := 0; i < numServers; i++ {
		go func() {
			defer pc.Close()
			for {
				// ReadFrom blocks until data received or timed-out
				n, _, err := pc.ReadFrom(buffer)
				if err != nil {
					log.Printf(fmt.Sprintf("ERROR: error while reading UDP data: %e", err))
				} else if n > 0 {
					go func(payload []byte) {
						if err := handleIncomingMessage(payload, true); err != nil {
							log.Printf(err.Error())
						}
					}(buffer[:n])
				}
			}
			wg.Done()
		}()
	}
	return true
}
