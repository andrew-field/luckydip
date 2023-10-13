package cloudFunction

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/go-rod/rod"
	"github.com/go-rod/stealth"
)

type person struct {
	Name          string
	Email         string
	Postcode      string
	MatchMain     bool
	MatchVideo    bool
	MatchSurvey   bool
	MatchStackpot bool
	MatchBonus    bool
	MatchMinidraw bool
	MatchAny      bool
	BonusMoney    string
}

type tickets struct {
	Main     string
	Video    string
	Survey   string
	Stackpot []string
	Bonus    []string
	Minidraw string
}

// HelloHTTP is an HTTP Cloud Function with a request parameter.
func HelloHTTP(w http.ResponseWriter, r *http.Request) {
	// Create all clients.
	people := []person{
		{Name: "Andrew", Email: "andrew_field@hotmail.co.uk", Postcode: "gu113nt"},
		{Name: "Dad               ", Email: "mikefield@emfield.net", Postcode: "gu113gz"},
		{Name: "David            ", Email: "david.jafield@gmail.com", Postcode: "gu101dd"},
		{Name: "James            ", Email: "j_field@hotmail.co.uk", Postcode: "gu113nt"},
		{Name: "Katherine", Email: "k_avery@outlook.com", Postcode: "gu307tg"},
	}

	// Create browser
	browser := rod.New().MustConnect().Trace(true) // -rod="show,trace,slow=1s,monitor=:1234"

	// browser.ServeMonitor("0.0.0.0:1234") // Open a browser and navigate to this address.

	// An effort to avoid bot detection.
	page := stealth.MustPage(browser).Timeout(time.Second * 250)

	loc, _ := time.LoadLocation("Europe/London")
	isMainDraw := time.Now().In(loc).Hour() == 18

	var winningTickets tickets

	err := rod.Try(func() {
		// Populate today's winning postcodes. Also gets the bonus money for the first client while there.
		winningTickets = getWinningTickets(page, isMainDraw, &people[0])

		if isMainDraw {
			// Login for each client and collect bonus. Already collected bonus for the first client so skip the first.
			for i := 1; i < len(people); i++ {
				loginAndGetBonus(page, &people[i])
			}
		}
	})

	if err != nil {
		summary := "Unknown error"
		if errors.Is(err, context.DeadlineExceeded) {
			summary = "Timeout error"
		}

		err = sendEmail(summary, err.Error(), page.CancelTimeout().MustScreenshot())
		if err != nil {
			panic(err)
		}

		return
	}

	// Check for a winner.
	result := false
	if !isMainDraw {
		for i := range people {
			if slices.Contains(winningTickets.Stackpot, people[i].Postcode) {
				people[i].MatchStackpot = true
				result = true // Don't break early in case of multiple winners.
			}
		}
	} else {
		for i := range people {
			people[i].MatchMain = winningTickets.Main == people[i].Postcode
			people[i].MatchVideo = winningTickets.Video == people[i].Postcode
			people[i].MatchSurvey = winningTickets.Survey == people[i].Postcode
			if slices.Contains(winningTickets.Bonus, people[i].Postcode) {
				people[i].MatchBonus = true // Don't break early in case of multiple winners.
			}
			people[i].MatchMinidraw = winningTickets.Minidraw == people[i].Postcode
			people[i].MatchAny = people[i].MatchMain || people[i].MatchVideo || people[i].MatchSurvey || people[i].MatchBonus || people[i].MatchMinidraw
			if people[i].MatchAny {
				result = true // Don't break early in case of multiple winners.
			}
		}
	}

	// Get overall WIN/LOSE.
	summary := " - Pick my postcode summary."
	if isMainDraw {
		summary = " - Main draw" + summary
	} else {
		summary = " - Stackpot" + summary
	}
	if result {
		summary = "WIN!" + summary
	} else {
		summary = "Lose" + summary
	}

	// Generate message content.
	var results string
	var postcodes string
	if isMainDraw {
		results = formatResultsMainDraw(people)
		postcodes = formatPostcodesMainDraw(winningTickets)
	} else {
		results = formatResultsStackpot(people)
		postcodes = formatPostcodesStackpot(winningTickets)
	}
	body := fmt.Sprintf(results + "\n\n" + postcodes)

	// Send email.
	err = sendEmail(summary, body, nil)
	if err != nil {
		panic(err)
	}
}

