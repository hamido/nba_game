package main

import (
	"log"
	"sync/atomic"
	"time"
)

type Team struct {
	Name         string
	Score        uint64
	attackCount  uint64
	teamBallChan <-chan *Ball
}

type Player struct {
	Name           string
	Team           *Team
	Assists        uint64
	Points2Attempt uint64
	Points2Count   uint64
	Points3Attempt uint64
	Points3Count   uint64
}

type Play interface {
	apply(b *Ball)
}

type Basket struct {
	player *Player
}

func (p *Basket) apply(b *Ball) {
	if Random() {
		atomic.AddUint64(&p.player.Points2Attempt, 1)
		b.Basket(p.player, 2)
	} else {
		atomic.AddUint64(&p.player.Points3Attempt, 1)
		b.Basket(p.player, 3)
	}
}

type Pass struct {
	passer *Player
}

func (p *Pass) apply(b *Ball) {
	b.Pass(p.passer)
}

func (p *Player) Play(whistle <-chan bool, play chan<- Play) {
	for {
		select {
		case ball := <-p.Team.teamBallChan:
			if ball != nil {
				log.Println(p.Name, " has the ball")
				time.Sleep(time.Second * 2)
				if Random() {
					play <- &Basket{
						player: p,
					}
				} else {
					play <- &Pass{
						passer: p,
					}
				}
				log.Println(p.Name, " passed the ball")
			}
		case <-whistle:
			log.Println(p.Name, " quits.")
			return
		}
	}
}
