package network

import (
	"context"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/kechako/go-udpsample/netutil"
)

//RaiseListen ...Listen Clients' udp packet
func RaiseListen() {

	listenConfig := &net.ListenConfig{
		Control: netutil.ListenControl,
	}
	conn, err := listenConfig.ListenPacket(context.Background(), "udp", "127.0.0.1:8804")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	var buf [1500]byte
	for {
		n, addr, err := conn.ReadFrom(buf[:])
		if err != nil {
			log.Print(err)
			break
		}
		go func() {
			msg := strings.Split(string(buf[:n]), "|")
			if msg[0] == "lagtest" {
				clientTime, _ := strconv.Atoi(msg[1])
				lag := clientTime - time.Now().UnixNano()
				conn.WriteTo([]byte(string(lag)), addr)
			}

			log.Printf("client joined : %v", addr)
		}()
	}
}
