package luckydip

import (
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/go-rod/rod"
)

type euromillionsPerson struct {
	Name        string
	Email       string
	Password    string
	Entry       []string // Same entry for daily and weekly draw
	MatchDaily  bool
	MatchWeekly bool
	MatchAny    bool
}

type euromillionsTickets struct {
	Daily  []string
	Weekly []string
}

func Euromillions(page *rod.Page) {
	// Create all clients.
	people := []euromillionsPerson{
		{Name: "Andrew", Email: "andrew_field@hotmail.co.uk", Password: "4Rdod7o!T6Hjyp", Entry: []string{"6", "11", "18", "19", "25", "32"}},
		{Name: "Dad               ", Email: "mikefield@emfield.net", Password: "Hzu#%m9VTx9Gty", Entry: []string{"2", "6", "14", "18", "24", "30"}},
		{Name: "David            ", Email: "david.jafield@gmail.com", Password: "fWT5r7eG8k@d@h", Entry: []string{"1", "7", "12", "17", "31", "32"}},
		{Name: "James            ", Email: "j_field@hotmail.co.uk", Password: "A8&2WikQbA47a!", Entry: []string{"4", "5", "15", "24", "25", "26"}},
		{Name: "Katherine", Email: "k_avery@outlook.com", Password: "T$tyfRx5&qkaoi", Entry: []string{"9", "13", "19", "23", "27", "28"}},
	}

	var winningTickets euromillionsTickets

	err := rod.Try(func() {
		// Reject cookies etc. one time. Popup takes a moment or two to show. For some reason, doesn't show with cloud function, so include a possible timeout.
		page.MustNavigate("https://www.euro-millions.com/account/login").MustWaitDOMStable()
		if el, elError := page.Timeout(time.Second * 5).Element("body > div.fc-consent-root > div.fc-dialog-container > div.fc-dialog.fc-choice-dialog > div.fc-footer-buttons-container > div.fc-footer-buttons > button.fc-button.fc-cta-manage-options.fc-secondary-button"); elError == nil {
			el.MustClick()
			page.MustElement("body > div.fc-consent-root > div.fc-dialog-container > div.fc-dialog.fc-data-preferences-dialog > div.fc-dialog-content > div > div.fc-preferences-container > div:nth-child(3) > label.fc-preference-slider-container.fc-legitimate-interest-preference-container > span.fc-preference-slider").MustClick()
			page.MustElement("body > div.fc-consent-root > div.fc-dialog-container > div.fc-dialog.fc-data-preferences-dialog > div.fc-dialog-content > div > div.fc-preferences-container > div:nth-child(8) > label.fc-preference-slider-container.fc-legitimate-interest-preference-container > span.fc-preference-slider").MustClick()
			page.MustElement("body > div.fc-consent-root > div.fc-dialog-container > div.fc-dialog.fc-data-preferences-dialog > div.fc-dialog-content > div > div.fc-preferences-container > div:nth-child(9) > label.fc-preference-slider-container.fc-legitimate-interest-preference-container > span.fc-preference-slider").MustClick()
			page.MustElement("body > div.fc-consent-root > div.fc-dialog-container > div.fc-dialog.fc-data-preferences-dialog > div.fc-dialog-content > div > div.fc-preferences-container > div:nth-child(10) > label.fc-preference-slider-container.fc-legitimate-interest-preference-container > span.fc-preference-slider").MustClick()
			page.MustElement("body > div.fc-consent-root > div.fc-dialog-container > div.fc-dialog.fc-data-preferences-dialog > div.fc-dialog-content > div > div.fc-preferences-container > div:nth-child(11) > label.fc-preference-slider-container.fc-legitimate-interest-preference-container > span.fc-preference-slider").MustClick()
			page.MustElement("body > div.fc-consent-root > div.fc-dialog-container > div.fc-dialog.fc-data-preferences-dialog > div.fc-dialog-content > div > div.fc-preferences-container > div:nth-child(12) > label.fc-preference-slider-container.fc-legitimate-interest-preference-container > span.fc-preference-slider").MustClick()
			page.MustElement("body > div.fc-consent-root > div.fc-dialog-container > div.fc-dialog.fc-data-preferences-dialog > div.fc-footer-buttons-container > div.fc-footer-buttons > button.fc-button.fc-confirm-choices.fc-primary-button").MustClick()
		}

		// Get today's winning tickets.
		winningTickets = euromillionsGetWinningTickets(page)

		// Login for each client and enter draw.
		for i := range people {
			euromillionsLogin(page, people[i])
		}
	})

	to := "andrew_field+euromillions@hotmail.co.uk"

	// If err is not nil, exit function.
	if checkProcessError(err, to, page) {
		return
	}

	// Check for a winner.
	result := false
	for i := range people {
		people[i].MatchDaily = slices.Equal(winningTickets.Daily, people[i].Entry)
		if winningTickets.Weekly != nil {
			people[i].MatchWeekly = slices.Equal(winningTickets.Weekly, people[i].Entry)
		}
		people[i].MatchAny = people[i].MatchDaily || people[i].MatchWeekly
		if people[i].MatchAny {
			result = true // Don't break early in case of multiple winners.
		}
	}

	// Get overall WIN/LOSE.
	outcome := LoseOutcome
	if result {
		outcome = WinOutcome
	}
	summary := outcome + " - Euromillions summary."

	// Generate message.
	body := fmt.Sprintf("%s\n\n%s", euromillionsFormatResults(people), euromillionsFormatTickets(winningTickets))

	// Send email.
	sendEmail(to, summary, body, nil)
}

