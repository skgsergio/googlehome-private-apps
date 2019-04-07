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
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/logger"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	CastIP     = "192.168.1.7"
	CastPort   = 8009
	ServerPort = "8080"
)

func main() {
	// Initialize custom zerolog console output
	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out:        os.Stderr,
			NoColor:    false,
			TimeFormat: time.RFC3339,
		},
	)

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Set debug level if gin is running as debug
	if gin.IsDebugging() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	r := gin.New()

	// Use recovery and logger middlewares
	r.Use(gin.Recovery())
	r.Use(logger.SetLogger())

	// Serve audio statics
	r.Static(AudioURLPath, AudioFSPath)

	// Serve paths
	r.GET(sayTTSURLPath, sayTTSHandler)

	r.Run("0.0.0.0:" + ServerPort)
}
