package main

import (
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/bitly/go-simplejson"
	"github.com/garyburd/redigo/redis"
	"github.com/soveran/redisurl"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

// Endpoint of nyanpass.com.
const URLEndpoint = "http://nyanpass.com"

// URL for getting nyampass count.
const GetURL = "http://nyanpass.com/get"

func main() {
	c := getRedisConn()
	defer c.Close()

	currentNyanpass := getCurrentNyanpass()
	pastNyanpass := getPastNyanpass(c)
	c.Do("set", "count", currentNyanpass)
	if pastNyanpass == 0 {
		os.Exit(0)
	}

	diff := currentNyanpass - pastNyanpass

	api := getTwitterAPI(c)
	defer api.Close()
	text := fmt.Sprintf("にゃんぱすー\n今日は%dにゃんぱすーなんなー\n合計%dにゃんぱすーなんなー\n%s", diff, currentNyanpass, URLEndpoint)
	tweet, err := api.PostTweet(text, nil)
	checkErr(err)
	fmt.Println(tweet.Text)
}

func getRedisConn() redis.Conn {
	if url := os.Getenv("REDIS_URL"); url != "" {
		conn, err := redisurl.ConnectToURL(url)
		checkErr(err)
		return conn
	}
	hostname, err := os.Hostname()
	checkErr(err)
	conn, err := redis.Dial("tcp", hostname+":6379")
	checkErr(err)
	return conn
}

func getCurrentNyanpass() int64 {
	v := url.Values{}
	v.Set("nyan", "pass")
	resp, err := http.PostForm(GetURL, v)
	checkErr(err)
	defer resp.Body.Close()
	rf, err := ioutil.ReadAll(resp.Body)
	checkErr(err)
	js, err := simplejson.NewJson(rf)
	checkErr(err)
	currentCount, err := js.Get("cnt").String()
	checkErr(err)
	countInt64, err := strconv.ParseInt(currentCount, 10, 64)
	return countInt64
}

func getPastNyanpass(c redis.Conn) int64 {
	var pastCount int64
	var err error
	if pastCount, err = redis.Int64(c.Do("get", "count")); err != nil {
		pastCount = 0
	}
	return pastCount
}

func getTwitterAPI(c redis.Conn) *anaconda.TwitterApi {
	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")
	accessToken := os.Getenv("ACCESS_TOKEN")
	accessTokenSecret := os.Getenv("ACCESS_TOKEN_SECRET")

	if consumerKey == "" || consumerSecret == "" || accessToken == "" || accessTokenSecret == "" {
		panic("うち、機械(Twitter)にはうといのん")
	}
	anaconda.SetConsumerKey(consumerKey)
	anaconda.SetConsumerSecret(consumerSecret)
	return anaconda.NewTwitterApi(accessToken, accessTokenSecret)
}

func checkErr(err error) {
	if err != nil {
		fmt.Println("すこー")
		panic(err)
	}
}
