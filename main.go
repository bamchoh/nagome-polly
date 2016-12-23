package main

import (
	"regexp"
	"sync"
	"path/filepath"
	"strconv"
	"bytes"
	"strings"
	"bufio"
	"os"
	"log"
	"encoding/json"
	"time"

	"github.com/bamchoh/nagome-polly/player"
	"github.com/aws/aws-sdk-go/service/polly"
)

var (
	logger *log.Logger
	save_dir string = "mp3"
	pc *PollyConfig
	started_time time.Time
	counter int
)

type Message struct {
	Domain  string          `json:"domain"`
	Command string          `json:"command"`
	Content json.RawMessage `json:"content,omitempty"` // The structure of Content is depend on the Command (and Domain).

	prgno int
}

type CtCommentGot struct {
	No      int       `json:"no"`
	Date    time.Time `json:"date"`
	Raw     string    `json:"raw"`
	Comment string    `json:"comment"`

	UserID           string `json:"user_id"`
	UserName         string `json:"user_name"`
	UserThumbnailURL string `json:"user_thumbnail_url,omitempty"`
	Score            int    `json:"score,omitempty"`
	IsPremium        bool   `json:"is_premium"`
	IsBroadcaster    bool   `json:"is_broadcaster"`
	IsStaff          bool   `json:"is_staff"`
	IsAnonymity      bool   `json:"is_anonymity"`
}

type CtNagomeBroadOpen struct {
	BroadID     string `json:"broad_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CommunityID string `json:"community_id"`
	OwnerID     string `json:"owner_id"`
	OwnerName   string `json:"owner_name"`
	OwnerBroad  bool   `json:"owner_broad"`

	OpenTime  time.Time `json:"open_time"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

func set_log() *log.Logger {
	basedir := filepath.Dir(os.Args[0])
	log_path := filepath.Join(basedir, "nagome-polly.log")
	f,_ := os.Create(log_path)
	return log.New(f, "nagome-polly:", 0)
}

func send_aws(msg string, m *sync.Mutex, speed int) (resp *polly.SynthesizeSpeechOutput, err error) {
	m.Lock()
	defer m.Unlock()
	packed_msg := `<speak><prosody rate="`+strconv.Itoa(speed)+`%"><![CDATA[`+msg+`]]></prosody></speak>`

	resp,err = synthesize_speech(pc, packed_msg)
	return
}

func play(resp *polly.SynthesizeSpeechOutput, m *sync.Mutex) (err error) {
	m.Lock()
	defer m.Unlock()
	err = player.Play(resp)

	if err != nil {
		logger.Println(err)
	}
	return
}

func increment_count() {
	counter++
}

func descrement_count() {
	counter--
}

func read_aloud(broad_id string, content []byte, m1 *sync.Mutex, m2 *sync.Mutex) (err error) {
	dec := json.NewDecoder(bytes.NewReader(content))
	com := new(CtCommentGot)
	err = dec.Decode(com)
	if err != nil {
		logger.Println(err)
		return
	}

	if com.Date.Before(started_time) {
		return
	}

	matched,err := regexp.MatchString("^/hb", com.Raw)
	if err != nil {
		logger.Println(err)
		return
	}

	if matched {
		return
	}

	go func(msg string) {
		increment_count()
		defer descrement_count()

		if counter > 10 {
			logger.Println("Skipped : ", msg)
			return
		}

		speed := 100
		speed += 20 * (counter-1)

		resp,err := send_aws(msg,m1,speed)
		if err != nil {
			logger.Println(err)
			return
		}
		play(resp,m2)
	}(string(com.Raw))

	return
}

func init_plugin() (err error) {
	basedir := filepath.Dir(os.Args[0])
	filepath := filepath.Join(basedir, "nagome-polly.yml")
	f,err := os.Open(filepath)
	if err != nil {
		log.Println(err)
		return err
	}
	defer f.Close()

	pc,err = load(f)
	if err != nil {
		log.Println(err)
		return err
	}

	started_time = time.Now()

	return
}

func pick_broad_id(content []byte) (broad_id string, err error) {
	dec := json.NewDecoder(bytes.NewReader(content))
	cnbo := new(CtNagomeBroadOpen)
	err = dec.Decode(cnbo)
	if err != nil {
		logger.Println(err)
		return
	}
	broad_id = cnbo.BroadID
	return
}

func main() {
	var m1 *sync.Mutex = new(sync.Mutex)
	var m2 *sync.Mutex = new(sync.Mutex)
	var broad_id string
	logger = set_log()

	err := init_plugin()
	if err != nil {
		logger.Println(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		txt := scanner.Text()
		dec := json.NewDecoder(strings.NewReader(txt))
		msg := new(Message)
		err := dec.Decode(msg)
		if err != nil {
			logger.Println(err)
		}
		switch msg.Command {
		case "Broad.Open":
			broad_id, err = pick_broad_id(msg.Content)
			if err != nil {
				logger.Println(msg.Command, err)
			}
		case "Got":
			read_aloud(broad_id, msg.Content, m1, m2)
		default:
			logger.Println(msg.Command)
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Println("reading standard input:", err)
	}
}
