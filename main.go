package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var sessionId = ""
var devId = ""
var authKey = ""

var timestamp = ""

func main() {

	//concurrent_sessions:  50
	//sessions_per_day: 500
	//session_time_limit:  15 minutes
	//request_day_limit:  7500

	go createSession()

	discordAuthToken := "" // Replace with Bot Auth Token
	dg, err := discordgo.New("Bot " + discordAuthToken)

	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()

}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Delimiter of !ranked to lookup player statistics and return the data into the Discord server
	if strings.Contains(m.Content, "!ranked ") {
		parts := strings.Split(m.Content, "!ranked ")
		fmt.Println(parts[1])

		rankedStats := gjson.Get(getPlayer(parts[1]), "0").String()

		if rankedStats != "" {
			accountLvl := gjson.Get(getPlayer(parts[1]), "0.Level").String()
			hoursPlayed := gjson.Get(getPlayer(parts[1]), "0.HoursPlayed").String()
			wins := gjson.Get(getPlayer(parts[1]), "0.RankedConquest.Wins").String()
			losses := gjson.Get(getPlayer(parts[1]), "0.RankedConquest.Losses").String()
			mmr := gjson.Get(getPlayer(parts[1]), "0.RankedConquest.Rank_Stat").String()
			rank := createRank(gjson.Get(getPlayer(parts[1]), "0.RankedConquest.Tier").String())

			s.ChannelMessageSend(m.ChannelID, parts[1]+": Account Level "+accountLvl+" / "+"Hours Played "+hoursPlayed)
			time.Sleep(1 * time.Second)
			s.ChannelMessageSend(m.ChannelID, parts[1]+": Wins "+wins+" / "+"Losses "+losses)
			time.Sleep(1 * time.Second)
			s.ChannelMessageSend(m.ChannelID, parts[1]+": Rank "+rank+" / "+"MMR "+mmr)

		} else {
			s.ChannelMessageSend(m.ChannelID, parts[1]+": Error on lookup / hidden profile")
		}

		//s.ChannelMessageSend(m.ChannelID, userInfo)
	}

	// Delimiter of !playerStatus to determine if the player is in a match or not
	if strings.Contains(m.Content, "!playerStatus ") {
		parts := strings.Split(m.Content, "!playerStatus ")
		fmt.Println(parts[1])

		userInfo := getPlayerStatus(parts[1])

		statusString := gjson.Get(userInfo, "0.status_string").String()
		matchQueueID := gjson.Get(userInfo, "0.match_queue_id").String()

		s.ChannelMessageSend(m.ChannelID, parts[1]+": Status "+statusString+" / "+"Match Queue ID "+matchQueueID)
	}

	// Delimiter of !playerID which returns the requested players ID
	if strings.Contains(m.Content, "!playerID ") {
		parts := strings.Split(m.Content, "!playerID ")
		fmt.Println(parts[1])

		userInfo := getPlayerIDName(parts[1])
		s.ChannelMessageSend(m.ChannelID, userInfo)
	}

}

// Used to create a session which lasts 15min to the HiRez API
func createSession() {
	for {
		timestamp = GetCurrentTime()
		signature := createSignature(devId, "createsession", authKey, GetCurrentTime())

		resp, err := http.Get("http://api.smitegame.com/smiteapi.svc/createsessionJson/" + devId + "/" + signature + "/" + GetCurrentTime())
		if err != nil {
			// handle error
		}

		bodyBytes, _ := ioutil.ReadAll(resp.Body) // TODO: handle error
		bodyString := string(bodyBytes)

		sessionId = gjson.Get(bodyString, "session_id").String()
		time.Sleep(15 * time.Minute)
	}
}

func getPlayer(username string) string {
	signature := createSignature(devId, "getplayer", authKey, GetCurrentTime())

	resp, err := http.Get("http://api.smitegame.com/smiteapi.svc/getplayerjson/" + devId + "/" + signature + "/" + sessionId + "/" + GetCurrentTime() + "/" + username)

	if err != nil {
		// handle error
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	return bodyString
}

func getPlayerStatus(playerID string) string {
	signature := createSignature(devId, "getplayerstatus", authKey, GetCurrentTime())

	resp, err := http.Get("http://api.smitegame.com/smiteapi.svc/getplayerstatusjson/" + devId + "/" + signature + "/" + sessionId + "/" + GetCurrentTime() + "/" + playerID)

	if err != nil {
		// handle error
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	return bodyString
}

func getPlayerIDName(username string) string {
	signature := createSignature(devId, "getplayeridbyname", authKey, GetCurrentTime())

	resp, err := http.Get("http://api.smitegame.com/smiteapi.svc/getplayeridbynamejson/" + devId + "/" + signature + "/" + sessionId + "/" + GetCurrentTime() + "/" + username)

	if err != nil {
		// handle error
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	return bodyString
}

func createSignature(devId, functionName, authKey, timestamp string) string {
	return GetMD5Hash(devId + functionName + authKey + timestamp)
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func GetCurrentTime() string {
	// string timestamp = DateTime.UtcNow.ToString("yyyyMMddHHmmss");
	t := time.Now().UTC()
	return t.Format("20060102150405") // https://stackoverflow.com/a/20234207
}

func createRank(rank string) string {

	switch {
	case rank == "1":
		return "Bronze V"
	case rank == "2":
		return "Bronze IV"
	case rank == "3":
		return "Bronze III"
	case rank == "4":
		return "Bronze II"
	case rank == "5":
		return "Bronze I"
	case rank == "6":
		return "Silver V"
	case rank == "7":
		return "Silver IV"
	case rank == "8":
		return "Silver III"
	case rank == "9":
		return "Silver II"
	case rank == "10":
		return "Silver I"
	case rank == "11":
		return "Gold V"
	case rank == "12":
		return "Gold IV"
	case rank == "13":
		return "Gold III"
	case rank == "14":
		return "Gold II"
	case rank == "15":
		return "Gold I"
	case rank == "16":
		return "Platinum V"
	case rank == "17":
		return "Platinum IV"
	case rank == "18":
		return "Platinum III"
	case rank == "19":
		return "Platinum II"
	case rank == "20":
		return "Platinum I"
	case rank == "21":
		return "Diamond V"
	case rank == "22":
		return "Diamond IV"
	case rank == "23":
		return "Diamond III"
	case rank == "24":
		return "Diamond II"
	case rank == "25":
		return "Diamond I"
	case rank == "26":
		return "Masters"
	case rank == "27":
		return "Grandmasters"
	}

	return "0"
}
