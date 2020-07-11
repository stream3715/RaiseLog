package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"fyne.io/fyne"

	"fyne.io/fyne/app"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func main() {
	sig := make(chan string, 100)
	res := make(chan string, 100)
	defer close(sig)
	defer close(res)

	flag.Parse()

	if flag.Arg(0) == "server" {
		go server(sig, res)
		serverview()
	} else {
		name := ""
		ip := "106.73.12.0"
		stats := "init"

		essential := senderEssentials{sig, res, &name, &ip, &stats}
		clientview(&essential)
	}
}

func server(sig chan string, res chan string) {
	udpAddr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 43983,
	}
	updLn, err := net.ListenUDP("udp", udpAddr)

	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	buf := make([]byte, 128)
	log.Println("Starting udp server...")

	for {
		n, addr, err := updLn.ReadFromUDP(buf)
		if err != nil {
			log.Fatalln(err)
			os.Exit(1)
		}

		go func() {
			log.Printf("Reciving data: %s from %s", string(buf[:n]), addr.String())

			//log.Printf("Sending data..")
			updLn.WriteTo([]byte("lock"), addr)
			//log.Printf("Complete Sending data..")
		}()
	}
}

func serverview() {
	a := app.New()
	w := a.NewWindow("Raise")
	w.SetContent(widget.NewVBox(
		widget.NewLabel("Raise Server"),
		widget.NewForm(),
		widget.NewVBox(
			widget.NewButton("Flush", func() {
				log.Println("----------------------Next Question----------------------")
			}),
			widget.NewButton("Quit", func() {
				a.Quit()
			}),
		),
	))
	w.Resize(fyne.NewSize(400, 300))
	w.ShowAndRun()
}

func send(essential *senderEssentials) {
	conn, err := net.Dial("udp", *essential.ip+":43983")
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	defer conn.Close()

	n, err := conn.Write([]byte(*essential.name))
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	if n != len(*essential.name) {
		log.Println("Err: buf length mismatch")
		err := "Send error, call the Judge member."
		essential.stats = &err
	}
	raised := "Raised"
	essential.stats = &raised
	time.Sleep(time.Second * 3)
	online := "Online"
	essential.stats = &online
}

//Screen
func makeWelcome(essential *senderEssentials) *widget.Box {
	//Label
	lbl := widget.NewLabel("Player name : " + *essential.name)
	lblstats := widget.NewLabel("Status : ")

	//Button
	btnRaise := widget.NewButton("Raise!", func() {
		go send(essential)
	})
	go func() {
		t := time.NewTicker(time.Second)
		for range t.C {
			lbl.SetText("Player name : " + *essential.name)
			lblstats.SetText("Status : " + *essential.stats)

		}
	}()

	return widget.NewVBox(lbl, lblstats, btnRaise)
}

//Screen
func makeSettings(essential *senderEssentials) *widget.Box {
	entryIP := widget.NewEntry()
	entryName := widget.NewEntry()
	entryIP.SetText(*essential.ip)

	form := &widget.Form{
		Items: []*widget.FormItem{ // we can specify items in the constructor
			{"IP", entryIP}, {"Name", entryName},
		},
		OnSubmit: func() { // optional, handle form submission
			essential.sig <- "ip|" + entryIP.Text
			essential.sig <- "name|" + entryName.Text
		},
	}

	return widget.NewVBox(form)
}

type senderEssentials struct {
	sig   chan string
	res   chan string
	name  *string
	ip    *string
	stats *string
}

func chanListener(essential *senderEssentials) bool {
	msg := <-essential.sig
	if msg != "" {
		log.Println("sig recv : " + msg)
		split := strings.Split(msg, "|")
		switch split[0] {
		case "ip":
			essential.ip = &split[1]
			return true
		case "name":
			essential.name = &split[1]
			return true
		default:
			break
		}

	}
	return false
}

func clientview(essential *senderEssentials) {
	a := app.New()
	w := a.NewWindow("Raise")
	tabs := widget.NewTabContainer(
		widget.NewTabItemWithIcon("AnswerButton", theme.HomeIcon(), makeWelcome(essential)),
		widget.NewTabItemWithIcon("Setting", theme.SettingsIcon(), makeSettings(essential)),
	)
	w.SetContent(tabs)

	w.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		switch key.Name {
		case fyne.KeySpace:
			go send(essential)
		default:
			break
		}
	})
	go func(essential *senderEssentials, tabs *widget.TabContainer) {
		for {
			chanListener(essential)
		}
	}(essential, tabs)
	w.ShowAndRun()

}
