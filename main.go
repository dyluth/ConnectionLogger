package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
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
	tracker := NewOutageTracker()
	//TODO Add sigterm handler to print total summary!
	defer file.Close()

	log.SetOutput(file)
	log.Print(fmt.Sprintf("Initted... %v\n", time.Now().Format(time.ANSIC)))

	ticker := time.NewTicker(2000 * time.Millisecond)
	pingID := 0
	lastPing := time.Now()
	lastSummary := time.Now()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		tracker.printTerminateSummary()
		os.Exit(0)
	}()

	for {
		select {
		case <-ticker.C:
			pingID++
			packetloss := pingNumber(3)
			//fmt.Printf("Ping: %v - loss: %v", pingID, packetloss)
			if lastPing.Sub(time.Now()) > 3*time.Second {
				fmt.Printf(" waking up \n")
				lastPing = time.Now()
				tracker.clearOutage()
				continue
			}

			lastPing = time.Now()
			if packetloss >= 90 {
				if tracker.InOutage() {
					fmt.Printf("X")

				} else {
					fmt.Printf("x")
				}
				tracker.startOutage()
			} else {
				fmt.Printf(".")
				tracker.noOutage()
				/*
					if outageStart != nil {
						// TODO put in file
						outageEnd := time.Now()
						length := outageEnd.Sub(*outageStart).Round(time.Millisecond)
						sumaryOutageTotal += length
						if length > 5*time.Second {
							log.Printf("OUTAGE TIME, FROM: %v  TO: %v  (Duration %v)\n", outageStart.Format(time.ANSIC), outageEnd.Format(time.ANSIC), length)
						}
						outageStart = nil
					}
				*/
				if lastSummary.Before(time.Now().Add(-10 * time.Minute)) {
					lastSummary = lastPing
					//fmt.Printf("\n%v (outagetotal: %v) ", lastPing.Format(time.ANSIC), sumaryOutageTotal.String())
					tracker.printOutageSummary()
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

type OutageTracker struct {
	outageStart       time.Time
	previousWasOutage bool
	totalStartTime    time.Time
	totalOutage       time.Duration
	outageAtLastPrint time.Duration
	timeOfLastPrint   time.Time
}

func NewOutageTracker() OutageTracker {
	return OutageTracker{
		totalStartTime:  time.Now(),
		timeOfLastPrint: time.Now(),
	}
}

func (ot *OutageTracker) startOutage() {
	if !ot.previousWasOutage {
		ot.outageStart = time.Now()
		ot.previousWasOutage = true
	}
}
func (ot *OutageTracker) clearOutage() {
	ot.previousWasOutage = false
}

func (ot *OutageTracker) InOutage() bool {
	if ot.previousWasOutage && time.Now().Sub(ot.outageStart) > 10*time.Second {
		return true
	}
	return false
}

func (ot *OutageTracker) noOutage() {
	if ot.previousWasOutage {
		end := time.Now()
		dur := end.Sub(ot.outageStart)
		if dur > 10*time.Second {
			ot.totalOutage += dur
			// now record to disk!
			log.Printf("OUTAGE TIME, FROM: %v  TO: %v  (Duration %v)\n", ot.outageStart.Format(time.ANSIC), end.Format(time.ANSIC), ot.totalOutage)
		}
	}
	ot.previousWasOutage = false
}

func (ot *OutageTracker) printOutageSummary() {
	diff := ot.totalOutage - ot.outageAtLastPrint
	timeChange := time.Now().Sub(ot.timeOfLastPrint)
	ot.timeOfLastPrint = time.Now()
	fmt.Printf("\n%v: %v/%v: ", ot.timeOfLastPrint.Format(time.ANSIC), diff, timeChange)
	ot.outageAtLastPrint = ot.totalOutage

}

func (ot *OutageTracker) printTerminateSummary() {
	timeChange := time.Now().Sub(ot.totalStartTime)
	percentage := (ot.totalOutage.Nanoseconds() / timeChange.Nanoseconds()) * 100
	fmt.Printf("\ntotal outage %v/%v: %v percent\n", ot.totalOutage, timeChange, percentage)
}