func getWinningTickets(page *rod.Page, isMainDraw bool, client *person) tickets {
	page.MustNavigate("https://pickmypostcode.com")

	// Login
	login(page, client)

	// Deny all cookies etc.
	page.MustSearch("#denyAll").MustClick()
	page.MustSearch("button.mat-focus-indicator.okButton.mat-raised-button.mat-button-base").MustClick()
	page.MustWaitDOMStable()

	winningTickets := tickets{}
	var err error

	if !isMainDraw {
		// Stackpot
		page.MustNavigate("https://pickmypostcode.com/stackpot/")
		page.MustWaitDOMStable()
		page.MustWaitElementsMoreThan("p.result--postcode", 3)
		stackpotPostcodes := page.MustElements("p.result--postcode")
		winningTickets.Stackpot = make([]string, len(stackpotPostcodes))
		for i, el := range stackpotPostcodes {
			if winningTickets.Stackpot[i], err = getPostcodeFromText(el.MustText()); err != nil {
				panic(errors.Join(err, errors.New("error while fetching the stackpot postcodes")))
			}
		}

		return winningTickets
	}

	// Main draw
	el := page.MustElement("#main-draw-header > div > div > p.result--postcode")
	if winningTickets.Main, err = getPostcodeFromText(el.MustText()); err != nil {
		panic(errors.Join(err, errors.New("error while fetching the main postcode")))
	}

	// Video
	page.MustNavigate("https://pickmypostcode.com/video/")
	el = page.MustElement("#result-header > div > p.result--postcode")
	if winningTickets.Video, err = getPostcodeFromText(el.MustText()); err != nil {
		panic(errors.Join(err, errors.New("error while fetching the video postcode")))
	}

	// Survey draw
	page.MustNavigate("https://pickmypostcode.com/survey-draw/")
	button := page.MustElement("#result-survey > div:nth-child(1) > div > div > div.survey-buttons > button.btn.btn-secondary").MustScrollIntoView()
	page.Timeout(5 * time.Second).MustElement("#v-aside-rt").Remove() // Sometimes this content thing blocks the button
	button.MustClick()
	el = page.MustElement("#result-header > div > p.result--postcode")
	if winningTickets.Survey, err = getPostcodeFromText(el.MustText()); err != nil {
		panic(errors.Join(err, errors.New("error while fetching the survey postcode")))
	}

	// Bonus
	page.MustNavigate("https://pickmypostcode.com/your-bonus/")
	page.MustWaitDOMStable()
	page.MustWaitElementsMoreThan("p.result--postcode", 2) // 3 fails for some reason
	winningTickets.Bonus = make([]string, 3)
	el = page.MustElement("#banner-bonus > div > div.result-bonus.draw.draw-five > div > div.result--header > p")
	if winningTickets.Bonus[0], err = getPostcodeFromText(el.MustText()); err != nil {
		panic(errors.Join(err, errors.New("error while fetching the bonus 5 postcode")))
	}

	el = page.MustElement("#banner-bonus > div > div.result-bonus.draw.draw-ten > div > div.result--header > p")
	if winningTickets.Bonus[1], err = getPostcodeFromText(el.MustText()); err != nil {
		panic(errors.Join(err, errors.New("error while fetching the bonus 10 postcode")))
	}

	el = page.MustElement("#banner-bonus > div > div.result-bonus.draw.draw-twenty > div > div.result--header > p")
	if winningTickets.Bonus[2], err = getPostcodeFromText(el.MustText()); err != nil {
		panic(errors.Join(err, errors.New("error while fetching the bonus 20 postcode")))
	}

	// Minidraw
	page.MustElement("#fpl-minidraw > section > div > p.postcode").MustScrollIntoView()
	time.Sleep(9 * time.Second)
	el = page.MustElement("#fpl-minidraw > section > div > p.postcode")
	if winningTickets.Minidraw, err = getPostcodeFromText(el.MustText()); err != nil {
		panic(errors.Join(err, errors.New("error while fetching the minidraw postcode")))
	}

	// While here, to save time later, populate the bonus money for the first client here.
	populateTotalBonusMoneyForClient(page, client)

	// Logout
	page.MustElement("#collapseMore > ul > li:nth-child(10) > a").MustClick()

	return winningTickets
}

func login(page *rod.Page, client *person) {
	page.MustElement("#v-rebrand > div.wrapper.top > div.wrapper--content.wrapper--content__relative > nav > ul > li.nav--buttons.nav--item > button.btn.btn-secondary.btn-cancel").MustClick()
	page.MustElement("#confirm-ticket").MustInput(client.Postcode)
	page.MustElement("#confirm-email").MustInput(client.Email)
	page.MustElement("#v-rebrand > div.wrapper.top > div.wrapper--content > main > div.overlay.overlay__open > section > div > div > div > form > button").MustClick()
	page.MustWaitDOMStable()
}

