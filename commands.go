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
	"fmt"
	"strings"
)

// Global variables
var (
	CommandMap = map[string]func(p *Player, s []string) error{
		"quit": CmdQuit,
		"who": CmdWho,
		"commands": CmdCommands,
		"shutdown": CmdShutdown,
		"say": CmdSay,
		"chat": CmdChat,
		"emote": CmdEmote,
		"create": CmdCreate,
		"score": CmdScore,
		"train": CmdTrain,
		"str": CmdStr,
		"dex": CmdDex,
		"sta": CmdSta,
		"siz": CmdSiz,
		"wit": CmdWit,
		"leave": CmdLeave,
		"challenge": CmdChallenge,
		"accept": CmdAccept,
	}
)

// Process commands
func (p *Player) Process(s string) error {
	tk := strings.Split(strings.ReplaceAll(strings.ReplaceAll(s, "\r", ""), "\n", ""), " ")
	handler, ok := CommandMap[strings.ToLower(tk[0])]
	if !ok || p.Status == FIGHTING {
		return p.Send(fmt.Sprintf("Unrecognised command '%s'. Type 'commands' to list available options.", tk[0]))
	} else {
		return handler(p, tk)
	}
}

// Quits the MUD
func CmdQuit(p *Player, s []string) error {
	p.Disconnect(); return nil
}

// Lists players in the MUD
func CmdWho(p *Player, s []string) error {
	vis := 0
	for _, q := range Players {
		if q.Status != CROWD {
			vis++
			if e:=p.Send(fmt.Sprintf("%v (Won:%v Lost:%v Kills:%v)", q.Name, q.Wins, q.Losses, q.Kills)); e != nil { p.Disconnect() }
		}
	}
	return p.Send(fmt.Sprintf("Total of %v visible %v, with %v other %v in the crowd.", vis, plurality(vis), len(Players)-vis, plurality(len(Players)-vis)))
}

// Lists available commands
func CmdCommands(p *Player, s []string) error {
	switch p.Status {
	case CROWD:
		return p.Send("quit who commands say emote create")
	case CITIZEN:
		return p.Send("quit who commands shutdown say chat emote score train")
	case TRAINING:
		return p.Send("quit who commands shutdown say chat emote score str dex sta siz wit leave")
	case GLADIATOR:
		return p.Send("quit who commands say chat emote score challenge accept")
	case CHALLENGER:
		return p.Send("quit who commands say chat emote score challenge")
	}
	return nil
}

// Shuts down the MUD
func CmdShutdown(p *Player, s []string) error {
	if p.Status == CROWD { return InvCmd(p, s) }
	if len(s) == 1 {
		return p.Send("Syntax: shutdown <password>")
	} else {
		if s[1] == PASSWORD {
			Shutdown.Lock()
			Shutdown.Shutdown = true
			Shutdown.By = p.Name
			Shutdown.Unlock()
		} else {
			return p.Send("Incorrect password.")
		}
	}
	return nil
}

// Say something
func CmdSay(p *Player, s []string) error {
	if len(s) == 1 {
		return p.Send("Syntax: say <sentence>")
	} else {
		BroadcastStatusEx(fmt.Sprintf("%s says '%s'", p.Name, strings.Join(s[1:], " ")), p.Status, p.ID)
		return p.Send(fmt.Sprintf("You say '%s'", strings.Join(s[1:], " ")))
	}
}

// Chat something
func CmdChat(p *Player, s []string) error {
	if p.Status == CROWD { return nil }
	if len(s) == 1 {
		return p.Send("Syntax: chat <sentence>")
	} else {
		BroadcastAllEx(fmt.Sprintf("%s chats '%s'", p.Name, strings.Join(s[1:], " ")), true, p.ID)
		return p.Send(fmt.Sprintf("You chat '%s'", strings.Join(s[1:], " ")))
	}
}

// Emote something
func CmdEmote(p *Player, s []string) error {
	if len(s) == 1 {
		return p.Send("Syntax: emote <action>")
	} else {
		BroadcastStatus(fmt.Sprintf("%s %s", p.Name, strings.Join(s[1:], " ")), p.Status)
	}
	return nil
}

// Create a character
func CmdCreate(p *Player, s []string) error {
	if p.Status != CROWD { return InvCmd(p, s) }
	if len(s) == 1 {
		return p.Send("Syntax: create <name>")
	}
	if len(s[1])>15 || len(s[1])<3 {
		return p.Send("Names must be 3-15 letters long.")
	}
	if _, found := FindPlayer(s[1]); found {
		return p.Send("That name is in use.")
	}
	p.Name = strings.Title(s[1])
	p.Status = CITIZEN
	BroadcastAllEx(fmt.Sprintf("%s steps from the crowd.", p.Name), false, p.ID)
	return p.Send("You step from the crowd.")
}

