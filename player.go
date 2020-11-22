/*
 * Copyright (C) 2020  The Gladiator Pits Go Authors
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"fmt"
	"net"
)

const (
	STR = iota
	DEX
	STA
	SIZ
	WIT
	CROWD
	CITIZEN
	TRAINING
	GLADIATOR
	CHALLENGER
	FIGHTING
)

type Player struct {
	// Player's name
	Name string
	// ID
	ID string
	// Player's connection
	Conn net.Conn
	// Was he disconnected
	Disconnected bool
	// Player's stats
	Stats map[int]uint16
	// Player's status
	Status int
	// Statistics
	Wins uint64
	Losses uint64
	Kills uint64
	// Fighting variables
	Dam uint16
	Spd uint16
	// Player we're challenged by/fighting
	Opponent string
}

// Create and return a new Player
func NewPlayer(c net.Conn) (Player) {
	return Player{Name: "Someone in the crowd", ID: c.RemoteAddr().String(), Conn: c, Stats: map[int]uint16{STR: uint16(1), DEX: uint16(1), STA: uint16(1), SIZ: uint16(1), WIT: uint16(1),}, Status: CROWD, Wins: uint64(0), Losses: uint64(0), Kills: uint64(0), Dam: uint16(0), Spd: uint16(0), Opponent: ""}
}

// Return health
func (p *Player) Health() uint16 {
	return p.Stats[STR]+(p.Stats[STA]*7)+(p.Stats[SIZ]*3)
}

// Return attack
func (p *Player) Attack() uint16 {
	return (p.Stats[DEX]*2)+p.Stats[SIZ]
}

// Return defence
func (p *Player) Defence() uint16 {
	return (p.Stats[WIT]*2)+p.Stats[DEX]
}

// Return damage
func (p *Player) Damage() uint16 {
	return (p.Stats[STR]*3)+p.Stats[SIZ]
}

// Return speed
func (p *Player) Speed() uint16 {
	return 10-(((p.Stats[WIT]*2)+p.Stats[DEX])/3)
}

// Return whether a player may still train
func (p *Player) CanTrain() bool {
	return MAX_TRAIN > p.Stats[STR]+p.Stats[DEX]+p.Stats[STA]+p.Stats[SIZ]+p.Stats[WIT]
}

// Return how many stats left to train
func (p *Player) StatsLeft() uint16 {
	return uint16(30)-(p.Stats[STR]+p.Stats[DEX]+p.Stats[STA]+p.Stats[SIZ]+p.Stats[WIT])
}

// Train one stat
func (p *Player) DoTrain(s int) error {
	if !p.CanTrain() {
		return p.Send("No remaining points.")
	}
	if p.Stats[s] >= 9 {
		return p.Send(fmt.Sprintf("%v already at max.", stat2String(s)))
	} else {
		p.Stats[s]++
		return p.Send(fmt.Sprintf("%v trained to %v (%d points left).", stat2String(s), p.Stats[s], p.StatsLeft()))
	}
}

// Send text to a player
func (p *Player) Send(s string) error {
	_, e := fmt.Fprintln(p.Conn, s)
	return e
}

// Disconnect a player
func (p *Player) Disconnect() {
	p.Conn.Close()
	RemovePlayer(p)
	if p.Status != CROWD {
		BroadcastAll(fmt.Sprintf("%v vanishes into the crowd.", p.Name), true)
	}
}