func euromillionsGetWinningTickets(page *rod.Page) euromillionsTickets {
	page.MustNavigate("https://www.euro-millions.com/free-lottery/results")

	selectorString := "#resultsTable > tbody > tr:nth-child(2) > td:nth-child(3) > ul > li:nth-child(%d)"

	// Get the daily winning ticket.
	winningTickets := euromillionsTickets{Daily: make([]string, 6)}
	for i := 0; i < 6; i++ {
		winningTickets.Daily[i] += page.MustElement(fmt.Sprintf(selectorString, i+1)).MustText()
	}

	// Check for weekly draw.
	if time.Now().Weekday() == time.Tuesday {
		selectorString = "#resultsTable > tbody > tr:nth-child(3) > td:nth-child(3) > ul > li:nth-child(%d)"

		winningTickets.Weekly = make([]string, 6)
		for i := 0; i < 6; i++ {
			winningTickets.Weekly[i] += page.MustElement(fmt.Sprintf(selectorString, i+1)).MustText()
		}
	}

	return winningTickets
}

func euromillionsLogin(page *rod.Page, client euromillionsPerson) {
	page.MustNavigate("https://www.euro-millions.com/account/login")

	// Login.
	page.MustElement("#Email").MustInput(client.Email)
	page.MustElement("#Password").MustInput(client.Password)
	page.MustElement("#Submit").MustClick()
	err := page.Timeout(time.Second * 7).WaitStable(time.Second) // WaitStable can sometimes take ~15+ seconds which adds up and may cause timeout.
	if err != nil {
		log.Println("Failed to wait for page stable after EuroMillion login. Error:", err.Error())
	}

	// Enter daily draw.
	page.MustNavigate("https://www.euro-millions.com/free-lottery/play?lottery=daily")
	enterDraw(page, client)

	// Enter weekly draw.
	if time.Now().Weekday() == time.Thursday { // Has to be Thursday because it should be played as soon as possible. Apparently, if the weekly can be played, the daily can not be played until the weekly has been played.
		page.MustNavigate("https://www.euro-millions.com/free-lottery/play?lottery=weekly")
		enterDraw(page, client)
	}

	// Logout.
	page.MustElement("body > header > div > div.fx.acen > a").MustClick()
	err = page.Timeout(time.Second * 7).WaitStable(time.Second)
	if err != nil {
		log.Println("Failed to wait for page stable after EuroMillion log out.", err.Error())
	}
}

func enterDraw(page *rod.Page, client euromillionsPerson) {
	page.MustElement("#reset_ticket").MustClick() // Selected numbers sometimes need resetting.
	for _, v := range client.Entry {              // Even when going from daily to weekly, sometimes the numbers need setting again.
		page.MustElement("#B0ID_" + v).MustWaitEnabled().MustClick() // Sometimes the buttons have not been enabled when clicking.
	}

	page.MustElement("#submit_ticket").MustClick()
	page.MustElement("#redirectTimer > a") // Wait for ticket to be processed.
}

func euromillionsFormatResults(people []euromillionsPerson) string {
	output := "Matches        Daily    Weekly      Any           Entry\n"
	for _, p := range people {
		// Format the output string with proper alignment.
		output += fmt.Sprintf(
			"%-15s%-10t%-15t%-15t%v\n", // Format string
			p.Name,                     // Name
			p.MatchDaily,               // MatchDaily
			p.MatchWeekly,              // MatchWeekly
			p.MatchAny,                 // MatchAny
			p.Entry,                    // Entry
		)
	}
	return output
}

func euromillionsFormatTickets(winningTickets euromillionsTickets) string {
	output := "Tickets          Daily                   Weekly\n"
	if winningTickets.Weekly != nil {
		output += fmt.Sprintf("                %v, %v\n", winningTickets.Daily, winningTickets.Weekly)
	} else {
		output += fmt.Sprintf("                %v, %s\n", winningTickets.Daily, "No weekly draw")
	}

	return output
}
