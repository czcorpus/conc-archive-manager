// Copyright 2024 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2024 Institute of the Czech National Corpus,
//                Faculty of Arts, Charles University
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cnf

import (
	"camus/archiver"
	"camus/cleaner"
	"camus/cncdb"
	"camus/indexer"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/czcorpus/cnc-gokit/logging"
	"github.com/czcorpus/hltscl"
	"github.com/rs/zerolog/log"
)

const (
	dfltServerWriteTimeoutSecs = 30
	dfltLanguage               = "en"
	dfltTimeZone               = "Europe/Prague"
)

type Conf struct {
	srcPath                string
	ListenAddress          string              `json:"listenAddress"`
	PublicURL              string              `json:"publicUrl"`
	ListenPort             int                 `json:"listenPort"`
	ServerReadTimeoutSecs  int                 `json:"serverReadTimeoutSecs"`
	ServerWriteTimeoutSecs int                 `json:"serverWriteTimeoutSecs"`
	CorsAllowedOrigins     []string            `json:"corsAllowedOrigins"`
	TimeZone               string              `json:"timeZone"`
	AuthHeaderName         string              `json:"authHeaderName"`
	AuthTokens             []string            `json:"authTokens"`
	Logging                logging.LoggingConf `json:"logging"`
	Redis                  *archiver.RedisConf `json:"redis"`
	MySQL                  *cncdb.DBConf       `json:"db"`
	Archiver               *archiver.Conf      `json:"archiver"`
	Indexer                *indexer.Conf       `json:"indexer"`
	Cleaner                cleaner.Conf        `json:"cleaner"`
	Reporting              hltscl.PgConf       `json:"reporting"`
}

func (conf *Conf) TimezoneLocation() *time.Location {
	// we can ignore the error here as we always call c.Validate()
	// first (which also tries to load the location and report possible
	// error)
	loc, _ := time.LoadLocation(conf.TimeZone)
	return loc
}

func LoadConfig(path string) *Conf {
	if path == "" {
		log.Fatal().Msg("Cannot load cnfig - path not specified")
	}
	rawData, err := os.ReadFile(path)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot load config")
	}
	var conf Conf
	conf.srcPath = path
	err = json.Unmarshal(rawData, &conf)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot load config")
	}
	return &conf
}

func ValidateAndDefaults(conf *Conf) {
	if conf.ServerWriteTimeoutSecs == 0 {
		conf.ServerWriteTimeoutSecs = dfltServerWriteTimeoutSecs
		log.Warn().Msgf(
			"serverWriteTimeoutSecs not specified, using default: %d",
			dfltServerWriteTimeoutSecs,
		)
	}
	if conf.PublicURL == "" {
		conf.PublicURL = fmt.Sprintf("http://%s", conf.ListenAddress)
		log.Warn().Str("address", conf.PublicURL).Msg("publicUrl not set, using listenAddress")
	}

	if conf.TimeZone == "" {
		log.Warn().
			Str("timeZone", dfltTimeZone).
			Msg("time zone not specified, using default")
	}
	if _, err := time.LoadLocation(conf.TimeZone); err != nil {
		log.Fatal().Err(err).Msg("invalid time zone")
	}

	if err := conf.Redis.ValidateAndDefaults(); err != nil {
		log.Fatal().Err(err).Msg("invalid Redis configuration")
	}

	if err := conf.Archiver.ValidateAndDefaults(); err != nil {
		log.Fatal().Err(err).Msg("invalid archiver configuration")
	}

	if err := conf.Cleaner.ValidateAndDefaults(conf.Archiver.CheckIntervalSecs); err != nil {
		log.Fatal().Err(err).Msg("invalid Clean configuration")
	}

	if err := conf.Indexer.ValidateAndDefaults(); err != nil {
		log.Fatal().Err(err).Msg("invalid indexer configuration")
	}
}
