package luckydip

import (
	"net/http"
	"time"
)

// HelloHTTP is an HTTP Cloud Function with a request parameter.
func HelloHTTP(w http.ResponseWriter, r *http.Request) {
	loc, _ := time.LoadLocation("Europe/London")

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
		sendEmail("Error in cloud function execution", "Could not select the correct function. Time: "+time.Now().In(loc).String(), nil)
	}
}
