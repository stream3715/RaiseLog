package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/kechako/go-udpsample/netutil"
	"github.com/stream3715/RaiseLog/network"

	_ "github.com/lib/pq"
)

func main() {
	//chan toMain := make(chan )
	quit := make(chan os.Signal)

	// 受け取るシグナルを設定
	signal.Notify(quit, os.Interrupt)

	// ここからUDP受信設定
	listenConfig := &net.ListenConfig{
		Control: netutil.ListenControl,
	}
	conn, err := listenConfig.ListenPacket(context.Background(), "udp", "0.0.0.0:8804")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// UUID生成
	uu, err := genUUID()
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	fmt.Println(uu)

	//DB接続の取得
	db := defineDbConnection()
	defer db.Exec("DROP TABLE IF EXISTS \"" + uu + "\"")
	defer db.Close()

	// 実行
	go network.RaiseListen(uu, conn, db)

	//以降SIGINT受け取り後処理
	<-quit
	fmt.Println("SIGINT")
	conn.Close()
	db.Exec("DROP TABLE IF EXISTS \"" + uu + "\"")
	db.Close()

}

func envLoad() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func genUUID() (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	uu := u.String()
	return uu, nil
}

func defineDbConnection() *sql.DB {
	envLoad()

	var (
		HOST     = os.Getenv("DB_HOST")
		DATABASE = os.Getenv("DB_DATABASE")
		USER     = os.Getenv("DB_USER")
		PASSWORD = os.Getenv("DB_PASSWORD")
	)

	var connectionString string = fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=require", HOST, USER, PASSWORD, DATABASE)
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	return db
}
