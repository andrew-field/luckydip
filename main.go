package main

import (
	"errors"
	"fmt"
	"net/smtp"
	"strings"
	"time"
	"unicode"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
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
	url := launcher.New().Set("headless", "new").MustLaunch()
	browser := rod.New().ControlURL(url).MustConnect()

	//browser.ServeMonitor("0.0.0.0:1234") // This can be coupled with -p 1234:1234 in the dockerfile. Open a browser and navigate to this address.

	// An effort to avoid bot detection.
	page := stealth.MustPage(browser)

	var errs error

	// Populate today's postcodes.
	postcodesToday := GetPostcodes(page, &people[0], &errs)

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

	// Login for each client and collect bonus.
	for i := range people {
		LoginAndGetBonus(page, &people[i], &errs)
	}

	// Get overall WIN/LOSE/ERROR.
	resultSummary := "LOSE"
	if result {
		resultSummary = "WIN!"
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

func GetPostcodes(page *rod.Page, client *person, errs *error) postcodes {
	page.MustNavigate("https://pickmypostcode.com")

	//Login
	page.MustElement("#v-rebrand > div.wrapper.top > div.wrapper--content.wrapper--content__relative > nav > ul > li.nav--buttons.nav--item > button.btn.btn-secondary.btn-cancel").MustClick()
	page.MustElement("#confirm-ticket").MustInput(client.Postcode)
	page.MustElement("#confirm-email").MustInput(client.Email)
	page.MustElement("#v-rebrand > div.wrapper.top > div.wrapper--content > main > div.overlay.overlay__open > section > div > div > div > form > button").MustClick()
	page.MustWaitStable()

	//Main draw
	page.MustElement("#gdpr-consent-tool-wrapper").Remove()
	el := page.MustElement("#main-draw-header > div > div > p.result--postcode")
	postcode, err := getPostcodeFromText(el.MustText())
	if err != nil {
		*errs = errors.Join(*errs, errors.New("Error while fetching the main postcode. "+err.Error()))
	}
	postcodesToday := postcodes{Main: postcode}

	page.MustNavigate("https://pickmypostcode.com/video/")
	page.MustElement("#gdpr-consent-tool-wrapper").Remove()
	el = page.MustElement("#result-header > div > p.result--postcode")
	postcode, err = getPostcodeFromText(el.MustText())
	if err != nil {
		*errs = errors.Join(*errs, errors.New("Error while fetching the video postcode. "+err.Error()))
	}
	postcodesToday.Video = postcode

	page.MustNavigate("https://pickmypostcode.com/survey-draw/")
	page.MustElement("#gdpr-consent-tool-wrapper").Remove()
	page.MustElement("#result-survey > div:nth-child(1) > div > div > div.survey-buttons > button.btn.btn-secondary").MustClick()
	el = page.MustElement("#result-header > div > p.result--postcode")
	postcode, err = getPostcodeFromText(el.MustText())
	if err != nil {
		*errs = errors.Join(*errs, errors.New("Error while fetching the survey postcode. "+err.Error()))
	}
	postcodesToday.Survey = postcode

	// For some reason these stackpot postcodes don't load properly all the time. Loop until they are loaded properly.
	page.MustNavigate("https://pickmypostcode.com/stackpot/").MustWaitLoad()
	page.MustElement("#gdpr-consent-tool-wrapper").Remove()
	time.Sleep(3 * time.Second) // Initial sleep to give unknown number of elements a chance to load
	runAgain := true
	for i := 0; runAgain; i++ {
		runAgain = false
		result := page.MustElements("p.result--postcode")
		for _, el := range result {
			postcode, err = getPostcodeFromText(el.MustText())
			if err != nil {
				if i > 10 {
					*errs = errors.Join(*errs, errors.New("Error while fetching the stackpot postcodes. Time out. "+err.Error()))
				}
				runAgain = true
				postcodesToday.Stackpot = make([]string, 0)
				time.Sleep(1 * time.Second)
				break
			}
			postcodesToday.Stackpot = append(postcodesToday.Stackpot, postcode)
		}
	}

	// For some reason these bonus postcodes don't load properly all the time. Loop until they are loaded properly.
	postcodesToday.Bonus = make([]string, 3)
	page.MustNavigate("https://pickmypostcode.com/your-bonus/").MustWaitLoad()
	page.MustElement("#gdpr-consent-tool-wrapper").Remove()
	for i := 0; ; i++ {
		el = page.MustElement("#banner-bonus > div > div.result-bonus.draw.draw-five > div > div.result--header > p")
		if postcode, err = getPostcodeFromText(el.MustText()); err == nil {
			postcodesToday.Bonus[0] = postcode
			break
		}
		if i > 10 {
			*errs = errors.Join(*errs, errors.New("Error while fetching the bonus 5 postcode. Time out. "+err.Error()))
			break
		}
		time.Sleep(1 * time.Second)
	}

	for i := 0; ; i++ {
		el = page.MustElement("#banner-bonus > div > div.result-bonus.draw.draw-ten > div > div.result--header > p")
		if postcode, err = getPostcodeFromText(el.MustText()); err == nil {
			postcodesToday.Bonus[1] = postcode
			break
		}
		if i > 10 {
			*errs = errors.Join(*errs, errors.New("Error while fetching the bonus 10 postcode. Time out. "+err.Error()))
			break
		}
		time.Sleep(1 * time.Second)
	}

	// for i := 0; ; i++ {
	// 	el = page.MustElement("#banner-bonus > div > div.result-bonus.draw.draw-twenty > div > div.result--header > p")
	// 	if postcode, err = getPostcodeFromText(el.MustText()); err == nil {
	//      postcodesToday.Bonus[2] = postcode
	// 		break
	// 	}
	// 	if i > 10 {
	// 		errs = errors.Join(errs, errors.New("Error while fetching the bonus 20 postcode. Time out. "+err.Error()))
	// 		break
	// 	}
	// 	time.Sleep(1 * time.Second)
	// }

	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		*errs = errors.Join(*errs, errors.New("Error loading location:"+err.Error()))
	} else {
		currentTime := time.Now().In(loc)

		// Check if it's after 18:00 in London
		if currentTime.Hour() >= 18 {
			for i := 0; ; i++ {
				el = page.MustElement("#fpl-minidraw > section > div > p.postcode")
				if postcode, err = getPostcodeFromText(el.MustText()); err == nil {
					postcodesToday.Minidraw = postcode
					break
				}
				if i > 10 {
					*errs = errors.Join(*errs, errors.New("Error while fetching the minidraw postcode. Time out. "+err.Error()))
					break
				}
				time.Sleep(1 * time.Second)
			}
		}
	}

	// One could populate the bonus money for the first client here. For sake of simplicity and organisation, do not.

	// Logout
	page.MustElement("#collapseMore > ul > li:nth-child(10) > a").MustClick().MustWaitStable()

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
		return "", errors.New("The postcode '" + s + "' is not valid.")
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

	page.MustElement("#gdpr-consent-tool-wrapper").Remove()

	//Login
	page.MustElement("#v-rebrand > div.wrapper.top > div.wrapper--content.wrapper--content__relative > nav > ul > li.nav--buttons.nav--item > button.btn.btn-secondary.btn-cancel").MustClick()
	page.MustElement("#confirm-ticket").MustInput(client.Postcode)
	page.MustElement("#confirm-email").MustInput(client.Email)
	page.MustElement("#v-rebrand > div.wrapper.top > div.wrapper--content > main > div.overlay.overlay__open > section > div > div > div > form > button").MustClick()
	page.MustWaitStable()

	page.MustNavigate("https://pickmypostcode.com/video/").MustWaitStable()
	page.MustNavigate("https://pickmypostcode.com/survey-draw/").MustWaitStable()
	page.MustElement("#gdpr-consent-tool-wrapper").Remove()

	// Populate bonus money. For some reason this doesn't always load properly. Loop until it does.
	var el *rod.Element
	for i := 0; ; i++ {
		el = page.MustElement("#v-main-header > div > div > a > p > span.tag.tag__xs.tag__success")
		if len(el.MustText()) < 10 {
			break
		}
		if i > 10 {
			*errs = errors.Join(*errs, errors.New("Error while fetching the bonus money for "+client.Name+". Time out. "+"Bonus text: "+el.MustText()))
			break
		}
		time.Sleep(1 * time.Second)
	}

	client.Bonus = el.MustText()

	// Logout
	page.MustElement("#collapseMore > ul > li:nth-child(10) > a").MustClick().MustWaitStable()
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
