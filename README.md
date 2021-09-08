## Connection Logger
This is a simple utility that simply pings a 8.8.8.8 every few seconds.
It fires 3 pings and counds as an outage if all of them are lost.
It will then time consecutive outages and record them into a file `info.log`
small outages (less than 10 seconds) will not be logged

Every 10 minutes an outage summary is printed to stdout

# run with:
`go run main.go`
