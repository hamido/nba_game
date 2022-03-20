package main

import (
	"log"
	"sync/atomic"
	"time"
)

type Ball struct {
	passer        *Player
	attackTimeout <-chan time.Time
	attackCount   uint64
	Whistle       <-chan bool
	OffensiveTeam chan<- *Ball
	DefensiveTeam chan<- *Ball
	play          <-chan Play
	gameOver      bool
}

func NewBall(whistle <-chan bool, offensiveTeam chan<- *Ball, defensiveTeam chan<- *Ball, play <-chan Play) Ball {
	return Ball{
		passer:        nil,
		attackTimeout: time.After(AttackTime),
		Whistle:       whistle,
		OffensiveTeam: offensiveTeam,
		DefensiveTeam: defensiveTeam,
		play:          play,
	}
}

func (b *Ball) Bounce() {
	for {
		select {
		case <-b.Whistle:
			<-b.play // take the ball from player so that he can quit when game is over
			return
		case <-b.attackTimeout:
			log.Println("Attack timeout violated! Switching teams")
			b.OffensiveTeam, b.DefensiveTeam = b.DefensiveTeam, b.OffensiveTeam
			atomic.AddUint64(&b.attackCount, 1)
			b.attackTimeout = time.After(AttackTime)
		case play := <-b.play:
			play.apply(b)
			log.Println("Ball back into game")
			b.OffensiveTeam <- b
		}
	}
}

func (b *Ball) Pass(p *Player) {
	if b.passer != p { // in case same player takes the ball
		b.passer = p
	}

}

func (b *Ball) Basket(p *Player, score int) {
	if Random() { // if basket
		if b.passer != nil && b.passer.Team == p.Team { // if basket and on the same team this is an assist
			atomic.AddUint64(&b.passer.Assists, 1)
		}
		if score == 3 {
			atomic.AddUint64(&p.Points3Count, 1)
			atomic.AddUint64(&p.Team.Score, 3)
		}
		if score == 2 {
			atomic.AddUint64(&p.Points2Count, 1)
			atomic.AddUint64(&p.Team.Score, 2)
		}
		// reset attack time
		log.Println("Basket!")
		b.attackTimeout = time.After(AttackTime)
		atomic.AddUint64(&b.attackCount, 1)
	}
}