// Display score
func CmdScore(p *Player, s []string) error {
	if p.Status == CROWD { return InvCmd(p, s) }
	return p.Send(
		fmt.Sprintf(
`<<<<<===--------[ Score ]--------===>>>>>
	Name:%v
	Win-Loss-Kill: %v - %v - %v
<<<<<===-------------------------===>>>>>
	Str:%v Dex:%v Sta:%v Siz:%v Wit:%v
	Att:%v Def:%v Dam:%v
	Wounds:%v (%v) Speed:%v
<<<<<===-------------------------===>>>>>`,
p.Name, p.Wins, p.Losses, p.Kills, p.Stats[STR], p.Stats[DEX],
p.Stats[STA], p.Stats[SIZ], p.Stats[WIT], p.Attack(), p.Defence(), p.Damage(),
p.Health()-p.Dam, p.Health(), p.Speed()))
}

// Enter training
func CmdTrain(p *Player, s []string) error {
	if p.Status != CITIZEN { return InvCmd(p, s) }
	p.Status = TRAINING
	BroadcastStatusEx(fmt.Sprintf("%s walks to the training room.", p.Name), CITIZEN, p.ID)
	BroadcastStatusEx(fmt.Sprintf("%s enters the training room.", p.Name), TRAINING, p.ID)
	return p.Send("You walk to the training room.  To get back, type 'leave'.")
}

// Train strength
func CmdStr(p *Player, s []string) error {
	if p.Status != TRAINING { return InvCmd(p, s) }
	return p.DoTrain(STR)
}

// Train dexterity
func CmdDex(p *Player, s []string) error {
	if p.Status != TRAINING { return InvCmd(p, s) }
	return p.DoTrain(DEX)
}

// Train stamina
func CmdSta(p *Player, s []string) error {
	if p.Status != TRAINING { return InvCmd(p, s) }
	return p.DoTrain(STA)
}

// Train size
func CmdSiz(p *Player, s []string) error {
	if p.Status != TRAINING { return InvCmd(p, s) }
	return p.DoTrain(SIZ)
}

// Train wit
func CmdWit(p *Player, s []string) error {
	if p.Status != TRAINING { return InvCmd(p, s) }
	return p.DoTrain(WIT)
}

// Leave training
func CmdLeave(p *Player, s []string) error {
	if p.Status != TRAINING { return InvCmd(p, s) }
	// If he cannot train anymore, he's a gladiator
	if !p.CanTrain() {
		p.Status = GLADIATOR
	} else {
		p.Status = CITIZEN
	}
	BroadcastStatusEx(fmt.Sprintf("%s leaves the training room.", p.Name), TRAINING, p.ID)
	BroadcastStatusEx(fmt.Sprintf("%s arrives from the training room.", p.Name), CITIZEN, p.ID)
	BroadcastStatusEx(fmt.Sprintf("%s arrives from the training room.", p.Name), GLADIATOR, p.ID)
	return p.Send("You leave the training room.")
}

// Challenge a player
func CmdChallenge(p *Player, s []string) error {
	if p.Status != GLADIATOR && p.Status != CHALLENGER { return InvCmd(p, s) }
	if len(s) == 1 {
		return p.Send("Syntax: challenge <gladiator>")
	}
	pid, ok := FindPlayer(s[1])
	if !ok {
		return p.Send(fmt.Sprintf("No such gladiator as '%v'.", s[1]))
	}
	opp, okk := Players[pid]
	if !okk { return nil }
	if opp.Status != GLADIATOR && opp.Status != CHALLENGER {
		return p.Send("They're no gladiator!")
	} else {
		p.Status = CHALLENGER
		p.Opponent = opp.ID
		opp.Status = GLADIATOR
		opp.Opponent = p.ID
		if e := opp.Send(fmt.Sprintf("%v challenges you to a fight.", p.Name)); e != nil {
			opp.Disconnect(); return nil
		}
		return p.Send(fmt.Sprintf("You challenge %v to a fight.", opp.Name))
	}
}

// Accept a challenge
func CmdAccept(p *Player, s []string) error {
	if p.Status != GLADIATOR && p.Status != CHALLENGER { return InvCmd(p, s) }
	opp, ok := Players[p.Opponent]
	if !ok {
		return p.Send("You've not been challenged.")
	} else {
		if e := opp.Send(fmt.Sprintf("%v accepts your challenge!", p.Name)); e != nil {
			opp.Disconnect(); return nil
		}
		p.Dam = 0; p.Spd = 0; opp.Dam = 0; opp.Spd = 0
		p.Status = FIGHTING; opp.Status = FIGHTING
		BroadcastAll(fmt.Sprintf("[%v and %v have entered the arena]", p.Name, opp.Name),false)
		return p.Send("Ok.")
	}
	return nil
}

// A command isn't available for a player at the moment
func InvCmd(p *Player, s []string) error {
	return p.Send(fmt.Sprintf("Unrecognised command '%s'. Type 'commands' to list available options.", s[0]))
}
