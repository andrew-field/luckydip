package main

import (
	"github.com/andrew-field/luckydip"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
)

// Used for testing. Can run for instance: export GOOGLEAPPPASSWORD=dasdagvsgsdgf and then go run . (-rod="show,trace,slow=1s,monitor=:1234") etc.
// browser.ServeMonitor("0.0.0.0:1234") // Open a browser and navigate to this address.

func main() {
	wsURL := launcher.New().Bin("/run/current-system/sw/bin/google-chrome-stable").MustLaunch()

	browser := rod.New().ControlURL(wsURL).Trace(true).MustConnect().NoDefaultDevice()

	// An effort to avoid bot detection.
	page := stealth.MustPage(browser)

	luckydip.FreeBirthdateLottery(page)
}
