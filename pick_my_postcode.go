package luckydip

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/go-rod/rod"
	"github.com/go-rod/stealth"
)

type pickMyPostcodePerson struct {
	Name          string
	Email         string
	Entry         string
	MatchMain     bool
	MatchVideo    bool
	MatchSurvey   bool
	MatchStackpot bool
	MatchBonus    bool
	MatchMinidraw bool
	MatchAny      bool
	BonusMoney    string
}

type pickMyPostcodeTickets struct {
	Main     string
	Video    string
	Survey   string
	Stackpot []string
	Bonus    []string
	Minidraw string
}

func PickMyPostcode() {
	// Create all clients.
	people := []pickMyPostcodePerson{
		{Name: "Andrew", Email: "andrew_field@hotmail.co.uk", Entry: "gu113nt"},
		{Name: "Dad               ", Email: "mikefield@emfield.net", Entry: "gu113gz"},
		{Name: "David            ", Email: "david.jafield@gmail.com", Entry: "gu101dd"},
		{Name: "James            ", Email: "j_field@hotmail.co.uk", Entry: "gu113nt"},
		{Name: "Katherine", Email: "k_avery@outlook.com", Entry: "gu307da"},
	}

	// Create browser
	browser := rod.New().MustConnect().Trace(true).Timeout(time.Second * 180)

	// An effort to avoid bot detection.
	page := stealth.MustPage(browser)

	// Load the London timezone.
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		panic("Could not load time in the location. Pick my postcode entry. Error: " + err.Error())
	}

	// Is the time of execution suitable to check the main draw or the stack pot draw.
	isMainDraw := time.Now().In(loc).Hour() == 18

	var winningTickets pickMyPostcodeTickets

	err = rod.Try(func() {
		// Populate today's winning postcodes. Also gets the bonus money for the first client while there.
		winningTickets = pickMyPostcodeGetWinningTickets(page, isMainDraw, &people[0])

		if isMainDraw {
			// Login for each client and collect bonus. Already collected bonus for the first client so skip the first.
			for i := 1; i < len(people); i++ {
				loginAndGetBonus(page, &people[i])
			}
		}
	})

	to := "andrew_field+pickmypostcode@hotmail.co.uk"

	// If err is not nil, exit function.
	if checkProcessError(err, to, page) {
		return
	}

	// Check for a winner.
	result := checkForWinner(isMainDraw, people, winningTickets)

	summaryType := "Stackpot"
	if isMainDraw {
		summaryType = "Main draw"
	}

	outcome := LoseOutcome
	if result {
		outcome = WinOutcome
	}
	summary := fmt.Sprintf("%s - %s - Pick my postcode summary.", outcome, summaryType)

	// Generate message content.
	var results string
	var postcodes string
	if isMainDraw {
		results = formatMainDrawResults(people)
		postcodes = formatMainDrawPostcodes(winningTickets)
	} else {
		results = formatStackpotResults(people)
		postcodes = formatStackpotPostcodes(winningTickets)
	}
	body := fmt.Sprintf("%s\n\n%s", results, postcodes)

	// Send email.
	sendEmail(to, summary, body, nil)
}

