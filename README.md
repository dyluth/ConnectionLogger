## Connection Logger
This is a simple utility that simply pings a URL every few seconds.
It fires 3 pings and counds as an outage if all of them are lost.
It will then time consecutive outages and record them into a file `info.log`

# run with:
`go run main.go`
