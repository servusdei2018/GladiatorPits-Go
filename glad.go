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
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// Configuration constants
const (
	// Port on which to run the MUD
	PORT = ":3000"
	// Maximum amount of trains allowed
	MAX_TRAIN = 30
	// Shutdown password
	PASSWORD = "topsecret"
)

// Global variables
var (
	// List of all players
	Players = make(map[string]*Player)
	// Player mutex
	PlayersLock sync.Mutex
	// MUD socket
	Mud net.Listener
	// Shutdown
	Shutdown sdi
	// Fight ticker
	Tick time.Time
)

// A shutdown interface
type sdi struct {
	// Whether the MUD is to be shutdown
	Shutdown bool
	// Who shutdown the MUD
	By string
	// Mutex for concurrency
	sync.Mutex
}

// Perform initialization
func init() {
	var e error
	Mud, e = net.Listen("tcp", PORT)
	if e != nil {
		log.Fatal("couldn't open socket,", e)
	}
}

// MUD entry point
func main() {
	log.Printf("Mud started on port %v.\n", PORT)
	// Accept new connections concurrently
	go listen()
	for {
		// Handle fighting
		if time.Now().After(Tick) {
			Fights()
			Tick = time.Now().Add(1*time.Second)
		}
		// Handle shutdowns
		if Shutdown.Shutdown {
			BroadcastAll(fmt.Sprintf("[Shutdown by %s]", Shutdown.By), false)
			log.Printf("Shutdown by %s.\n", Shutdown.By)
			break
		}
	}
}

// Accept new connections
func listen() {
	for {
		// Grab a new Conn
		c, e := Mud.Accept()
		if e != nil {
			log.Fatal("error accepting connection,", e)
		}
		// Append him to the Players list
		PlayersLock.Lock()
		n := NewPlayer(c)
		Players[n.ID] = &n
		log.Printf("Connection from %v", n.ID)
		// Handle his input concurrently
		go Handle(n.ID)
		if e = n.Send("Welcome to the Gladiator Pits!"); e != nil {
			n.Disconnect()
		}
		PlayersLock.Unlock()
	}
}

// Handle fighting
func Fights() {
	PlayersLock.Lock()
	defer PlayersLock.Unlock()
	for _, p := range Players {
		p.Spd++
		if p.Status == FIGHTING {
			p.Spd++
			if  p.Spd > p.Speed()+3 {
				p.Spd = 0
				opp, ok := Players[p.Opponent]
				// The opponent has quit, claim victory
				if !ok {
					p.Status = GLADIATOR; p.Kills++; p.Wins++
					if e := p.Send("Your opponent has vanished into the crowd."); e != nil {
						p.Disconnect()
					}
					continue
				}
				var szYou, szThem string
				// Calculate attack success
				result := p.Attack()+rnd()+rnd()-opp.Defence()-rnd()
				// Calculate damage
				dam := p.Damage()

				if result < 1 {
					szYou = "miss"; szThem = "misses"
					dam = 0
				} else if result < 6 {
					szYou = "punch"; szThem = "punches"
					dam /= 2
				} else {
					szYou = "kick"; szThem = "kicks"
				}
				// Send combat messages
				if e := p.Send(fmt.Sprintf("You %v your opponent!", szYou)); e != nil {
					p.Disconnect()
				}
				if e := opp.Send(fmt.Sprintf("Your opponent %v you!", szThem)); e != nil {
					opp.Disconnect()
				}
				// Inflict damage
				opp.Dam += dam
				// Check for death
				if opp.Dam > opp.Health() {
					BroadcastAll(fmt.Sprintf("[%v has killed %v in the arena!]", p.Name, opp.Name), false)
					// Reward the winner
					p.Wins++; p.Kills++; p.Status = GLADIATOR
					// Punish the loser
					opp.Losses++; opp.Status = GLADIATOR
				}
			}
		}
	}
}

// Handle a Player
func Handle(pid string) {
	p := Players[pid]
	r := bufio.NewReader(p.Conn)
	for {
		c, e := r.ReadString('\n')
		if e != nil {
			// Disconnect on error
			PlayersLock.Lock(); p.Disconnect(); PlayersLock.Unlock(); return
		} else {
			PlayersLock.Lock()
			if e = p.Process(c); e != nil {
				// Disconnect on error
				p.Disconnect(); PlayersLock.Unlock(); return
			}; PlayersLock.Unlock()
		}
	}
}
