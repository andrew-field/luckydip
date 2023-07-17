package main

import (
	"errors"
	"fmt"
	"net/smtp"
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
	MatchAny      bool
	MatchMinidraw bool
	Bonus         string
}

type postcodes struct {
	Main     string
	Video    string
	Survey   string
	Stackpot []string
	Bonus    []string
	Minidraw string
}

func main() {
	// Create all clients.
	people := []person{
		{Name: "Andrew", Email: "andrew_field+pickmypostcode.com@hotmail.co.uk", Postcode: "gu113nt"},
		{Name: "Dad               ", Email: "mikefield@emfield.net", Postcode: "gu113gz"},
		{Name: "David            ", Email: "david.jafield@gmail.com", Postcode: "gu101dd"},
		{Name: "James            ", Email: "j_field@hotmail.co.uk", Postcode: "gu113nt"},
		{Name: "Katherine", Email: "k_avery@outlook.com", Postcode: "gu307tg"},
	}

	// Create browser
	browser := rod.New().MustConnect() // -rod="show,trace,slow=1s,monitor=:1234"

	// browser.ServeMonitor("0.0.0.0:1234") // This can be coupled with -p 1234:1234 in the dockerfile. Open a browser and navigate to this address.

	// An effort to avoid bot detection.
	page := stealth.MustPage(browser)

	var errs error

	loc, _ := time.LoadLocation("Europe/London")
	isMainDraw := time.Now().In(loc).Hour() == 18

	// Populate today's postcodes.
	postcodesToday := GetPostcodes(page, isMainDraw, &people[0], &errs)

	// See if any postcodes match.
	result := false
	for i := range people {
		people[i].MatchMain = postcodesToday.Main == people[i].Postcode
		people[i].MatchVideo = postcodesToday.Video == people[i].Postcode
		people[i].MatchSurvey = postcodesToday.Survey == people[i].Postcode
		for _, stackpotPostcode := range postcodesToday.Stackpot {
			people[i].MatchStackpot = people[i].MatchStackpot || stackpotPostcode == people[i].Postcode
		}
		for _, bonusPostcode := range postcodesToday.Bonus {
			people[i].MatchBonus = people[i].MatchBonus || bonusPostcode == people[i].Postcode
		}
		people[i].MatchMinidraw = postcodesToday.Minidraw == people[i].Postcode
		people[i].MatchAny = people[i].MatchMain || people[i].MatchVideo || people[i].MatchSurvey || people[i].MatchStackpot || people[i].MatchBonus || people[i].MatchMinidraw
		result = result || people[i].MatchAny
	}

	if isMainDraw {
		// Login for each client and collect bonus.
		for i := range people {
			LoginAndGetBonus(page, &people[i], &errs)
		}
	}

	// Get overall WIN/LOSE/ERROR.
	resultSummary := "Stackpot -"
	if isMainDraw {
		resultSummary = "Main draw -"
	}
	if result {
		resultSummary += " WIN!"
	} else {
		resultSummary += " LOSE"
	}
	if errs != nil {
		resultSummary = "Error - " + resultSummary
	}

	body := fmt.Sprintf(formatResults(people) + "\n\n" + formatPostcodes(postcodesToday))
	if errs != nil {
		body += fmt.Sprintf("\n\nErrors:\n" + errs.Error())
	}

	err := sendEmail("andrew_field+pickmypostcodesummary@hotmail.co.uk", resultSummary+" - Pick my postcode summary", body)
	if err != nil {
		panic(err)
	}
}

