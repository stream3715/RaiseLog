package util

//Receive ...struct of receive packet data
type Receive struct {
	Name    string `json:"name"`
	Command int    `json:"command"`
	Payload string `json:"payload"`
}
