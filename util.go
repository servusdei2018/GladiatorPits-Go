/*
 * Copyright (C) 2020  The GladiatorPits-Go Authors
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
	"math/rand"
	"strings"
)

// Determine what version of "people" to use based on an amount of people present
func plurality(i int) string {
	if i == 1 {
		// E.g. 1 person
		return "person"
	} else { return "people" } // E.g. 0 people, 5 people...
}

// Return a string version of a stat
func stat2String(s int) string {
	switch s {
	case STR:
		return "strength"
	case DEX:
		return "dexterity"
	case STA:
		return "stamina"
	case SIZ:
		return "size"
	case WIT:
		return "wit"
	}; return ""
}

// Send a message to everyone, optionally excluding the crowd
func BroadcastAll(s string, nocrowd bool) {
	for _, p := range Players {
		if !nocrowd || p.Status != CROWD {
			if e := p.Send(s); e != nil { p.Disconnect() }
		}
	}
}

// Send a message to everyone, excluding one Player and optionally excluding the crowd
func BroadcastAllEx(s string, nocrowd bool, exclude string) {
	for id, p := range Players {
		if (!nocrowd || p.Status != CROWD) && id != exclude {
			if e := p.Send(s); e != nil { p.Disconnect() }
		}
	}
}


// Send a message to a particular status
func BroadcastStatus(s string, status int) {
	for _, p := range Players {
		if p.Status == status {
			if e := p.Send(s); e != nil { p.Disconnect() }
		}
	}
}

// Send a message to a particular status excluding one Player
func BroadcastStatusEx(s string, status int, exclude string) {
	for id, p := range Players {
		if p.Status == status && id != exclude {
			if e := p.Send(s); e != nil { p.Disconnect() }
		}
	}
}

// Find a Player, return his id and whether he was found
func FindPlayer(n string) (string, bool) {
	n=strings.ToLower(n)
	for id, p := range Players {
		if strings.ToLower(p.Name) == n {
			return id, true
		}
	}
	return "", false
}

// Remove a Player
func RemovePlayer(p *Player) {
	delete(Players, p.ID)
}

// Function to add randomness to fighting
func rnd() uint16 {
	return uint16(rand.Int31n(10))
}
