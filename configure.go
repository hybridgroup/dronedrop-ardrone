package main

import (
	"log"
	"os"

	"github.com/jlaffaye/ftp"
	"github.com/ziutek/telnet"
)

func main() {
	c, err := ftp.Connect("192.168.1.1:21")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Pushing dronedrop...")
	if b, err := os.Open("configure_drone_drop.sh"); err != nil {
		log.Fatal("open", err)
	} else {
		if err := c.Stor("/configure_drone_drop.sh", b); err != nil {
			log.Fatal(err)
		}
	}

	if b, err := os.Open("ardrone_commander"); err != nil {
		log.Fatal("open", err)
	} else {
		if err := c.Stor("/ardrone_commander", b); err != nil {
			log.Fatal(err)
		}
	}

	c.Quit()

	t, err := telnet.Dial("tcp", "192.168.1.1:23")
	if err != nil {
		log.Fatal(err)
	}
	t.SetUnixWriteMode(true)

	log.Println("Configuring dronedrop...")
	buf := []byte("sh /data/video/configure_drone_drop.sh\n")
	if _, err := t.Write(buf); err != nil {
		log.Fatal(err)
	}
	t.Close()
	log.Println("Rebooting drone...")
}
