package main

import (
	"bytes"
	"encoding/json"
)

const STATS_FILE = "mud-stats.json"

var mudStats mudStatsData

type mudStatsData struct {
	Version   int
	LoginEver uint64
	MostEver  int

	loginCount uint64
}

func writeMudStats() {

	mudStats.Version = MUDSTATS_VERSION
	fileName := DATA_DIR + STATS_FILE
	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	err := enc.Encode(&mudStats)
	if err != nil {
		critLog("writeStats: enc.Encode: %v", err.Error())
		return
	}

	err = saveFile(fileName, outbuf.Bytes())
	if err != nil {
		critLog("writeStats: saveFile failed %v", err.Error())
		return
	}

}

func loadMudStats() error {

	fileName := DATA_DIR + STATS_FILE
	data, err := readFile(fileName)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &mudStats)
	if err != nil {
		critLog("readStats: Unable to unmarshal the data.")
		return err
	}

	return nil
}
