package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

func main() {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		// chromedp.UserDataDir(""),
		chromedp.Flag("headless", false),
		// chromedp.Flag("disable-extensions", false),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// create chrome instance
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	postcode := "GU11 3NT"

	var res string
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://pickmypostcode.com"),
		chromedp.Click(`//*[@id="v-rebrand"]/div[2]/div[2]/nav/ul/li[5]/button[2]`, chromedp.NodeVisible),
		chromedp.WaitVisible(`#confirm-ticket`),
		chromedp.SendKeys(`#confirm-ticket`, postcode),
		chromedp.WaitVisible(`#confirm-email`),
		chromedp.SendKeys(`#confirm-email`, "andrew_field@hotmail.co.uk"),
		chromedp.Click(`/html/body/div[2]/div[2]/div[2]/main/div[1]/section/div/div/div/form/button`, chromedp.NodeVisible),
		chromedp.Sleep(10*time.Second),
		// chromedp.Click(`/html/body/app-root/app-theme/div/div/app-notice/app-theme/div/div/app-home/div/div[2]/app-footer/div/div[2]/app-action-buttons/div/button[1]`, chromedp.NodeVisible),
		// chromedp.Click(`/html/body/div[2]/div[2]/div/mat-dialog-container/ng-component/app-theme/div/div/div[2]/button[2]`, chromedp.NodeVisible),
		// chromedp.WaitVisible(`/html/body/div[2]/div[2]/div/main/div/section/div/div[1]/p[2]/span`),
		// chromedp.Text(`.result--postcode`, &res, chromedp.NodeVisible),

		// chromedp.Navigate("https://gener8ads.com/account/login?utm_source=extension&utm_medium=auth&utm_campaign=tabs"),
		// chromedp.Sleep(5*time.Second),
		// chromedp.WaitVisible(`//button[@aria-label="Continue with email"]`),
		// chromedp.Click(`//button[@aria-label="Continue with email"]`, chromedp.NodeVisible),
		// chromedp.WaitVisible(`//input[@name="email"]`),
		// chromedp.SendKeys(`//input[@name="email"]`, "andrew_field+gener8ads.com@hotmail.co.uk"),
		// chromedp.WaitVisible(`//input[@name="password"]`),
		// chromedp.SendKeys(`//input[@name="password"]`, "rE4!HrEueasf4T"),
		// chromedp.WaitVisible(`//button[@aria-label="Let's go!"]`),
		// chromedp.Click(`//button[@aria-label="Let's go!"]`, chromedp.NodeVisible),
		// chromedp.Sleep(5*time.Second),
	)

	if err != nil {
		log.Fatal(err)
	}
	println("here")

	var iframes []*cdp.Node
	if err := chromedp.Run(ctx, chromedp.Nodes(`iframe`, &iframes, chromedp.ByQuery)); err != nil {
		log.Fatal(err)
	}

	var text string
	if err := chromedp.Run(ctx,
		chromedp.Text("/html/body/app-root/app-theme/div/div/app-notice/app-theme/div/div/app-home/div/div[1]/div/div[1]/p/app-custom-html/p", &text, chromedp.BySearch, chromedp.FromNode(iframes[0])),
		// chromedp.Click(`#denyAll`, chromedp.ByQuery, chromedp.FromNode(iframes[0])),
		// chromedp.Text("#some_id", &text, chromedp.ByQuery, chromedp.FromNode(iframes[0])),
		chromedp.Sleep(10*time.Second),
	); err != nil {
		log.Fatal(err)
	}
	println("here2")
	println(text)
	return

	erdr := chromedp.Run(ctx,
		chromedp.Click(`/html/body/app-root/app-theme/div/div/app-notice/app-theme/div/div/app-home/div/div[2]/app-footer/div/div[2]/app-action-buttons/div/button[1]`, chromedp.NodeVisible),
		chromedp.Click(`/html/body/div[2]/div[2]/div/mat-dialog-container/ng-component/app-theme/div/div/div[2]/button[2]`, chromedp.NodeVisible),
		chromedp.WaitVisible(`/html/body/div[2]/div[2]/div/main/div/section/div/div[1]/p[2]/span`),
		chromedp.Text(`.result--postcode`, &res, chromedp.NodeVisible),
	)
	println("here2")

	if erdr != nil {
		log.Fatal(erdr)
	}

	otherpostcode := "SW9 8AU"

	fmt.Printf("Main draw text: %v. Contains my postcode: %v. Contains other postcode: %v\n", res, strings.Contains(res, postcode), strings.Contains(res, otherpostcode))

	// err = chromedp.Run(ctx,
	// 	chromedp.Navigate("https://pickmypostcode.com/video/"),
	// 	chromedp.Text(`.result--postcode`, &res, chromedp.NodeVisible),
	// )

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// otherpostcode = "KY11 3BN"

	// fmt.Printf("Video text: %v. Contains my postcode: %v. Contains other postcode: %v\n", res, strings.Contains(res, postcode), strings.Contains(res, otherpostcode))

	err = chromedp.Run(ctx,
		chromedp.Navigate("https://pickmypostcode.com/survey-draw/"),
		// chromedp.Sleep(500*time.Second),

		// chromedp.WaitVisible(`#result-survey > div > div > div > div > button:nth-child(2)`, chromedp.ByQuery),
		// chromedp.Click(`#result-survey > div > div > div > div > button:nth-child(2)`, chromedp.ByQuery),
		chromedp.Click(`/html/body/div[2]/div[2]/div/main/div/section/div/div[2]/div/div[1]/div/div/div[2]/button[2]`, chromedp.NodeVisible),
		// chromedp.Click(`#denyAll`, chromedp.NodeVisible),
		// chromedp.WaitVisible(`//*[@id="result-survey"]/div/div/div/div/button/span[text()="No thanks, not today"]`, chromedp.BySearch),
		// chromedp.Click(`//*[@id="result-survey"]/div/div/div/div/button/span[text()="No thanks, not today"]`, chromedp.BySearch),
		chromedp.Text(`.result--postcode`, &res, chromedp.NodeVisible),
		// chromedp.Sleep(5*time.Second),
	)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Survey text: %v. Contains my postcode: %v. Contains other postcode: %v\n", res, strings.Contains(res, postcode), strings.Contains(res, otherpostcode))

	// var ids []cdp.NodeID

	// err = chromedp.Run(ctx,
	// 	chromedp.Navigate("https://pickmypostcode.com/stackpot/"),
	// 	chromedp.WaitVisible(`#result-header`),
	// 	chromedp.NodeIDs(`.result--postcode`, &ids, chromedp.BySearch),
	// 	chromedp.Sleep(5*time.Second),
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// err = chromedp.Run(ctx,
	// 	chromedp.Text(ids, &res, chromedp.ByNodeID),
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// otherpostcode = "NP20 3LA"

	// fmt.Printf("Stackpot text: %v. Contains other postcode: %v\n", res, strings.Contains(res, otherpostcode))

	// for _, id := range ids {
	// 	err = chromedp.Run(ctx,
	// 		chromedp.Text(id, &res, chromedp.ByNodeID),
	// 	)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	otherpostcode = "NP20 3LA"

	// 	fmt.Printf("Stackpot text: %v. Contains other postcode: %v\n", res, strings.Contains(res, otherpostcode))
	// }
}

// chromedp.Navigate(`https://pkg.go.dev/time`),
// chromedp.Text(`.Documentation-overview`, &res, chromedp.NodeVisible),
