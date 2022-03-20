package main

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"
)

func Match(first string, second string, updateChan <-chan bool, quitChan <-chan bool, stats chan<- Stats) {
	teamOddChannel := make(chan *Ball)
	teamEvenChannel := make(chan *Ball)
	whistle := make(chan bool)
	play := make(chan Play)

	var players [10]*Player
	teamOdd := Team{
		Name:         first,
		Score:        0,
		attackCount:  0,
		teamBallChan: teamOddChannel,
	}
	teamEven := Team{
		Name:         second,
		Score:        0,
		attackCount:  0,
		teamBallChan: teamEvenChannel,
	}

	for i := range players {
		var name string
		var team *Team
		if i%2 == 0 {
			name = fmt.Sprint("Player", i)
			team = &teamEven
		} else {
			name = fmt.Sprint("Player", i)
			team = &teamOdd
		}
		players[i] = &Player{
			Name:           name,
			Team:           team,
			Assists:        0,
			Points2Attempt: 0,
			Points2Count:   0,
			Points3Attempt: 0,
			Points3Count:   0,
		}
	}

	for _, player := range players {
		go player.Play(whistle, play)
	}

	ball := NewBall(whistle, teamOddChannel, teamEvenChannel, play)

	go ball.Bounce()

	go func() {
		gameOver := time.After(GameTime)
		select {
		case <-gameOver:
			log.Println("Whistle blown, game over.")
			for i := 0; i < 10+1; i++ {
				whistle <- true
			}
			log.Println("Whistle blower quits")
			close(teamOddChannel)
			close(teamEvenChannel)
			close(play)
			close(whistle)
		}
	}()

	teamOddChannel <- &ball //start the game

	for {
		select {
		case <-quitChan:
			log.Println(first, " vs ", second, " ended updates.")
			return
		case <-updateChan:
			playerStats := make(map[string]PlayerStats)

			for _, player := range players {
				st := PlayerStats{}
				st.Assists = atomic.LoadUint64(&player.Assists)
				st.Points2Rate = fmt.Sprintf("%d/%d", atomic.LoadUint64(&player.Points2Count), atomic.LoadUint64(&player.Points2Attempt))
				st.Points3Rate = fmt.Sprintf("%d/%d", atomic.LoadUint64(&player.Points3Count), atomic.LoadUint64(&player.Points3Attempt))

				playerStats[player.Name] = st
			}

			stats <- Stats{
				Name:        fmt.Sprint(first, " vs ", second),
				AttackCount: atomic.LoadUint64(&ball.attackCount),
				TotalScore:  teamOdd.Score + teamEven.Score,
				PlayerStats: playerStats,
			}
		}
	}

}