func getPostcodeFromText(s string) (string, error) {
	// Get the first line of the string
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}

	// Remove all white spaces from the string
	s = strings.ReplaceAll(s, " ", "")

	// Changes to lower case
	s = strings.ToLower(s)

	// Checks if the postcode is valid.
	if !isValidPostcode(s) {
		return "", errors.New("The postcode: " + s + " , is not valid.")
	}

	return s, nil
}

func isValidPostcode(s string) bool {
	if len(s) < 5 || len(s) > 7 {
		return false
	}

	for _, c := range s {
		if !unicode.IsLetter(c) && !unicode.IsNumber(c) {
			return false
		}
	}

	return true
}

func loginAndGetBonus(page *rod.Page, client *person) {
	// Login.
	login(page, client)

	// Get bonus money for the account.
	page.MustNavigate("https://pickmypostcode.com/video/").MustWaitDOMStable()
	page.MustNavigate("https://pickmypostcode.com/survey-draw/").MustWaitDOMStable()

	populateTotalBonusMoneyForClient(page, client)

	// Logout.
	page.MustElement("#collapseMore > ul > li:nth-child(10) > a").MustClick()
}

func populateTotalBonusMoneyForClient(page *rod.Page, client *person) {
	if client.BonusMoney = page.MustElement("#v-main-header > div > div > a > p > span.tag.tag__xs.tag__success").MustText(); len(client.BonusMoney) > 10 {
		panic(errors.New("error while fetching the bonus money for " + client.Name + " Bonus text: " + client.BonusMoney))
	}
}

func formatResultsMainDraw(people []person) string {
	output := "Matches        Main    Video    Survey    Bonus     Minidraw    Any      Bonus Money\n"
	for _, p := range people {
		output += fmt.Sprintf("%-15s%-10t%-11t%-13t%-12t%-16t%-11t%-10s\n", p.Name, p.MatchMain, p.MatchVideo, p.MatchSurvey, p.MatchBonus, p.MatchMinidraw, p.MatchAny, p.BonusMoney)
	}
	return output
}

func formatResultsStackpot(people []person) string {
	output := "Matches        Stackpot\n"
	for _, p := range people {
		output += fmt.Sprintf("%-15s%-15t\n", p.Name, p.MatchStackpot)
	}
	return output
}

func formatPostcodesMainDraw(winningTickets tickets) string {
	output := "Postcodes     Main             Video           Survey       Bonus          Minidraw\n"
	output += fmt.Sprintf("                     %-14s%-14s%-14s%-14s%-14s\n", winningTickets.Main, winningTickets.Video, winningTickets.Survey, winningTickets.Bonus[0], winningTickets.Minidraw)
	output += fmt.Sprintf("%85s\n", winningTickets.Bonus[1])
	output += fmt.Sprintf("%85s\n", winningTickets.Bonus[2])

	return output
}

func formatPostcodesStackpot(winningTickets tickets) string {
	output := "Postcodes     Stackpot\n"
	for _, postcode := range winningTickets.Stackpot {
		output += fmt.Sprintf("%30s\n", postcode)
	}

	return output
}

func sendEmail(subject, body string, pic []byte) error {
	to := "andrew_field+pickmypostcodesummary@hotmail.co.uk"
	from := "andrewpcfield@gmail.com"

	// Set up authentication information.
	auth := smtp.PlainAuth("", from, os.Getenv("GOOGLEAPPPASSWORD"), "smtp.gmail.com")

	// Message
	msg := bytes.NewBuffer(nil)
	msg.WriteString("From: " + from + "\n")
	msg.WriteString("To: " + to + "\n")
	msg.WriteString("Subject: " + subject + "\n")

	if pic != nil { // There was an error
		msg.WriteString("MIME-Version: 1.0\n")
		msg.WriteString(`Content-Type: multipart/related; boundary="myboundary"` + "\n\n")
		msg.WriteString("--myboundary\n")

		// This is the body
		msg.WriteString(`Content-Type: text/plain; charset="utf-8"` + "\n")
		msg.WriteString("Content-Transfer-Encoding: quoted-printable" + "\n\n")
		msg.WriteString(body + "\n\n")
		msg.WriteString("--myboundary\n")

		// This is the attachment
		encodedImage := base64.StdEncoding.EncodeToString(pic)
		msg.WriteString(`Content-Type: image/jpeg;name="image.jpg"` + "\n")
		msg.WriteString("Content-Transfer-Encoding: base64" + "\n")
		msg.WriteString("Content-Disposition: attachment;filename=\"image.jpg\"" + "\n\n")
		msg.WriteString(encodedImage + "\n\n")
		msg.WriteString("--myboundary--")
	} else {
		msg.WriteString(body + "\n\n")
	}

	// Send the message.
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, msg.Bytes())
	if err != nil {
		return err
	}
	return nil
}
