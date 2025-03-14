package luckydip

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/stealth"
)

const UnknownError = "Unknown error"
const TimeoutError = "Timeout error"
const WinOutcome = "WIN!"
const LoseOutcome = "Lose"

// HelloHTTP is an HTTP Cloud Function with a request parameter.
func HelloHTTP(_ http.ResponseWriter, _ *http.Request) {
	// Create browser
	browser := rod.New().MustConnect().Trace(true).Timeout(time.Second * 200)

	// An effort to avoid bot detection.
	page := stealth.MustPage(browser)

	// Load the London time zone.
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		panic("Could not load time in the location. HelloHTTP Entry. Error: " + err.Error())
	}

	switch time.Now().In(loc).Hour() {
	case 8:
		Euromillions(page)
	case 9:
		PickMyPostcode(page)
	case 11:
		FreeBirthdateLottery(page)
	case 16:
		WinADinner(page)
	case 18:
		PickMyPostcode(page)
	case 21:
		PickMyPostcode(page)
	default:
		sendEmail("andrew_field+luckdiperror@hotmail.co.uk", "Error in cloud function execution", "Could not select the correct function. Time: "+time.Now().In(loc).String(), nil)
	}
}

func checkProcessError(err error, to string, page *rod.Page) bool {
	if err != nil {
		summary := UnknownError
		if errors.Is(err, context.DeadlineExceeded) {
			summary = TimeoutError
		}

		// If err is not nil, send an error email.
		sendEmail(to, summary, err.Error(), page.CancelTimeout().MustScreenshot())

		return true
	}
	return false
}
