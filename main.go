package main

import (
	"github.com/stream3715/RaiseLog/network"

	_ "github.com/lib/pq"
)

func main() {
	//chan toMain := make(chan )
	network.RaiseListen()
}
