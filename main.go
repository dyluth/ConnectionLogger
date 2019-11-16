package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sparrc/go-ping"
)

func main() {

	fmt.Println("Initted... " + time.Now().Format(time.ANSIC))
	// TODO log out the start time for the log file
	file, err := os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	log.SetOutput(file)
	log.Print(fmt.Sprintf("Initted... %v\n", time.Now().Format(time.ANSIC)))

	ticker := time.NewTicker(2000 * time.Millisecond)
	pingID := 0
	var outageStart *time.Time = nil
	for {
		select {
		case <-ticker.C:
			pingID++
			packetloss := pingNumber(3)
			fmt.Printf("Ping: %v - loss: %v", pingID, packetloss)
			if packetloss >= 90 {
				fmt.Printf(" BAD!!\n")
				// todo log outage to file
				if outageStart == nil {
					now := time.Now()
					outageStart = &now
				}
			} else {
				fmt.Printf(" OK\n")
				if outageStart != nil {
					// TODO put in file
					outageEnd := time.Now()
					log.Printf("OUTAGE TIME, FROM: %v  TO: %v  (Duration %v)\n", outageStart.Format(time.ANSIC), outageEnd.Format(time.ANSIC), outageEnd.Sub(*outageStart).Round(time.Millisecond))
					outageStart = nil
				}
			}
		}
	}
}

func pingNumber(count int) (packetLoss float64) {
	pinger, err := ping.NewPinger("www.google.com")
	// error is returned if DNS name resolution fails - in this case record 100% packet loss
	if err != nil {
		return 100
	}

	pinger.Count = count
	pinger.Timeout = 1000 * time.Millisecond
	pinger.Interval = 100 * time.Millisecond
	pinger.Run()                 // blocks until finished
	stats := pinger.Statistics() // get send/receive/rtt stats
	return stats.PacketLoss
}
