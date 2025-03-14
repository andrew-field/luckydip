package luckydip

import (
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/go-rod/rod"
)

type winADinnerPerson struct {
	Name     string
	Email    string
	Password string
	Entry    string
	Match    bool
}

func WinADinner(page *rod.Page) {
	// Create all clients.
	people := []winADinnerPerson{
		{Name: "Andrew", Email: "andrew_field@hotmail.co.uk", Password: "@8pMrqr8LXbaEq", Entry: "raddish5"},
		{Name: "Dad               ", Email: "mikefield@emfield.net", Password: "N8waUimb#SnDN7", Entry: "PorkChops12"},
		{Name: "David            ", Email: "david.jafield@gmail.com", Password: "#f#2GZjrodNSRX", Entry: "fancyeats"},
		{Name: "James            ", Email: "j_field@hotmail.co.uk", Password: "TB3sGkv62MCb4#", Entry: "lemonIsNice"},
		{Name: "Katherine", Email: "k_avery@outlook.com", Password: "Lsyjf5Lq*EuFh2", Entry: "YumYum8"},
	}

	var winningTickets []string

	err := rod.Try(func() {
		// Get the winning tickets.
		winningTickets = winADinnerGetWinningTickets(page)

		// Cycle through the people so each person gets a login. Otherwise, their entry may be disabled if they have not logged in for a while.
		winADinnerLogin(page, people[time.Now().Day()%len(people)])
	})

	to := "andrew_field+winadinner@hotmail.co.uk"

	// If err is not nil, exit function.
	if checkProcessError(err, to, page) {
		return
	}

	// Check for a winner.
	isWinner := false
	for i := range people {
		if slices.Contains(winningTickets, people[i].Entry) {
			people[i].Match = true
			isWinner = true // Don't break early for the slim chance there are multiple winners.
		}
	}

	// Get overall WIN/LOSE.
	outcome := LoseOutcome
	if isWinner {
		outcome = WinOutcome
	}
	summary := outcome + " - Get a dinner summary."

	// Generate message.
	body := fmt.Sprintf(winADinnerFormatResults(people)+"\n\nTickets: %v", winningTickets)

	// Send email.
	sendEmail(to, summary, body, nil)
}

func winADinnerGetWinningTickets(page *rod.Page) []string {
	page.MustNavigate("https://winadinner.com/daily-draw/")

	// Get tickets
	page.MustWaitElementsMoreThan("p.name", 2)
	winningElements := page.MustElements("p.name")
	winningTickets := make([]string, len(winningElements))
	for i, el := range winningElements {
		winningTickets[i] = el.MustText()
	}

	return winningTickets
}

func winADinnerLogin(page *rod.Page, clientToday winADinnerPerson) { // Already on a good page to login from the getWinningTickets method.
	// For some reason, this selector doesn't work with a '#' sign at the start.
	page.MustElement("body > header > div > div > ul > li:nth-child(2) > button").MustClick()

	page.MustElement("#user_name").MustInput(clientToday.Email)
	page.MustElement("#password").MustInput(clientToday.Password)

	page.MustElement("#sign-in-submit").MustClick()

	// Can't be bothered to log out after this. WaitDOMStable/Stable don't seem to work without a timeout. It does log out.
	err := page.Timeout(time.Second*10).WaitDOMStable(time.Second, 0)

	if err != nil {
		log.Println(fmt.Errorf("failed to wait for page dom stable after win a dinner login. Error: %w", err))
	}
}

func winADinnerFormatResults(people []winADinnerPerson) string {
	output := "Matches        Main      Entry\n"
	for _, p := range people {
		output += fmt.Sprintf("%-15s%-12t%v\n", p.Name, p.Match, p.Entry)
	}
	return output
}
