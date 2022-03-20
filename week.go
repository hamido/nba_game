package main

import (
	"fmt"
	"time"
)

type MatchChannels struct {
	updateChan chan bool
	quitChan   chan bool
	statsChan  chan Stats
}

type Stats struct {
	Name        string
	AttackCount uint64
	TotalScore  uint64
	PlayerStats map[string]PlayerStats
}

type PlayerStats struct {
	Assists     uint64
	Points2Rate string
	Points3Rate string
}

func Week(numberOfMatches int, updates chan<- []Stats) {
	channels := make(map[int]*MatchChannels)

	for i := 0; i < numberOfMatches; i++ {
		updateChan := make(chan bool)
		quitChan := make(chan bool)
		statsChan := make(chan Stats)
		channels[i] = &MatchChannels{updateChan, quitChan, statsChan}
		go Match(fmt.Sprint(i, "-Home"), fmt.Sprint(i, "-Visitor"), updateChan, quitChan, statsChan)
	}

	for i := 0; i < 48; i++ {
		var stats []Stats
		for _, c := range channels { // for each match
			c.updateChan <- true // query for update
			select {
			case sts := <-c.statsChan:
				stats = append(stats, sts) // collect statistics
			}
		}
		updates <- stats
		time.Sleep(time.Second * 5)
	}

	for _, c := range channels {
		c.quitChan <- true
		close(c.statsChan)
		close(c.updateChan)
		close(c.quitChan)
	}
}