func pickMyPostcodeGetWinningTickets(page *rod.Page, isMainDraw bool, client *pickMyPostcodePerson) pickMyPostcodeTickets {
	page.MustNavigate("https://pickmypostcode.com")

	// Login.
	login(page, client)

	// Deny all cookies etc. Sometimes the cloud function does not show this box, so times out.
	denyAllBox, err := page.Timeout(time.Second * 7).Element("#qc-cmp2-ui > div.qc-cmp2-footer.qc-cmp2-footer-overlay.qc-cmp2-footer-scrolled > div > button:nth-child(2)")
	if err == nil {
		denyAllBox.MustClick() // Deny all.
	}

	winningTickets := pickMyPostcodeTickets{}

	if !isMainDraw {
		// If not the main draw then it is the stack pot draw.
		page.MustNavigate("https://pickmypostcode.com/stackpot/")
		page.MustWaitDOMStable()
		page.MustWaitElementsMoreThan("p.result--postcode", 2)
		stackpotPostcodes := page.MustElements("p.result--postcode")
		winningTickets.Stackpot = make([]string, len(stackpotPostcodes))
		for i, el := range stackpotPostcodes {
			if winningTickets.Stackpot[i], err = getPostcodeFromText(el.MustText()); err != nil {
				panic("error while fetching the stackpot postcodes" + err.Error())
			}
		}

		return winningTickets
	}

	// Main draw.
	el := page.MustElement("#main-draw-header > div > div > p.result--postcode")
	if winningTickets.Main, err = getPostcodeFromText(el.MustText()); err != nil {
		panic("error while fetching the main postcode" + err.Error())
	}

	// Video.
	page.MustNavigate("https://pickmypostcode.com/video/")
	el = page.MustElement("#result-header > div > p.result--postcode")
	if winningTickets.Video, err = getPostcodeFromText(el.MustText()); err != nil {
		panic("error while fetching the video postcode" + err.Error())
	}

	// Survey draw.
	page.MustNavigate("https://pickmypostcode.com/survey-draw/")
	button := page.MustElement("#result-survey > div:nth-child(1) > div > div > div.survey-buttons > button.btn.btn-secondary").MustScrollIntoView()
	err = page.Timeout(5 * time.Second).MustElement("#v-aside-rt").Remove() // Sometimes this content thing blocks the button.
	if err != nil {
		log.Println("Failed to remove content at pick my postcode survey draw. Error:", err.Error())
	}
	button.MustClick()
	el = page.MustElement("#result-header > div > p.result--postcode")
	if winningTickets.Survey, err = getPostcodeFromText(el.MustText()); err != nil {
		panic("error while fetching the survey postcode" + err.Error())
	}

	// Bonus 5.
	page.MustNavigate("https://pickmypostcode.com/your-bonus/")
	page.MustWaitDOMStable()
	page.MustWaitElementsMoreThan("p.result--postcode", 2) // 3 fails for some reason.
	winningTickets.Bonus = make([]string, 3)
	el = page.MustElement("#banner-bonus > div > div.result-bonus.draw.draw-five > div > div.result--header > p")
	if winningTickets.Bonus[0], err = getPostcodeFromText(el.MustText()); err != nil {
		panic("error while fetching the bonus 5 postcode" + err.Error())
	}

	// Bonus 10.
	el = page.MustElement("#banner-bonus > div > div.result-bonus.draw.draw-ten > div > div.result--header > p")
	if winningTickets.Bonus[1], err = getPostcodeFromText(el.MustText()); err != nil {
		panic("error while fetching the bonus 10 postcode" + err.Error())
	}

	// Bonus 20.
	el = page.MustElement("#banner-bonus > div > div.result-bonus.draw.draw-twenty > div > div.result--header > p")
	if winningTickets.Bonus[2], err = getPostcodeFromText(el.MustText()); err != nil {
		panic("error while fetching the bonus 20 postcode" + err.Error())
	}

	// Minidraw.
	page.MustElement("#fpl-minidraw > section > div > p.postcode").MustScrollIntoView()
	time.Sleep(9 * time.Second)
	el = page.MustElement("#fpl-minidraw > section > div > p.postcode")
	if winningTickets.Minidraw, err = getPostcodeFromText(el.MustText()); err != nil {
		panic("error while fetching the minidraw postcode" + err.Error())
	}

	// While here, to save time later, populate the bonus money for the first client here.
	populateTotalBonusMoneyForClient(page, client)

	// Logout
	page.MustElement("#collapseMore > ul > li:nth-child(10) > a").MustClick()

	return winningTickets
}

func login(page *rod.Page, client *pickMyPostcodePerson) {
	page.MustElement("#v-rebrand > div.wrapper.top > div.wrapper--content.wrapper--content__relative > nav > ul > li.nav--buttons.nav--item > button.btn.btn-secondary.btn-cancel").MustClick()
	page.MustElement("#confirm-ticket").MustInput(client.Entry)
	page.MustElement("#confirm-email").MustInput(client.Email)
	page.MustElement("#v-rebrand > div.wrapper.top > div.wrapper--content > main > div.overlay.overlay__open > section > div > div > div > form > button").MustClick()
	if err := page.Timeout(time.Second*7).WaitDOMStable(time.Second, 0); err != nil {
		panic(fmt.Errorf("failed to wait for page dom stable after pick my postcode login: %w", err))
	}
}

func getPostcodeFromText(s string) (string, error) {
	// Extract the first line of the string to ensure we only process the postcode part.
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}

	// Remove all spaces to normalize the postcode format.
	s = strings.ReplaceAll(s, " ", "")

	// Converts the postcode to lowercase for uniformity.
	s = strings.ToLower(s)

	// Checks if the postcode is valid.
	if !isValidPostcode(s) {
		return "", errors.New("the postcode " + s + " is not valid")
	}

	// Return the cleaned and validated postcode.
	return s, nil
}

// A valid postcode must be between 5 to 7 characters long and contain only letters and numbers.
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

