package cloudFunction

import (
	"net/http"
	"time"
)

// HelloHTTP is an HTTP Cloud Function with a request parameter.
func HelloHTTP(w http.ResponseWriter, r *http.Request) {
	loc, _ := time.LoadLocation("Europe/London")

	switch time.Now().In(loc).Hour() {
	case 6:
		euromillions()
	case 9:
		pickMyPostcode()
	case 11:
		freeBirthdateLottery()
	case 16:
		winADinner()
	case 18:
		pickMyPostcode()
	case 20:
		freemoji()
	case 21:
		pickMyPostcode()
	default:
		sendEmail("andrew_field@hotmail.co.uk", "Error", "Could not select the correct function. Time: "+time.Now().In(loc).String(), nil)
	}
}
