package network

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/stream3715/RaiseLog/util"
)

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

//RaiseListen ...Listen Clients' udp packet
func RaiseListen(uu string, conn net.PacketConn, db *sql.DB) {
	createTableString := "CREATE TABLE \"" + uu + "\"(time bigint PRIMARY KEY, name varchar(32));"
	_, err := db.Exec(createTableString)
	checkError(err)
	reset := false
	release := false

	for {
		buf := make([]byte, 1500)
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Print(err)
			break
		} else if n == 0 {
			continue
		}

		go func() {
			var recv []util.Receive
			n := bytes.IndexByte(buf, 0)
			if err := json.Unmarshal(buf[:n], &recv); err != nil {
				if err, ok := err.(*json.SyntaxError); ok {
					fmt.Println(string(buf[err.Offset-15 : err.Offset+15]))
				}
				log.Fatal(err)
			}
			for _, data := range recv {
				nanoNow := time.Now().UnixNano()
				if data.Command == 0 {
					controlBit := ""
					if reset == true {
						controlBit = "8"
					} else if release == true {
						controlBit = "9"
					}
					clientTime, _ := util.StrToInt64(data.Payload, 10)
					lag := nanoNow - clientTime
					conn.WriteTo([]byte(controlBit+","+strconv.FormatInt(lag, 10)), addr)
				} else if data.Command == 1 {
					// 構造体のインスタンス化
					lag, _ := util.StrToInt64(data.Payload, 10)
					recvTime := nanoNow + lag
					sqlStatement := "INSERT INTO \"" + uu + "\" VALUES (" + fmt.Sprint(recvTime) + ", '" + data.Name + "');"
					// INSERTを実行
					log.Printf("%v client raised : %v with time: %v", nanoNow, addr, recvTime)
					fmt.Println(sqlStatement)
					_, err = db.Exec(sqlStatement)

					conn.WriteTo([]byte("lock"), addr)
				} else if data.Command == 3 {
					reset = true
					release = false
					sqlStatement := "TRUNCATE \"" + uu + "\";"
					// INSERTを実行
					fmt.Println("Wait")
					_, err = db.Exec(sqlStatement)
				} else if data.Command == 4 {
					fmt.Println("Ready")
					reset = false
					release = true
					time.Sleep(time.Second)
					release = false
				}
			}
		}()
	}
}