func loginAndGetBonus(page *rod.Page, client *pickMyPostcodePerson) {
	// Login.
	login(page, client)

	// Get bonus money for the account.
	page.MustNavigate("https://pickmypostcode.com/video/").MustWaitDOMStable()
	page.MustNavigate("https://pickmypostcode.com/survey-draw/").MustWaitDOMStable()

	populateTotalBonusMoneyForClient(page, client)

	// Logout.
	page.MustElement("#collapseMore > ul > li:nth-child(10) > a").MustClick()
}

func populateTotalBonusMoneyForClient(page *rod.Page, client *pickMyPostcodePerson) {
	bonusElement := page.MustElement("#v-main-header > div > div > a > p > span.tag.tag__xs.tag__success")
	// Sometimes there is a strange error and the return value text is not loaded properly and a long value. Check for this by checking if it is > 10, if so, run it again after some time.
	if client.BonusMoney = bonusElement.MustText(); len(client.BonusMoney) > 10 {
		log.Println("Error getting total bonus money for " + client.Name + ". Retrying after 5 seconds...")
		time.Sleep(time.Second * 5)
		bonusElement = page.MustElement("#v-main-header > div > div > a > p > span.tag.tag__xs.tag__success")
		if client.BonusMoney = bonusElement.MustText(); len(client.BonusMoney) > 10 {
			panic("error while fetching the bonus money for " + client.Name + " Bonus text: " + client.BonusMoney)
		}
	}
}

func checkForWinner(isMainDraw bool, people []pickMyPostcodePerson, winningTickets pickMyPostcodeTickets) bool {
	result := false

	// Helper function for Main Draw match checking.
	checkMainDraw := func(person *pickMyPostcodePerson) {
		person.MatchMain = winningTickets.Main == person.Entry
		person.MatchVideo = winningTickets.Video == person.Entry
		person.MatchSurvey = winningTickets.Survey == person.Entry
		person.MatchMinidraw = winningTickets.Minidraw == person.Entry
		person.MatchBonus = slices.Contains(winningTickets.Bonus, person.Entry)

		// Check if the person matches any draw.
		person.MatchAny = person.MatchMain || person.MatchVideo || person.MatchSurvey || person.MatchBonus || person.MatchMinidraw

		// Update result if any match found.
		if person.MatchAny {
			result = true
		}
	}

	// Helper function for Stackpot match checking.
	checkStackpot := func(person *pickMyPostcodePerson) {
		if slices.Contains(winningTickets.Stackpot, person.Entry) {
			person.MatchStackpot = true
			result = true // Don't break early in case of multiple winners.
		}
	}

	if isMainDraw {
		for i := range people {
			checkMainDraw(&people[i])
		}
	} else {
		for i := range people {
			checkStackpot(&people[i])
		}
	}

	return result
}

func formatMainDrawResults(people []pickMyPostcodePerson) string {
	output := "Matches        Main    Video    Survey    Bonus     Minidraw    Any      Bonus Money       Entry\n"
	for _, p := range people {
		output += fmt.Sprintf("%-15s%-10t%-11t%-13t%-12t%-16t%-11t%-23s%v\n", p.Name, p.MatchMain, p.MatchVideo, p.MatchSurvey, p.MatchBonus, p.MatchMinidraw, p.MatchAny, p.BonusMoney, p.Entry)
	}
	return output
}

func formatStackpotResults(people []pickMyPostcodePerson) string {
	var output strings.Builder
	output.WriteString("Matches        Stackpot       Entry\n")
	for _, p := range people {
		output.WriteString(fmt.Sprintf("%-15s%-18t%v\n", p.Name, p.MatchStackpot, p.Entry))
	}
	return output.String()
}

func formatMainDrawPostcodes(winningTickets pickMyPostcodeTickets) string {
	output := "Postcodes     Main             Video           Survey        Bonus          Minidraw\n"
	output += fmt.Sprintf("                     %-14s%-14s%-14s%-15s%-14s\n", winningTickets.Main, winningTickets.Video, winningTickets.Survey, winningTickets.Bonus[0], winningTickets.Minidraw)
	output += fmt.Sprintf("%87s\n", winningTickets.Bonus[1])
	output += fmt.Sprintf("%87s\n", winningTickets.Bonus[2])

	return output
}

func formatStackpotPostcodes(winningTickets pickMyPostcodeTickets) string {
	var builder strings.Builder
	builder.WriteString("Postcodes       Stackpot\n")
	for _, postcode := range winningTickets.Stackpot {
		builder.WriteString(fmt.Sprintf("%30s\n", postcode))
	}

	return builder.String()
}
