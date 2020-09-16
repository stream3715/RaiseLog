package network

import (
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

var (
	HOST     = os.Getenv("HOST")
	DATABASE = os.Getenv("DATABASE")
	USER     = os.Getenv("USER")
	PASSWORD = os.Getenv("PASSWORD")
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
	conn, err := listenConfig.ListenPacket(context.Background(), "udp", ":8804")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	var buf []byte
	for {
		_, addr, err := conn.ReadFrom(buf[:])
		if buf == nil {
			continue
		}
		if err != nil {
			log.Print(err)
			break
		}
		go func() {
			nanoNow := time.Now().UnixNano()
			var recv []util.Receive
			if err := json.Unmarshal(buf, &recv); err != nil {
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

					sqlStatement := "INSERT INTO $1 (name, quantity) VALUES ($2, $3);"
					// INSERTを実行
					_, err = db.Exec(sqlStatement, uu, recvTime, data.Name)

					conn.WriteTo([]byte("lock"), addr)
				}
			}
		}()
	}
}
