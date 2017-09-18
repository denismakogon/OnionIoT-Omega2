package main

import (
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/denismakogon/omega2-apps/serverless/twitter-daemon/api"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

func asyncRunner(omega *api.OnionOmega2, fnAPIURL, fnToken string, proc func(tweet anaconda.Tweet, httpClient *http.Client, fnAPIURL, fnToken string) error) {
	httpClient := api.SetupHTTPClient()

	tweetID := os.Getenv("InitialTweetID")
	if tweetID == "" {
		panic("Initial tweet ID env var is not set, but suppose to be!")
	}
	omega.SetTweetIDToStartFrom(tweetID)
	wg := new(sync.WaitGroup)
	for {
		tweets, err := omega.GetRecentMentions()
		if err != nil {
			fmt.Println(err.Error())
			panic(err.Error())
		}
		if len(tweets) != 0 {
			wg.Add(len(tweets))
			for _, tweet := range tweets {
				omega.PrintTweetInfo(tweet)
				go func() {
					defer wg.Done()
					err = proc(tweet, httpClient, fnAPIURL, fnToken)
					if err != nil {
						panic(err.Error())
					}
				}()
			}
			wg.Wait()
		}
		time.Sleep(time.Second * 6)
	}
}

func EmotionRecognition() {
	pgConf := new(api.PostgresConfig)
	pgConf.FromFile()
	twitter := new(api.TwitterSecret)
	twitterAPI, err := twitter.FromFile()
	if err != nil {
		panic(err.Error())
	}
	fnAPIURL, fnToken, err := api.SetupEmoKognitionFunctions(twitter, pgConf)
	if err != nil {
		panic(err.Error())
	}

	// get latest 10 tweets fro InitialTweet
	v := url.Values{}
	v.Set("count", "10")

	omega := api.OnionOmega2{
		TwitterAPI:   twitterAPI,
		SearchValues: &v,
	}
	asyncRunner(&omega, fnAPIURL, fnToken, api.ProcessTweetWithEmotion)
}

func LandmarkRecognition() {

	twitter := new(api.TwitterSecret)
	twitterAPI, err := twitter.FromFile()
	if err != nil {
		panic(err.Error())
	}

	gc := new(api.GCloudSecret)
	err = gc.FromFile()
	if err != nil {
		panic(err.Error())
	}

	fnAPIURL, fnToken, err := api.SetupLandmarkRecognitionFunctions(gc, twitter)
	if err != nil {
		panic(err.Error())
	}

	// get latest 200 tweets
	v := url.Values{}
	v.Set("count", "200")

	omega := api.OnionOmega2{
		TwitterAPI:   twitterAPI,
		SearchValues: &v,
		GCloudAuth:   gc,
	}
	asyncRunner(&omega, fnAPIURL, fnToken, api.ProcessTweetWithLandmark)
}

func main() {
	botType := os.Getenv("TwitterBotType")
	if botType == "landmark" {
		LandmarkRecognition()
	}
	if botType == "emokognition" {
		EmotionRecognition()
	}
	if botType == "" {
		panic("Recognition type is not set.")
	}
}
