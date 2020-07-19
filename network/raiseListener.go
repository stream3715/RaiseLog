package network

import (
	"context"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/kechako/go-udpsample/netutil"
	"github.com/stream3715/RaiseLog/util"
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
			data := string(buf[:n])
			msg := strings.Split(data, "|")
			if msg[0] == "lagtest" {
				clientTime, _ := util.StrToInt64(msg[1], 10)
				lag := clientTime - time.Now().UnixNano()
				conn.WriteTo([]byte(strconv.FormatInt(lag, 10)), addr)
			}

			log.Printf("client joined : %v", addr)
		}()
	}
}
