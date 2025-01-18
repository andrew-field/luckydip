package luckydip

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-rod/rod"
)

const UnknownError = "Unknown error"
const TimeoutError = "Timeout error"
const WinOutcome = "WIN!"
const LoseOutcome = "Lose"

// HelloHTTP is an HTTP Cloud Function with a request parameter.
func HelloHTTP(_ http.ResponseWriter, _ *http.Request) {

	// Load the London time zone.
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		panic("Could not load time in the location. HelloHTTP Entry. Error: " + err.Error())
	}

	switch time.Now().In(loc).Hour() {
	case 8:
		Euromillions()
	case 9:
		PickMyPostcode()
	case 11:
		FreeBirthdateLottery()
	case 16:
		WinADinner()
	case 18:
		PickMyPostcode()
	case 21:
		PickMyPostcode()
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
