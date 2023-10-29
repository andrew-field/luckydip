package cloudFunction

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/stealth"
)

type freeBirthdateLotteryPerson struct {
	Name     string
	Email    string
	Password string
	Entry    string
	Match    bool
}

func FreeBirthdateLottery() {
	// Create all clients.
	people := []freeBirthdateLotteryPerson{
		{Name: "Andrew", Email: "andrew_field@hotmail.co.uk", Password: "$Ha!bUdk#f3c7Y", Entry: "02/07/1994"},
		{Name: "Dad                  ", Email: "mikefield@emfield.net", Password: "xjx%@jVde&2nD*", Entry: "02/07/1954"},
		// {Name: "David    ", Email: "david.jafield@gmail.com", Password: "*7Dg6VW&K9m2R**", Entry: "23/12/1991"}, // Email currently unverified.
		{Name: "James          ", Email: "j_field@hotmail.co.uk", Password: "X4TwSDp$n8EU5z", Entry: "28/02/1988"},
		{Name: "Katherine", Email: "k_avery@outlook.com", Password: "FNZXf5ZMpWS$uv", Entry: "09/08/1985"},
	}

	to := "andrew_field+freebirthdaylottery@hotmail.co.uk"

	// Create browser
	browser := rod.New().MustConnect().Trace(true).Timeout(time.Second * 180) // -rod="show,trace,slow=1s,monitor=:1234"

	// browser.ServeMonitor("0.0.0.0:1234") // Open a browser and navigate to this address.

	// An effort to avoid bot detection.
	page := stealth.MustPage(browser)

	var winningTicket string

	err := rod.Try(func() {
		// Cycle through the people so each person gets a login. Otherwise, their entry may be disabled if they have not logged in for a while.
		winningTicket = freeBirthdateLotteryGetWinningTicket(page, people[time.Now().Day()%len(people)])
	})

	if err != nil {
		summary := "Unknown error"
		if errors.Is(err, context.DeadlineExceeded) {
			summary = "Timeout error"
		}

		sendEmail(to, summary, err.Error(), page.CancelTimeout().MustScreenshot())

		return
	}

	// Check for a winner.
	result := false
	for i := range people {
		if winningTicket == people[i].Entry {
			people[i].Match = true
			result = true // Don't break early for the slim chance there are multiple winners.
		}
	}

	// Get overall WIN/LOSE.
	summary := " - Free birthday lottery summary."
	if result {
		summary = "WIN!" + summary
	} else {
		summary = "Lose" + summary
	}

	// Generate message.
	body := fmt.Sprintf(freeBirthdateLotteryFormatResults(people) + "\n\n" + "Ticket: " + winningTicket)

	// Send email.
	sendEmail(to, summary, body, nil)
}

func freeBirthdateLotteryGetWinningTicket(page *rod.Page, client freeBirthdateLotteryPerson) string {
	page.MustNavigate("https://www.freebirthdatelottery.com/login/")

	// Login
	page.MustElement("#user_login").MustInput(client.Email)
	page.MustElement("#user_pass").MustInput(client.Password)
	page.MustElement("#wp-submit").MustClick()
	page.MustWaitDOMStable()

	// Get ticket
	page.MustNavigate("https://www.freebirthdatelottery.com/birthdate-draw/")
	fullText := page.MustElement("#post-13 > div > div.resultbox.fullwidthbox.checkresults > h1:nth-child(2)").MustText()

	return fullText[len(fullText)-10:]
}

func freeBirthdateLotteryFormatResults(people []freeBirthdateLotteryPerson) string {
	output := "Matches        Main\n"
	for _, p := range people {
		output += fmt.Sprintf("%-15s%-10t\n", p.Name, p.Match)
	}
	return output
}
