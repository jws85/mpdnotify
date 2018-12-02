package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fhs/gompd/mpd"

	"github.com/esiqveland/notify"
	"github.com/godbus/dbus"
)

var MPD_SERVER = "localhost:6600"

func main() {
	dbusconn, err := dbus.SessionBus()
	if err != nil {
		panic(err)
	}

	watcher, err := mpd.NewWatcher("tcp", MPD_SERVER, "")
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	oldsummary := ""
	oldbody := ""
	for event := range watcher.Event {
		if event == "player" {
			mpdconn, err := mpd.Dial("tcp", MPD_SERVER)
			if err != nil {
				panic(err)
			}

			summary, body, err := buildNotifyStrings(mpdconn)
			if err != nil {
				log.Fatalln(err)
			}

			mpdconn.Close()

			if summary != oldsummary || body != oldbody {
				oldbody = body
				oldsummary = summary
				sendNotification(dbusconn, summary, body)
			}
		}
	}
}

func buildNotifyStrings(conn *mpd.Client) (string, string, error) {
	status, err := conn.Status()
	if err != nil {
		return "", "", err
	}

	song, err := conn.CurrentSong()
	if err != nil {
		return "", "", err
	}

	artist, exist := song["Artist"]
	if !exist {
		artist = "No artist"
	}

	title, exist := song["Title"]
	if !exist {
		title = "No title"
	}

	album, albumExist := song["Album"]
	track, exist := song["Track"]
	if !exist {
		track = "0"
	}

	summary := fmt.Sprintf("MPD: %s", status["state"])
	body := ""
	if status["state"] != "stopped" {
		body = fmt.Sprintf("%s\n%s", title, artist)
		if albumExist {
			body += fmt.Sprintf("\n#%s %s", track, album)
		}
	}

	return summary, body, nil
}

func sendNotification(dbusconn *dbus.Conn, summary string, body string) {
	iconName := "emblem-music"
	exe, err := os.Executable()
	if err == nil {
		iconName = filepath.Dir(exe) + "/music-note.svg"
	}

	// Basic usage
	// Create a Notification to send
	n := notify.Notification{
		AppName:       "mpd",
		ReplacesID:    uint32(0),
		AppIcon:       iconName,
		Summary:       summary,
		Body:          body,
		Hints:         map[string]dbus.Variant{},
		ExpireTimeout: int32(5000),
	}

	// Ship it!
	_, err = notify.SendNotification(dbusconn, n)
	if err != nil {
		log.Printf("error sending notification: %v", err.Error())
	}
}
