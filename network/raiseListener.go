package network

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/kechako/go-udpsample/netutil"
	"github.com/stream3715/RaiseLog/util"
)

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func envLoad() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

//RaiseListen ...Listen Clients' udp packet
func RaiseListen() {
	envLoad()

	var (
		HOST     = os.Getenv("HOST")
		DATABASE = os.Getenv("DATABASE")
		USER     = os.Getenv("USER")
		PASSWORD = os.Getenv("PASSWORD")
	)
	u, err := uuid.NewRandom()
	if err != nil {
		fmt.Println(err)
		return
	}
	uu := u.String()
	fmt.Println(uu)

	var connectionString string = fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=require", HOST, USER, PASSWORD, DATABASE)
	db, err := sql.Open("postgres", connectionString)
	checkError(err)
	defer db.Exec("DROP TABLE \"" + uu + "\"")
	defer db.Close()
	createTableString := "CREATE TABLE \"" + uu + "\"(time bigint PRIMARY KEY, name varchar(32));"
	_, err = db.Exec(createTableString)
	checkError(err)

	listenConfig := &net.ListenConfig{
		Control: netutil.ListenControl,
	}
	conn, err := listenConfig.ListenPacket(context.Background(), "udp", "localhost:8804")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buf := make([]byte, 1500)
	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Print(err)
			break
		} else if n == 0 {
			continue
		}

		fmt.Printf("data recv, %X\n", buf)
		go func() {
			nanoNow := time.Now().UnixNano()
			var recv []util.Receive
			n := bytes.IndexByte(buf, 0)
			if err := json.Unmarshal(buf[:n], &recv); err != nil {
				if err, ok := err.(*json.SyntaxError); ok {
					fmt.Println(string(buf[err.Offset-15 : err.Offset+15]))
				}
				log.Fatal(err)
			}
			for _, data := range recv {
				if data.Command == 0 {
					clientTime, _ := util.StrToInt64(data.Payload, 10)
					lag := nanoNow - clientTime
					conn.WriteTo([]byte(strconv.FormatInt(lag, 10)), addr)
					log.Printf("%v client joined : %v with time: %v", nanoNow, addr, clientTime)
				} else if data.Command == 1 {
					// 構造体のインスタンス化
					recvTime, _ := util.StrToInt64(data.Payload, 10)

					sqlStatement := "INSERT INTO \"" + uu + "\" VALUES (" + fmt.Sprint(recvTime) + ", '" + data.Name + "');"
					// INSERTを実行
					fmt.Println(sqlStatement)
					_, err = db.Exec(sqlStatement)

					conn.WriteTo([]byte("lock"), addr)
				}
			}
		}()
	}
}
