/*
 Copyright 2015 Patrick Moroney

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"encoding/csv"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/quipo/statsd"
)

func getStats() {
	file, err := ioutil.TempFile("", "sar")
	if err != nil {
		log.Print(err)
		return
	}
	defer func() {
		err = os.Remove(file.Name())
		if err != nil {
			log.Print(err)
		}
	}()
	err = exec.Command("sar", "1", "1", "-A", "-o", file.Name()).Run()
	if err != nil {
		log.Print(err)
		return
	}
	cmd := exec.Command("sadf", "1", "1", "--", "-A", file.Name())
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Print(err)
		return
	}
	if err := cmd.Start(); err != nil {
		log.Print(err)
		return
	}
	csv := csv.NewReader(stdout)
	csv.Comma = '	'
	records, err := csv.ReadAll()
	if err != nil {
		log.Print(err)
		return
	}
	if err := cmd.Wait(); err != nil {
		log.Print(err)
		return
	}
	for _, record := range records {
		value, err := strconv.ParseFloat(record[5], 64)
		if err != nil {
			//log.Print(err)
			continue
		}
		metric := record[4]
		if metric[:2] == "kb" {
			metric = metric[2:]
			value = value * 1024
		} else if len(metric) > 3 && metric[2:4] == "kB" {
			metric = metric[:2] + "bit" + metric[4:]
			value = value * 1024 * 8
		}
		if record[3] != "-" && record[3] != "all" {
			metric = record[3] + "." + metric
		}
		metric = strings.Replace(metric, "%", "pct_", -1)
		sd.FGauge(metric, value)
		//fmt.Print(metric, "=", value, "\n")
	}

}

var statsdServer = flag.String("d", ":8125", "Destination statsd server address")
var statsdPrefix = flag.String("p", "sar", "Statsd prefix for metrics")
var interval = flag.Duration("i", 10*time.Second, "Interval to send metrics")
var sd = statsd.NewStatsdClient(*statsdServer, *statsdPrefix+".")

func main() {
	flag.Parse()
	err := sd.CreateSocket()
	if err != nil {
		log.Fatal(err)
	}
	c := time.Tick(*interval)
	for _ = range c {
		go getStats()
	}
}
