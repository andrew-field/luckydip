package main

import "github.com/andrew-field/luckydip"

// Used for testing. Can run for instance: export GOOGLEAPPPASSWORD=dasdagvsgsdgf and then go run . -rod="show,trace,slow=1s,monitor=:1234" etc.
// browser.ServeMonitor("0.0.0.0:1234") // Open a browser and navigate to this address.

func main() {
	luckydip.WinADinner()
}
