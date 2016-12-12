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
	"os/exec"
	"log"
	"encoding/json"
	"time"

	"github.com/bamchoh/nagome-talk/polly"
)

var (
	logger *log.Logger
	save_dir string = "mp3"
	wg *sync.WaitGroup = new(sync.WaitGroup)
	m *sync.Mutex = new(sync.Mutex)
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
	f,_ := os.Create("test.log")
	return log.New(f, "nagome-talk:", 0)
}

func play(save_file string, m *sync.Mutex) {
	m.Lock()
	defer m.Unlock()
	var err error

	err = polly.Play(save_file)
	if err != nil {
		logger.Println(err)
		return
	}

	err = os.Remove(save_file)
	if err != nil {
		logger.Println(err)
		return
	}

	return
}

func send_aws(msg, file string) (err error) {
	packed_msg := `<speak><prosody rate="100%"><![CDATA[`+msg+`]]></prosody></speak>`
	cmd := exec.Command(
		"aws",
		"polly",
		"synthesize-speech",
		"--region", "us-west-2",
		"--output-format", "mp3",
		"--voice-id", "Mizuki",
		"--text-type", "ssml",
		"--text", packed_msg,
		file,
	)

	err = cmd.Run()

	if err != nil {
		logger.Println(err)
	}
	return
}

func read_aloud(broad_id string, content []byte) (err error) {
	dec := json.NewDecoder(bytes.NewReader(content))
	com := new(CtCommentGot)
	err = dec.Decode(com)
	if err != nil {
		logger.Println(err)
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

	no := strconv.Itoa(com.No)
	save_file := filepath.Join(save_dir, broad_id+"_"+no+".mp3")

	go func(msg, file string, m *sync.Mutex) {
		send_aws(msg,file)
		wg.Add(1)
		play(file, m)
		wg.Done()
	}(string(com.Raw), save_file, m)

	return
}

func init_plugin() (err error) {
	if _,err = os.Stat(save_dir); err == nil {
		if err = os.RemoveAll(save_dir); err != nil {
			return
		}
	}

	if _,err = os.Stat(save_dir); err != nil {
		logger.Println(err)
		err = os.MkdirAll(save_dir, 0755)
		return
	}

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
			pick_broad_id(msg.Content)
		case "Got":
			read_aloud(broad_id, msg.Content)
		default:
			logger.Println(msg.Command)
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Println("reading standard input:", err)
	}

	wg.Wait()
	// scanner := bufio.NewScanner(os.Stdin)
	// for scanner.Scan() {
	// 	fmt.Println("--- got a comment ---")
	// }
}
