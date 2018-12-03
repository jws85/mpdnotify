package main

import (
	"errors"
	"fmt"
	"log"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"crypto/sha1" // digest "hash" for short strings
	"encoding/base64" // digest "hash" for short strings

	"github.com/fhs/gompd/mpd"
	"github.com/dhowden/tag"

	"github.com/esiqveland/notify"
	"github.com/godbus/dbus"
)

type Config struct {
	MPDServerURL string
	MusicDirectory string
	DefaultIcon string
	PopupTimeout int32
}

func main() {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}

	var config Config
	config.MPDServerURL = "localhost:6600"
	config.MusicDirectory = usr.HomeDir + "/Music"
	config.PopupTimeout = int32(5000)

	// Default icon
	config.DefaultIcon = "emblem-music"
	exe, err := os.Executable()
	if err == nil {
		config.DefaultIcon = filepath.Dir(exe) + "/music-note.svg"
	}

	dbusconn, err := dbus.SessionBus()
	if err != nil {
		panic(err)
	}

	watcher, err := mpd.NewWatcher("tcp", config.MPDServerURL, "")
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	oldsummary := ""
	oldbody := ""
	for event := range watcher.Event {
		if event == "player" {
			mpdconn, err := mpd.Dial("tcp", config.MPDServerURL)
			if err != nil {
				panic(err)
			}

			status, err := mpdconn.Status()
			if err != nil {
				panic(err)
			}

			song, err := mpdconn.CurrentSong()
			if err != nil {
				panic(err)
			}

			summary, body := buildNotifyStrings(song, status)

			icon := config.DefaultIcon
			if albumArt, err := getAlbumArt(config, song); err == nil {
				icon = albumArt
			}

			mpdconn.Close()

			if summary != oldsummary || body != oldbody {
				oldbody = body
				oldsummary = summary
				sendNotification(config, dbusconn, summary, body, icon)
			}
		}
	}
}

func buildNotifyStrings(song mpd.Attrs, status mpd.Attrs) (string, string) {
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

	return summary, body
}

func getAlbumArt(config Config, song mpd.Attrs) (string, error) {
	// Attempt to read from music file
	songfile, err := os.Open(config.MusicDirectory + "/" + song["file"])
	if err == nil {
		data, err := tag.ReadFrom(songfile)
		if err == nil {
			picture := data.Picture()

			if picture != nil {
				// digest album art image so we don't create zillons of files in /tmp
				hash := sha1.New()
				hash.Write(picture.Data)
				digest := base64.URLEncoding.EncodeToString(hash.Sum(nil))
				icon := fmt.Sprintf("/tmp/mpdnotify.%s.%s", digest, picture.Ext)

				if _, err = os.Stat(icon); os.IsNotExist(err) {
					ioutil.WriteFile(icon, picture.Data, 0644)
				}

				return icon, nil
			}
		}
	}

	// Attempt to find file in music directory
	songdir := config.MusicDirectory + "/" + filepath.Dir(song["file"])
	if _, err = os.Stat(songdir + "/album.jpg"); err == nil {
		return songdir + "/album.jpg", nil
	}
	if _, err = os.Stat(songdir + "/album.png"); err == nil {
		return songdir + "/album.png", nil
	}
	if _, err = os.Stat(songdir + "/cover.jpg"); err == nil {
		return songdir + "/cover.jpg", nil
	}
	if _, err = os.Stat(songdir + "/cover.png"); err == nil {
		return songdir + "/cover.png", nil
	}

	// Couldn't find anything
	return "", errors.New("Failed to find album art")
}

func sendNotification(config Config, dbusconn *dbus.Conn, summary string, body string, icon string) {
	// Basic usage
	// Create a Notification to send
	n := notify.Notification{
		AppName:       "mpd",
		ReplacesID:    uint32(0),
		AppIcon:       icon,
		Summary:       summary,
		Body:          body,
		Hints:         map[string]dbus.Variant{},
		ExpireTimeout: config.PopupTimeout,
	}

	// Ship it!
	_, err := notify.SendNotification(dbusconn, n)
	if err != nil {
		log.Printf("error sending notification: %v", err.Error())
	}
}
