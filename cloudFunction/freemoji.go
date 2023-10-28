package cloudFunction

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/stealth"
)

type freemojiPerson struct {
	Name       string
	Email      string
	Password   string
	Entry      string
	MatchMain  bool
	MatchFiver bool
	MatchAny   bool
}

type freemojiTickets struct {
	Main   string
	Fivers []string
}

func freemoji() {
	// Create all clients.
	people := []freemojiPerson{
		{Name: "Andrew", Email: "andrew_field@hotmail.co.uk", Password: "DgZshM95&6zdNw", Entry: "🦅🌌🏝🎼🐧"},
		{Name: "Dad              ", Email: "mikefield@emfield.net", Password: "kc!aIkt^HCAp", Entry: "🍡🍌🌔🕸🐃"},
		{Name: "David            ", Email: "david.jafield@gmail.com", Password: "AEE3NRCOhCns", Entry: "👣👕🎅👏👹"},
		{Name: "James          ", Email: "j_field@hotmail.co.uk", Password: "lABxTUk4UKqF", Entry: "🐍🕶👔🕵😼"},
		{Name: "Katherine", Email: "k_avery@outlook.com", Password: "g$H!fWMk7@hu", Entry: "🌼🐿🐻👘👚"},
		{Name: "Eric                ", Email: "twintree47@pm.me", Password: "gWqktKmmOsWg", Entry: "🐸🌳🍓🦊😯"},
		{Name: "Nathan         ", Email: "budn8@hotmail.com", Password: "NYfxoKaY8YMR", Entry: "👻🐹🐁🍓♥"},
	}

	to := "andrew_field+freemojisummary@hotmail.co.uk"

	// Create browser
	browser := rod.New().MustConnect().Trace(true).Timeout(time.Second * 180) // -rod="show,trace,slow=1s,monitor=:1234"

	// browser.ServeMonitor("0.0.0.0:1234") // Open a browser and navigate to this address.

	// An effort to avoid bot detection.
	page := stealth.MustPage(browser)

	winningTickets := freemojiTickets{Main: "", Fivers: make([]string, 5)}

	err := rod.Try(func() {
		getFiverWinningTickets(page, &winningTickets)

		// Cycle through the people so each person gets a login. Otherwise, their entry may be disabled if they have not logged in for a while.
		getMainWinningTicket(page, &winningTickets, people[time.Now().Day()%len(people)])
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
		people[i].MatchMain = winningTickets.Main == people[i].Entry
		if slices.Contains(winningTickets.Fivers, people[i].Entry) {
			people[i].MatchFiver = true // Don't break early in case of multiple winners.
		}

		people[i].MatchAny = people[i].MatchMain || people[i].MatchFiver
		if people[i].MatchAny {
			result = true // Don't break early in case of multiple wins.
		}
	}

	// Get overall WIN/LOSE.
	summary := " - Freemoji summary."
	if result {
		summary = "WIN!" + summary
	} else {
		summary = "Lose" + summary
	}

	// Generate message.
	body := fmt.Sprintf(freemojiFormatResults(people) + "\n\n" + freemojiFormatTickets(winningTickets))

	// Send email.
	sendEmail(to, summary, body, nil)
}

func getFiverWinningTickets(page *rod.Page, ticketsToday *freemojiTickets) {
	page.MustNavigate("https://freemojilottery.com/fivers/")

	selectorString := "#fiversDraw > div:nth-child(%d) > div > div.result-combo > div > div:nth-child(%d) > div > object"

	for i := 0; i < 5; i++ {
		ticketsToday.Fivers[i] = ""
		for j := 0; j < 5; j++ {
			ticketsToday.Fivers[i] = ticketsToday.Fivers[i] + page.MustElement(fmt.Sprintf(selectorString, i+1, j+1)).MustProperty("standby").String()
		}
	}
}

func getMainWinningTicket(page *rod.Page, ticketsToday *freemojiTickets, client freemojiPerson) {
	page.MustNavigate("https://freemojilottery.com/")

	// Login
	page.MustElement("body > div.main-container > div.main-wrapper.header > div > header > div.col-xs-7.col-md-3.col-md-push-6 > a").MustClick()
	page.MustElement("#rnlr_user_login").MustInput(client.Email)
	page.MustElement("#rnlr_user_pass").MustInput(client.Password)
	page.MustElement("#rnlr_login_submit").MustClick()

	// Deny all cookies etc.
	page.MustSearch("#qc-cmp2-ui > div.qc-cmp2-footer.qc-cmp2-footer-overlay.qc-cmp2-footer-scrolled > div > button.css-1hy2vtq").MustClick()
	page.MustSearch("#qc-cmp2-ui > div.qc-cmp2-consent-info > div > div.qc-cmp2-header-links > button:nth-child(1)").MustClick()
	page.MustSearch("#qc-cmp2-ui > div.qc-cmp2-footer > div.qc-cmp2-footer-links > button:nth-child(2)").MustClick()
	page.MustSearch("#qc-cmp2-ui > div.qc-cmp2-footer > div.qc-cmp2-buttons-desktop > button").MustClick()

	// Get main ticket
	selectorString := "body > div.main-container > div:nth-child(4) > div > div.section-intro > div > div > div.col-xs-12.col-sm-10.col-sm-push-1.col-md-8.col-md-push-2.col-lg-6.col-lg-push-3.signup > div > div > div > div.freemoji-display-name.clearfix > div:nth-child(%d) > div > object"
	for i := 1; i < 6; i++ {
		ticketsToday.Main = ticketsToday.Main + page.MustElement(fmt.Sprintf(selectorString, i)).MustProperty("standby").String()
	}
}

func freemojiFormatResults(people []freemojiPerson) string {
	output := "Matches        Main    Fiver      Any\n"
	for _, p := range people {
		output += fmt.Sprintf("%-15s%-10t%-11t%-13t\n", p.Name, p.MatchMain, p.MatchFiver, p.MatchAny)
	}
	return output
}

func freemojiFormatTickets(ticketsToday freemojiTickets) string {
	output := "Tickets          Main                 Fivers\n"
	output += fmt.Sprintf("                %-13s%-13s\n", ticketsToday.Main, ticketsToday.Fivers[0])
	for _, ticket := range ticketsToday.Fivers[1:] {
		output += fmt.Sprintf("                                           %-14s\n", ticket)
	}

	return output
}
