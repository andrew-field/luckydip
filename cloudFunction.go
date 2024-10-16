package luckydip

import (
	"net/http"
	"time"
)

const UnknownError = "Unknown error"
const TimeoutError = "Timeout error"

// HelloHTTP is an HTTP Cloud Function with a request parameter.
func HelloHTTP(_ http.ResponseWriter, r *http.Request) {
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
	case 20:
		Freemoji()
	case 21:
		PickMyPostcode()
	default:
		sendEmail("andrew_field+luckdiperror@hotmail.co.uk", "Error in cloud function execution", "Could not select the correct function. Time: "+time.Now().In(loc).String(), nil)
	}
}
