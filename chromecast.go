/*
   googlehome-private-apps
   Copyright (C) 2019, Sergio Conde Gomez <skgsergio@gmail.com>

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"context"
	"net"
	"time"

	"github.com/barnybug/go-cast"
	"github.com/barnybug/go-cast/controllers"

	"github.com/rs/zerolog/log"
)

const (
	AudioFSPath  = "audio"
	AudioURLPath = "/audio"
)

// Return audio file path
func getAudioFilePath(filename string) string {
	return AudioFSPath + "/" + filename
}

// Return audio file URL
func getAudioFileURL(filename string) (string, error) {
	localIP, err := getLocalIP()

	if err != nil {
		return "", err
	}

	return "http://" + localIP.String() + ":" + ServerPort + AudioURLPath + "/" + filename, nil
}

// Cast audio file
func sendAudioFile(filename string) {
	url, err := getAudioFileURL(filename)

	if err != nil {
		log.Error().Err(err).Msg("sendAudioFile")
		return
	}

	sendAudioURL(url)
}

// Cast audio url
func sendAudioURL(url string) {
	log.Debug().Str("action", "sending").Str("url", url).Msg("sendAudioURL")

	// Initialize cast client
	ctx := context.Background()
	client := cast.NewClient(net.ParseIP(CastIP), CastPort)
	err := client.Connect(ctx)

	if err != nil {
		log.Error().Err(err).Msg("sendAudioURL")
		return
	}

	// Get media controller
	media, err := client.Media(ctx)

	if err != nil {
		client.Close()
		log.Error().Err(err).Msg("sendAudioURL")
		return
	}

	// Create media item and load it
	item := controllers.MediaItem{
		ContentId:  url,
		StreamType: "BUFFERED",
	}

	_, err = media.LoadMedia(ctx, item, 0, true, map[string]interface{}{})

	if err != nil {
		log.Error().Err(err).Msg("sendAudioURL")

		_, err := client.Receiver().QuitApp(ctx)
		if err != nil {
			log.Error().Err(err).Msg("sendAudioURL")
		}

		client.Close()
		return
	}

	// Wait until the media item finished playing (this way we can close the client)
	errs := 0
	for {
		time.Sleep(time.Second)

		response, err := media.GetStatus(ctx)

		if err != nil {
			log.Error().Err(err).Msg("sendAudioURL")
			errs += 1

			if errs > 4 {
				break
			}

			continue
		}

		// Iterate all statuses in search for our contentId
		errs = 0
		found := false

		for _, status := range response.Status {
			found = status.Media.ContentId == item.ContentId
		}

		// If the contentId is not present we can asume it finished playing
		if !found {
			break
		}

		log.Debug().Str("action", "playing").Str("url", url).Msg("sendAudioURL")
	}

	// Quit the app and close the client
	log.Debug().Str("action", "closing").Str("url", url).Msg("sendAudioURL")

	_, err = client.Receiver().QuitApp(ctx)

	if err != nil {
		log.Error().Err(err).Msg("sendAudioURL")
	}

	client.Close()
}
