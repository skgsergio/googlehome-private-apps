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
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/evalphobia/google-tts-go/googletts"

	"github.com/rs/zerolog/log"
)

const DefaultTTSLang = "es"

// Grabs a TTS audio file with the text and lang specified
func getTTSFile(text string, lang string) (string, error) {
	// This hash is for caching purposes, as we are using Google Translator TTS
	text = strings.ToLower(text)
	hash := sha256.Sum256([]byte(lang + text))

	filename := "tts-" + hex.EncodeToString(hash[:]) + ".mp3"
	filepath := getAudioFilePath(filename)

	// If the file already exists just
	if fileExists(filepath) {
		return filename, nil
	}

	// Get Google Translator TTS url
	url, err := googletts.GetTTSURL(text, lang)

	if err != nil {
		return filename, err
	}

	// Download the audio file
	err = downloadFile(url, filepath)

	return filename, err
}

// Casts a message using Google Translator TTS
func sayTTS(text string, lang string) (string, error) {
	filename, err := getTTSFile(text, lang)

	if err != nil {
		return "", err
	}

	url, err := getAudioFileURL(filename)

	// Casts the audio we are serving
	if err == nil {
		go sendAudioURL(url)
	}

	return url, err
}

const sayTTSURLPath = "/say"

// Say TTS route handler
func sayTTSHandler(ctx *gin.Context) {
	var text string
	var lang string
	var qs = ctx.Request.URL.Query()

	// Get the t url query string parameter, only one allowed
	if value, ok := qs["t"]; ok {
		if len(value) > 1 {
			ctx.JSON(400, gin.H{"error": "multiple text parameters, only one allowed"})
			return
		}

		text = value[0]
	} else {
		ctx.JSON(400, gin.H{"error": "missing text parameter"})
		return
	}

	// Gets the optional language query string parameter, only one allowed
	if value, ok := qs["l"]; ok {
		if len(value) > 1 {
			ctx.JSON(400, gin.H{"error": "multiple lang parameters, only one allowed"})
			return
		}

		lang = value[0]
	} else {
		lang = DefaultTTSLang
	}

	// Get the audio file and casts it
	url, err := sayTTS(text, lang)

	if err != nil {
		log.Error().Err(err).Msg("sayTTSHandler")
		ctx.JSON(500, gin.H{"error": "couldn't play TTS, check logs"})
		return
	}

	ctx.JSON(200, gin.H{"sent": url})
}
