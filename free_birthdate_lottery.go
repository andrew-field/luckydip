package luckydip

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
)

type freeBirthdateLotteryPerson struct {
	Name               string
	Email              string
	Password           string
	Entry              string
	MatchBirthdateDraw bool
	MatchSurveyDraw    bool
	MatchAny           bool
}

type freeBirthdateLotteryTickets struct {
	BirthdateDraw string
	SurveyDraw    string
}

func FreeBirthdateLottery(page *rod.Page) {
	// Create all clients.
	people := []freeBirthdateLotteryPerson{
		{Name: "Andrew", Email: "andrew_field@hotmail.co.uk", Password: "$Ha!bUdk#f3c7Y", Entry: "02/07/1994"},
		{Name: "Dad               ", Email: "mikefield@emfield.net", Password: "xjx%@jVde&2nD*", Entry: "02/07/1954"},
		// {Name: "David    ", Email: "david.jafield@gmail.com", Password: "*7Dg6VW&K9m2R**", Entry: "23/12/1991"}, // Email currently unverified.
		{Name: "James            ", Email: "j_field@hotmail.co.uk", Password: "X4TwSDp$n8EU5z", Entry: "28/02/1988"},
		{Name: "Katherine", Email: "k_avery@outlook.com", Password: "FNZXf5ZMpWS$uv", Entry: "09/08/1985"},
	}

	winningTickets := freeBirthdateLotteryTickets{}

	err := rod.Try(func() {
		// Cycle through the people so each person gets a login. Otherwise, their entry may be disabled if they have not logged in for a while.
		winningTickets = freeBirthdateLotteryLoginAndGetWinningTicket(page, people[time.Now().Day()%len(people)])
	})

	to := "andrew_field+freebirthdatelottery@hotmail.co.uk"

	// If err is not nil, exit function.
	if checkProcessError(err, to, page) {
		return
	}

	// Check for a winner.
	result := false
	for i := range people {
		people[i].MatchBirthdateDraw = winningTickets.BirthdateDraw == people[i].Entry
		people[i].MatchSurveyDraw = winningTickets.SurveyDraw == people[i].Entry
		people[i].MatchAny = people[i].MatchBirthdateDraw || people[i].MatchSurveyDraw
		if people[i].MatchAny {
			result = true // Don't break early in case of multiple winners.
		}
	}

	// Get overall WIN/LOSE.
	outcome := LoseOutcome
	if result {
		outcome = WinOutcome
	}
	summary := outcome + " - Free birthdate lottery summary."

	// Generate message.
	body := fmt.Sprintf("%s\n\n%s", freeBirthdateLotteryFormatResults(people), freeBirthdateLotteryFormatTickets(winningTickets))

	// Send email.
	sendEmail(to, summary, body, nil)
}

func freeBirthdateLotteryLoginAndGetWinningTicket(page *rod.Page, client freeBirthdateLotteryPerson) freeBirthdateLotteryTickets {
	winningTickets := freeBirthdateLotteryTickets{}

	page.MustNavigate("https://www.freebirthdatelottery.com/login/")

	// Deny all cookies etc. Sometimes the cloud function does not show this box, so times out.
	cookiesBox, err := page.Timeout(time.Second * 7).Element("body > div.fc-consent-root")
	if err == nil {
		cookiesBox.MustRemove() // Remove the box in order to click the login button. Although the box will reappear I can still scrape the necessary data.
	}

	// Login.
	page.MustElement("#user_login").MustInput(client.Email)
	page.MustElement("#user_pass").MustInput(client.Password)
	page.MustElement("#wp-submit").MustClick()
	page.MustWaitDOMStable()

	// Get birthdate draw ticket.
	page.MustNavigate("https://www.freebirthdatelottery.com/birthdate-draw/")
	fullText := page.MustElementR("h2", "Winning Birthdate").MustText()
	winningTickets.BirthdateDraw = fullText[19:]

	// Get survey draw ticket.
	page.MustNavigate("https://www.freebirthdatelottery.com/survey-draw/")
	fullText = page.MustElementR("h2", "Winning Birthdate").MustText()
	winningTickets.SurveyDraw = fullText[19:]

	return winningTickets
}

func freeBirthdateLotteryFormatResults(people []freeBirthdateLotteryPerson) string {
	var output strings.Builder
	output.WriteString("Matches        Birthdate Draw        Survey Draw        Any       Entry\n")
	for _, p := range people {
		output.WriteString(fmt.Sprintf("%-15s%-30t%-25t%-10t%v\n", p.Name, p.MatchBirthdateDraw, p.MatchSurveyDraw, p.MatchAny, p.Entry))
	}
	return output.String()
}

func freeBirthdateLotteryFormatTickets(tickets freeBirthdateLotteryTickets) string {
	var output strings.Builder
	output.WriteString("Tickets       Birthdate Draw        Survey Draw\n")
	output.WriteString(fmt.Sprintf("%28s%23s\n", tickets.BirthdateDraw, tickets.SurveyDraw))

	return output.String()
}