func GetPostcodes(page *rod.Page, isMainDraw bool, client *person, errs *error) postcodes {
	page.MustNavigate("https://pickmypostcode.com")

	// Login
	page.MustElement("#v-rebrand > div.wrapper.top > div.wrapper--content.wrapper--content__relative > nav > ul > li.nav--buttons.nav--item > button.btn.btn-secondary.btn-cancel").MustClick()
	page.MustElement("#confirm-ticket").MustInput(client.Postcode)
	page.MustElement("#confirm-email").MustInput(client.Email)
	page.MustElement("#v-rebrand > div.wrapper.top > div.wrapper--content > main > div.overlay.overlay__open > section > div > div > div > form > button").MustClick()

	// Deny all cookies etc.
	page.MustSearch("#denyAll").MustClick()
	page.MustSearch("button.mat-focus-indicator.okButton.mat-raised-button.mat-button-base").MustClick()
	page.MustWaitStable()

	postcodesToday := postcodes{}
	postcodesToday.Bonus = make([]string, 3)
	var err error

	if !isMainDraw {
		// Stackpot
		page.MustNavigate("https://pickmypostcode.com/stackpot/")
		page.MustWaitStable()
		page.MustWaitElementsMoreThan("p.result--postcode", 3)
		stackpotPostcodes := page.MustElements("p.result--postcode")
		postcodesToday.Stackpot = make([]string, len(stackpotPostcodes))
		for i, el := range stackpotPostcodes {
			if postcodesToday.Stackpot[i], err = getPostcodeFromText(el.MustText()); err != nil {
				*errs = errors.Join(*errs, errors.New("Error while fetching the stackpot postcodes. "+err.Error()))
			}
		}

		return postcodesToday
	}

	// Main draw
	el := page.MustElement("#main-draw-header > div > div > p.result--postcode")
	if postcodesToday.Main, err = getPostcodeFromText(el.MustText()); err != nil {
		*errs = errors.Join(*errs, errors.New("Error while fetching the main postcode. "+err.Error()))
	}

	// Video
	page.MustNavigate("https://pickmypostcode.com/video/")
	el = page.MustElement("#result-header > div > p.result--postcode")
	if postcodesToday.Video, err = getPostcodeFromText(el.MustText()); err != nil {
		*errs = errors.Join(*errs, errors.New("Error while fetching the video postcode. "+err.Error()))
	}

	// Survey draw
	page.MustNavigate("https://pickmypostcode.com/survey-draw/")
	page.MustElement("#result-survey > div:nth-child(1) > div > div > div.survey-buttons > button.btn.btn-secondary").MustClick()
	el = page.MustElement("#result-header > div > p.result--postcode")
	if postcodesToday.Survey, err = getPostcodeFromText(el.MustText()); err != nil {
		*errs = errors.Join(*errs, errors.New("Error while fetching the survey postcode. "+err.Error()))
	}

	// Bonus
	page.MustNavigate("https://pickmypostcode.com/your-bonus/")
	page.MustWaitStable()
	page.MustWaitElementsMoreThan("p.result--postcode", 2)
	el = page.MustElement("#banner-bonus > div > div.result-bonus.draw.draw-five > div > div.result--header > p")
	if postcodesToday.Bonus[0], err = getPostcodeFromText(el.MustText()); err != nil {
		*errs = errors.Join(*errs, errors.New("Error while fetching the bonus 5 postcode. "+err.Error()))
	}

	el = page.MustElement("#banner-bonus > div > div.result-bonus.draw.draw-ten > div > div.result--header > p")
	if postcodesToday.Bonus[1], err = getPostcodeFromText(el.MustText()); err != nil {
		*errs = errors.Join(*errs, errors.New("Error while fetching the bonus 10 postcode. "+err.Error()))
	}

	el = page.MustElement("#banner-bonus > div > div.result-bonus.draw.draw-twenty > div > div.result--header > p")
	if postcodesToday.Bonus[2], err = getPostcodeFromText(el.MustText()); err != nil {
		*errs = errors.Join(*errs, errors.New("Error while fetching the bonus 20 postcode. "+err.Error()))
	}

	// Minidraw
	page.MustElement("#fpl-minidraw > section > div").MustScrollIntoView()
	time.Sleep(10 * time.Second)
	el = page.MustElement("#fpl-minidraw > section > div > p.postcode")
	if postcodesToday.Minidraw, err = getPostcodeFromText(el.MustText()); err != nil {
		*errs = errors.Join(*errs, errors.New("Error while fetching the minidraw postcode. "+err.Error()))
	}

	// One could populate the bonus money for the first client here. For sake of simplicity and organisation, do not.

	// Logout
	page.MustElement("#collapseMore > ul > li:nth-child(10) > a").MustClick()
	page.MustWaitStable()

	return postcodesToday
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

func LoginAndGetBonus(page *rod.Page, client *person, errs *error) {
	page.MustNavigate("https://pickmypostcode.com")

	// Login
	page.MustElement("#v-rebrand > div.wrapper.top > div.wrapper--content.wrapper--content__relative > nav > ul > li.nav--buttons.nav--item > button.btn.btn-secondary.btn-cancel").MustClick()
	page.MustElement("#confirm-ticket").MustInput(client.Postcode)
	page.MustElement("#confirm-email").MustInput(client.Email)
	page.MustElement("#v-rebrand > div.wrapper.top > div.wrapper--content > main > div.overlay.overlay__open > section > div > div > div > form > button").MustClick()
	page.MustWaitStable()

	// Get bonus
	page.MustNavigate("https://pickmypostcode.com/video/").MustWaitStable()
	page.MustNavigate("https://pickmypostcode.com/survey-draw/").MustWaitStable()

	// Get total bonus money
	el := page.MustElement("#v-main-header > div > div > a > p > span.tag.tag__xs.tag__success")
	if client.Bonus = el.MustText(); len(client.Bonus) > 10 {
		*errs = errors.Join(*errs, errors.New("Error while fetching the bonus money for "+client.Name+" Bonus text: "+el.MustText()))
	}

	// Logout
	page.MustElement("#collapseMore > ul > li:nth-child(10) > a").MustClick()
	page.MustWaitStable()
}

func formatResults(people []person) string {
	output := "Matches        Main    Video    Survey    Stackpot    Bonus    Minidraw    Any      Bonus Money\n"
	for _, p := range people {
		output += fmt.Sprintf("%-15s%-10t%-11t%-13t%-15t%-12t%-16t%-9t%-10s\n", p.Name, p.MatchMain, p.MatchVideo, p.MatchSurvey, p.MatchStackpot, p.MatchBonus, p.MatchMinidraw, p.MatchAny, p.Bonus)
	}
	return output
}

func formatPostcodes(postcodesToday postcodes) string {
	output := "Postcodes     Main             Video           Survey         Stackpot       Bonus          Minidraw\n"
	output += fmt.Sprintf("                     %-14s%-14s%-14s%-14s%-14s%-14s\n", postcodesToday.Main, postcodesToday.Video, postcodesToday.Survey, postcodesToday.Stackpot[0], postcodesToday.Bonus[0], postcodesToday.Minidraw)
	output += fmt.Sprintf("                                                                                  %-14s%-14s\n", postcodesToday.Stackpot[1], postcodesToday.Bonus[1])
	output += fmt.Sprintf("                                                                                  %-14s%-14s\n", postcodesToday.Stackpot[2], postcodesToday.Bonus[2])
	for _, postcode := range postcodesToday.Stackpot[3:] {
		output += fmt.Sprintf("                                                                                  %-14s\n", postcode)
	}

	return output
}

func sendEmail(to, subject, body string) error {
	// Set up authentication information.
	auth := smtp.PlainAuth("", "andrewpcfield@gmail.com", "", "smtp.gmail.com") // Need to remove app password from code.

	// Compose the message.
	msg := "To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body

	// Send the message.
	err := smtp.SendMail("smtp.gmail.com:587", auth, "andrewpcfield@gmail.com", []string{to}, []byte(msg))
	if err != nil {
		return err
	}
	return nil
}
