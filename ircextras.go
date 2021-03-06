package ircchans

import (
	"os"
	"fmt"
	"strings"
	"strconv"
	"time"
)
//irc reply message types
var replies = map[string]string{
	"ERR_NOSUCHNICK":       "401",
	"ERR_NOSUCHSERVER":     "402",
	"ERR_NOSUCHCHANNEL":    "403",
	"ERR_CANNOTSENDTOCHAN": "404",
	"ERR_TOOMANYCHANNELS":  "405",
	"ERR_WASNOSUCHNICK":    "406",
	"ERR_TOOMANYTARGETS":   "407",
	"ERR_NOORIGIN":         "409",
	"ERR_NORECIPIENT":      "411",
	"ERR_NOTEXTTOSEND":     "412",
	"ERR_NOTOPLEVEL":       "413",
	"ERR_WILDTOPLEVEL":     "414",
	"ERR_UNKNOWNCOMMAND":   "421",
	"ERR_NOMOTD":           "422",
	"ERR_NOADMININFO":      "423",
	"ERR_FILEERROR":        "424",
	"ERR_NONICKNAMEGIVEN":  "431",
	"ERR_ERRONEUSNICKNAME": "432",
	"ERR_NICKNAMEINUSE":    "433",
	"ERR_NICKCOLLISION":    "436",
	"ERR_USERNOTINCHANNEL": "441",
	"ERR_NOTONCHANNEL":     "442",
	"ERR_USERONCHANNEL":    "443",
	"ERR_NOLOGIN":          "444",
	"ERR_SUMMONDISABLED":   "445",
	"ERR_USERSDISABLED":    "446",
	"ERR_NOTREGISTERED":    "451",
	"ERR_NEEDMOREPARAMS":   "461",
	"ERR_ALREADYREGISTRED": "462",
	"ERR_NOPERMFORHOST":    "463",
	"ERR_PASSWDMISMATCH":   "464",
	"ERR_YOUREBANNEDCREEP": "465",
	"ERR_KEYSET":           "467",
	"ERR_CHANNELISFULL":    "471",
	"ERR_UNKNOWNMODE":      "472",
	"ERR_INVITEONLYCHAN":   "473",
	"ERR_BANNEDFROMCHAN":   "474",
	"ERR_BADCHANNELKEY":    "475",
	"ERR_NOPRIVILEGES":     "481",
	"ERR_CHANOPRIVSNEEDED": "482",
	"ERR_CANTKILLSERVER":   "483",
	"ERR_NOOPERHOST":       "491",
	"ERR_UMODEUNKNOWNFLAG": "501",
	"ERR_USERSDONTMATCH":   "502",
	"RPL_NONE":             "300",
	"RPL_USERHOST":         "302",
	"RPL_ISON":             "303",
	"RPL_AWAY":             "301",
	"RPL_UNAWAY":           "305",
	"RPL_NOWAWAY":          "306",
	"RPL_WHOISUSER":        "311",
	"RPL_WHOISSERVER":      "312",
	"RPL_WHOISOPERATOR":    "313",
	"RPL_WHOISIDLE":        "317",
	"RPL_ENDOFWHOIS":       "318",
	"RPL_WHOISCHANNELS":    "319",
	"RPL_WHOWASUSER":       "314",
	"RPL_ENDOFWHOWAS":      "369",
	"RPL_LISTSTART":        "321",
	"RPL_LIST":             "322",
	"RPL_LISTEND":          "323",
	"RPL_CHANNELMODEIS":    "324",
	"RPL_NOTOPIC":          "331",
	"RPL_TOPIC":            "332",
	"RPL_INVITING":         "341",
	"RPL_SUMMONING":        "342",
	"RPL_VERSION":          "351",
	"RPL_WHOREPLY":         "352",
	"RPL_ENDOFWHO":         "315",
	"RPL_NAMREPLY":         "353",
	"RPL_ENDOFNAMES":       "366",
	"RPL_LINKS":            "364",
	"RPL_ENDOFLINKS":       "365",
	"RPL_BANLIST":          "367",
	"RPL_ENDOFBANLIST":     "368",
	"RPL_INFO":             "371",
	"RPL_ENDOFINFO":        "374",
	"RPL_MOTDSTART":        "375",
	"RPL_MOTD":             "372",
	"RPL_ENDOFMOTD":        "376",
	"RPL_YOUREOPER":        "381",
	"RPL_REHASHING":        "382",
	"RPL_TIME":             "391",
	"RPL_USERSSTART":       "392",
	"RPL_USERS":            "393",
	"RPL_ENDOFUSERS":       "394",
	"RPL_NOUSERS":          "395",
	"RPL_TRACELINK":        "200",
	"RPL_TRACECONNECTING":  "201",
	"RPL_TRACEHANDSHAKE":   "202",
	"RPL_TRACEUNKNOWN":     "203",
	"RPL_TRACEOPERATOR":    "204",
	"RPL_TRACEUSER":        "205",
	"RPL_TRACESERVER":      "206",
	"RPL_TRACENEWTYPE":     "208",
	"RPL_TRACELOG":         "261",
	"RPL_STATSLINKINFO":    "211",
	"RPL_STATSCOMMANDS":    "212",
	"RPL_STATSCLINE":       "213",
	"RPL_STATSNLINE":       "214",
	"RPL_STATSILINE":       "215",
	"RPL_STATSKLINE":       "216",
	"RPL_STATSYLINE":       "218",
	"RPL_ENDOFSTATS":       "219",
	"RPL_STATSLLINE":       "241",
	"RPL_STATSUPTIME":      "242",
	"RPL_STATSOLINE":       "243",
	"RPL_STATSHLINE":       "244",
	"RPL_UMODEIS":          "221",
	"RPL_LUSERCLIENT":      "251",
	"RPL_LUSEROP":          "252",
	"RPL_LUSERUNKNOWN":     "253",
	"RPL_LUSERCHANNELS":    "254",
	"RPL_LUSERME":          "255",
	"RPL_ADMINME":          "256",
	"RPL_ADMINEMAIL":       "259"}

func timeout(lag int64) int64 {
	t := lag * 3
	if t > second*15 {
		return second * 15
	}
	return t
}

func (n *Network) Register() os.Error {
	var err os.Error
	welcome := make(chan *IrcMessage, 1)
	if err = n.Listen.RegListener("001", "register", welcome); err != nil {
		return os.NewError("Couldn't register listener for welcome messages (001)")
	}
	defer n.Listen.DelListener("001", "register")
	if n.password != "" {
		err = n.Pass()
		if err != nil {
			return os.NewError("Couldn't register with password")
		}
	}
	nret := make(chan bool, 1)
	go func(n *Network, ret chan bool) {
		_, err = n.Nick(n.nick)
		i := 0
		for err != nil {
			if i > 8 {
				ret <- false
				return
			}
			n.nick = fmt.Sprintf("_%s", n.nick)
			_, err = n.Nick(n.nick)
			i++
		}
		ret <- true
		return
	}(n, nret)
	//TODO: reglistener for cmd 001 (welcome) which means user and nick commands were successful
	n.user, err = n.User(n.user)
	if err != nil {
		return os.NewError("Unable to register username")
	}
	select {
	case ok := <-nret:
		if !ok {
			return os.NewError("Failed to acquire any alternate nick")
		}
	case <-welcome:
		return nil
	}
	return nil
}

func (n *Network) Pass() os.Error {
	t := strconv.Itoa64(time.Nanoseconds())
	myreplies := []string{"ERR_NEEDMOREPARAMS", "ERR_ALREADYREGISTRED"}
	var err os.Error
	repch := make(chan *IrcMessage)
	for _, rep := range myreplies {
		if err := n.Listen.RegListener(replies[rep], t, repch); err != nil {
			err = os.NewError(fmt.Sprintf("Couldn't authenticate with password, exiting: %s", err.String()))
		}
	}
	ticker := time.NewTicker(timeout(n.lag))
	defer func(myreplies []string, t string, tick *time.Ticker) {
		for _, rep := range myreplies {
			n.Listen.DelListener(replies[rep], t)
		}
		tick.Stop()
		return
	}(myreplies, t, ticker)
	n.queueOut <- &IrcMessage{"", "PASS", []string{n.password}}
	select {
	case msg := <-repch:
		if msg.Cmd == replies["ERR_NEEDMOREPARAMS"] {
			err = os.NewError(fmt.Sprintf("Need more parameters for password: %s", msg.String()))
		}
		break
	case <-ticker.C:
		break
	}
	return err
}


func (n *Network) GetNick() string {
	return n.nick
}

func (n *Network) Nick(newnick string) (string, os.Error) {
	t := strconv.Itoa64(time.Nanoseconds())
	ticker := time.NewTicker(timeout(n.lag))
	defer ticker.Stop()
	myreplies := []string{"ERR_NONICKNAMEGIVEN", "ERR_ERRONEUSNICKNAME", "ERR_NICKNAMEINUSE", "ERR_NICKCOLLISION"}
	if newnick == "" {
		return n.nick, os.NewError("Empty nicknames are not accepted in IRC")
	}
	//TODO: check for correct nick (illegal characters)
	if len(newnick) > 9 {
		newnick = newnick[:9]
	}
	repch := make(chan *IrcMessage)
	defer func(myreplies []string, t string) {
		for _, rep := range myreplies {
			n.Listen.DelListener(replies[rep], t)
		}
		return
	}(myreplies, t)
	for _, rep := range myreplies {
		if err := n.Listen.RegListener(replies[rep], t, repch); err != nil {
			for _, rep := range myreplies {
				n.Listen.DelListener(replies[rep], t)
			}
			return n.nick, os.NewError("Unable to register new listener")
		}
	}
	n.queueOut <- &IrcMessage{"", "NICK", []string{newnick}}
	select {
	case msg := <-repch:
		if msg.Cmd == replies["ERR_ERRONEUSNICKNAME"] || msg.Cmd == replies["ERR_NICKNAMEINUSE"] || msg.Cmd == replies["ERR_NICKCOLLISION"] {
			for key, _ := range replies {
				if replies[key] == msg.Cmd {
					return n.nick, os.NewError(key)
				}
			}
			return n.nick, os.NewError("Unknown error")
		}
	case <-ticker.C:
		break
	}
	n.nick = newnick
	return n.nick, nil
}

func (n *Network) GetUser(newuser string) string {
	return n.user
}

func (n *Network) User(newuser string) (string, os.Error) {
	t := strconv.Itoa64(time.Nanoseconds())
	ticker := time.NewTicker(timeout(n.lag))
	defer ticker.Stop()
	myreplies := []string{"ERR_NEEDMOREPARAMS", "ERR_ALREADYREGISTRED", "RPL_ENDOFMOTD", "ERR_NOTREGISTERED"}
	if newuser == "" {
		return n.user, os.NewError("Can't have an empty user field")
	} else if len(newuser) > 9 {
		newuser = newuser[:9]
	}
	repch := make(chan *IrcMessage)
	defer func(myreplies []string, t string) {
		for _, rep := range myreplies {
			n.Listen.DelListener(replies[rep], t)
		}
		return
	}(myreplies, t)
	for _, rep := range myreplies {
		if err := n.Listen.RegListener(replies[rep], t, repch); err != nil {
			return "", os.NewError(fmt.Sprintf("Couldn't register Listener for %s: %s", replies[rep], err.String()))
		}
	}
	n.queueOut <- &IrcMessage{"", "USER", []string{n.user, "0.0.0.0", "0.0.0.0", n.realname}}
	select {
	case msg := <-repch:
		if msg.Cmd == replies["ERR_NEEDMOREPARAMS"] {
			return n.user, os.NewError("ERR_NEEDMOREPARAMS")
		} else if msg.Cmd == replies["ERR_ALREADYREGISTRED"] {
			return n.user, os.NewError("ERR_ALREADYREGISTRED")
		} else if msg.Cmd == replies["ERR_NOTREGISTERED"] {
			return n.user, os.NewError("ERR_NOTREGISTERED")
		} else if msg.Cmd == replies["RPL_ENDOFMOTD"] {
			return n.user, nil
		}
	case <-ticker.C:
		n.user = newuser
	}
	return n.user, nil
}

func (n *Network) Realname(newrn string) string {
	//TODO: call user from here
	if n.conn == nil {
		//TODO: see User: can we change realname after we are connected? -> if we can change the user after connected
		n.realname = newrn
	}
	return n.realname
}

func (n *Network) GetNetName() string {
	return n.network
}

func (n *Network) NetName(newname string, reason string) (string, os.Error) {
	if newname != "" {
		n.network = newname
		return n.network, n.Reconnect(reason)
	} else {
		return n.network, os.NewError("Empty name")
	}
	return n.network, nil //BUG: why do we need this?
}

func (n *Network) SysOpMe(user, pass string) {
	n.queueOut <- &IrcMessage{"", "OPER", []string{user, pass}}
	//TODO: replies:
	//ERR_NEEDMOREPARAMS              RPL_YOUREOPER
	//ERR_NOOPERHOST                  ERR_PASSWDMISMATCH
	return
}

func (n *Network) Quit(reason string) {
	n.queueOut <- &IrcMessage{"", "QUIT", []string{reason}}
	return
}

func (n *Network) Join(chans []string, keys []string) os.Error { //return: topic, list?
	if len(chans) == 0 {
		return os.NewError("No channels given")
	}
	t := strconv.Itoa64(time.Nanoseconds())
	ticker := time.NewTicker(timeout(n.lag))
	myreplies := []string{"ERR_NEEDMOREPARAMS", "ERR_BANNEDFROMCHAN",
		"ERR_INVITEONLYCHAN", "ERR_BADCHANNELKEY",
		"ERR_CHANNELISFULL", "ERR_BADCHANMASK",
		"ERR_NOSUCHCHANNEL", "ERR_TOOMANYCHANNELS",
		"RPL_TOPIC", "JOIN"}
	for _, ch := range chans {
		if !strings.HasPrefix(ch, "#") && !strings.HasPrefix(ch, "&") && !strings.HasPrefix(ch, "+") && !strings.HasPrefix(ch, "!") {
			return os.NewError(fmt.Sprintf("Channel %s doesn't start with a legal prefix", ch))
		}
		if strings.Contains(ch, string(' ')) || strings.Contains(ch, string(7)) || strings.Contains(ch, ",") {
			return os.NewError(fmt.Sprintf("Channel %s contains illegal characters", ch))
		}
	}
	repch := make(chan *IrcMessage, 10)
	defer func(myreplies []string, t string) {
		for _, rep := range myreplies {
			_, ok := replies[rep]
			if ok {
				n.Listen.DelListener(replies[rep], t)
			} else {
				n.Listen.DelListener(rep, t)
			}
		}
		return
	}(myreplies, t)
	for _, rep := range myreplies {
		_, ok := replies[rep]
		if ok {
			if err := n.Listen.RegListener(replies[rep], t, repch); err != nil {
				return os.NewError(fmt.Sprintf("Couldn't register listener %s: %s", replies[rep], err.String()))
			}
		} else {
			if err := n.Listen.RegListener(rep, t, repch); err != nil {
				return os.NewError(fmt.Sprintf("Couldn't register listener %s: %s", rep, err.String()))
			}
		}
	}
	n.queueOut <- &IrcMessage{"", "JOIN", []string{strings.Join(chans, ","), strings.Join(keys, ",")}}
	joined := 0
	for {
		select {
		case msg := <-repch:
			if msg.Cmd == "JOIN" {
				for _, chn := range chans {
					if msg.Params[0] == chn {
						joined++
						break
					}
				}
			} else {
				for key, _ := range replies {
					if replies[key] == msg.Cmd {
						if key[:3] == "ERR" {
							ticker.Stop()
							return os.NewError(key)
						}
					}
				}
			}
			if joined == len(chans) {
				ticker.Stop()
				return nil
			}
			ticker.Stop()
			ticker = time.NewTicker(timeout(n.lag))
		case <-ticker.C:
			ticker.Stop()
			return os.NewError("Didn't receive join reply")
		}
	}
	ticker.Stop()
	return nil
}

func (n *Network) Part(chans []string, reason string) {
	n.queueOut <- &IrcMessage{"", "PART", []string{strings.Join(chans, ","), reason}}
	//TODO: replies:
	//ERR_NEEDMOREPARAMS              ERR_NOSUCHCHANNEL
	//ERR_NOTONCHANNEL
	return
}

func (n *Network) Mode(target, mode, params string) {
	chmodes := []byte{'o', 'p', 's', 'i', 't', 'n', 'm', 'l', 'b', 'v', 'k'}
	usrmodes := []byte{'i', 's', 'w', 'o'}
	var found bool
	for _, c := range mode { //is it a channel mode?
		found = false
		for _, m := range chmodes {
			if m == byte(c) {
				found = true
				break
			}
		}
		if !found {
			break
		}
	}
	if !found { //maybe it's a user mode?
		for _, c := range mode {
			found := false
			for _, m := range usrmodes {
				if m == byte(c) {
					found = true
					break
				}
			}
			if !found { //neither a channel nor a user mode, don't touch this
				return
				//TODO: return error?
			}
		}
	}
	n.queueOut <- &IrcMessage{"", "MODE", []string{target, mode, params}}
	//TODO: replies:
	//ERR_NEEDMOREPARAMS              RPL_CHANNELMODEIS
	//ERR_CHANOPRIVSNEEDED            ERR_NOSUCHNICK
	//ERR_NOTONCHANNEL                ERR_KEYSET
	//RPL_BANLIST                     RPL_ENDOFBANLIST
	//ERR_UNKNOWNMODE                 ERR_NOSUCHCHANNEL
	//
	//ERR_USERSDONTMATCH              RPL_UMODEIS
	//ERR_UMODEUNKNOWNFLAG
	return
}

func (n *Network) SetTopic(ch, topic string) {
	n.queueOut <- &IrcMessage{"", "TOPIC", []string{ch, topic}}
	//TODO: replies
	//ERR_NEEDMOREPARAMS              ERR_NOTONCHANNEL
	//RPL_NOTOPIC                     RPL_TOPIC
	//ERR_CHANOPRIVSNEEDED
	return
}

func (n *Network) GetTopic(ch string) string {
	n.queueOut <- &IrcMessage{"", "TOPIC", []string{ch}}
	//TODO: replies
	//ERR_NEEDMOREPARAMS              ERR_NOTONCHANNEL
	//RPL_NOTOPIC                     RPL_TOPIC
	//ERR_CHANOPRIVSNEEDED
	return ""
}

func (n *Network) Names(chans []string) {
	n.queueOut <- &IrcMessage{"", "NAMES", []string{strings.Join(chans, ",")}}
	//TODO: replies:
	//RPL_NAMREPLY                    RPL_ENDOFNAMES
	return
}

func (n *Network) List(chans []string, server string) {
	msg := &IrcMessage{"", "LIST", []string{}}
	if len(chans) > 0 {
		msg.Params = append(msg.Params, strings.Join(chans, ","))
	}
	if server != "" {
		msg.Params = append(msg.Params, server)
	}
	n.queueOut <- msg
	//TODO: replies:
	//ERR_NOSUCHSERVER                RPL_LISTSTART
	//RPL_LIST                        RPL_LISTEND
	return
}

func (n *Network) Invite(target, ch string) {
	n.queueOut <- &IrcMessage{"", "INVITE", []string{target, ch}}
	//TODO: replies:
	//ERR_NEEDMOREPARAMS              ERR_NOSUCHNICK
	//ERR_NOTONCHANNEL                ERR_USERONCHANNEL
	//ERR_CHANOPRIVSNEEDED
	//RPL_INVITING                    RPL_AWAY
	return
}

func (n *Network) Kick(ch, target, reason string) {
	n.queueOut <- &IrcMessage{"", "KICK", []string{ch, target, reason}}
	//TODO: replies:
	//ERR_NEEDMOREPARAMS              ERR_NOSUCHCHANNEL
	//ERR_BADCHANMASK                 ERR_CHANOPRIVSNEEDED
	//ERR_NOTONCHANNEL
	return
}

func (n *Network) Privmsg(target []string, msg string) os.Error { //BUG: make privmsg hack up messages that are too long
	t := strconv.Itoa64(time.Nanoseconds())
	ticker := time.NewTicker(timeout(n.lag))
	myreplies := []string{"ERR_NORECIPIENT", "ERR_NOTEXTTOSEND",
		"ERR_CANNOTSENDTOCHAN", "ERR_NOTOPLEVEL",
		"ERR_WILDTOPLEVEL", "ERR_TOOMANYTARGETS",
		"ERR_NOSUCHNICK", "RPL_AWAY"}
	repch := make(chan *IrcMessage, 10)
	for _, rep := range myreplies {
		if err := n.Listen.RegListener(replies[rep], t, repch); err != nil {
			return os.NewError(fmt.Sprintf("Couldn't register nick %s: %s", replies[rep], err.String()))
		}
	}
	defer func(myreplies []string, t string) {
		for _, rep := range myreplies {
			n.Listen.DelListener(replies[rep], t)
		}
		return
	}(myreplies, t)
	n.queueOut <- &IrcMessage{"", "PRIVMSG", []string{strings.Join(target, ","), msg}}
	for {
		select {
		case msg := <-repch:
			for key, _ := range replies {
				if replies[key] == msg.Cmd && key[:3] == "ERR" {
					ticker.Stop()
					return os.NewError(key)
				}
			}
			ticker.Stop()
			ticker = time.NewTicker(timeout(n.lag))
		case <-ticker.C:
			ticker.Stop()
			return nil
		}
	}
	ticker.Stop()
	return nil
}

func (n *Network) Notice(target, text string) { //BUG: make notice hack up messages that are too long
	n.queueOut <- &IrcMessage{"", "NOTICE", []string{target, text}}
	//TODO: replies:
	//ERR_NORECIPIENT                 ERR_NOTEXTTOSEND
	//ERR_CANNOTSENDTOCHAN            ERR_NOTOPLEVEL
	//ERR_WILDTOPLEVEL                ERR_TOOMANYTARGETS
	//ERR_NOSUCHNICK
	//RPL_AWAY
	return
}

func (n *Network) Who(target string) {
	n.queueOut <- &IrcMessage{"", "WHO", []string{target}}
	//TODO: replies:
	//ERR_NOSUCHSERVER
	//RPL_WHOREPLY                    RPL_ENDOFWHO
	return
}

func (n *Network) Whois(target []string, server string) (map[string][]string, os.Error) { //TODO: return a map[string][][]string? map[string][]IrcMessage?
	t := strconv.Itoa64(time.Nanoseconds())
	ret := make(map[string][]string)
	ticker := time.NewTicker(timeout(n.lag))
	myreplies := []string{"ERR_NOSUCHSERVER", "ERR_NONICKNAMEGIVEN",
		"RPL_WHOISUSER", "RPL_WHOISCHANNELS",
		"RPL_WHOISSERVER", "RPL_AWAY",
		"RPL_WHOISOPERATOR", "RPL_WHOISIDLE",
		"ERR_NOSUCHNICK", "RPL_ENDOFWHOIS"}
	repch := make(chan *IrcMessage, 10)
	defer func(myreplies []string, t string) {
		for _, rep := range myreplies {
			n.Listen.DelListener(replies[rep], t)
		}
		return
	}(myreplies, t)
	for _, rep := range myreplies {
		if err := n.Listen.RegListener(replies[rep], t, repch); err != nil {
			ticker.Stop()
			return ret, os.NewError(fmt.Sprintf("Couldn't whois %s=%s: %s", replies[rep], rep, err.String()))
		}
	}

	if server == "" {
		n.queueOut <- &IrcMessage{"", "WHOIS", []string{strings.Join(target, ",")}}
	} else {
		n.queueOut <- &IrcMessage{"", "WHOIS", []string{server, strings.Join(target, ",")}}
	}
	for _, rep := range myreplies {
		ret[replies[rep]] = make([]string, 0)
	}
	done := 0
	err := os.Error(nil)
	for {
		select {
		case m := <-repch:
			ret[m.Cmd] = append(ret[m.Cmd], strings.Join((*m).Params, " "))
			if m.Cmd == replies["RPL_ENDOFWHOIS"] {
				ticker.Stop()
				return ret, err
			} else if m.Cmd == replies["ERR_NOSUCHNICK"] {
				for _, targ := range target {
					if m.Params[1] == targ {
						if err == nil {
							err = os.NewError(fmt.Sprintf("No such nick: %s", targ))
						} else {
							err = os.NewError(fmt.Sprintf("%s, No such nick: %s", err.String(), targ))
						}
						done++
					}
				}
				if done == len(target) {
					return ret, err
				}
			}
			ticker.Stop()
			ticker = time.NewTicker(timeout(n.lag)) //restart the ticker to timeout correctly
		case <-ticker.C:
			ticker.Stop()
			return ret, err
		}
	}
	ticker.Stop()
	return ret, err //BUG: why do we need this?
}

func (n *Network) Whowas(target string, count int, server string) {
	msg := &IrcMessage{"", "WHOIS", []string{}}
	msg.Params = append(msg.Params, target)
	if count != 0 {
		msg.Params = append(msg.Params, strconv.Itoa(count))
	}
	if server != "" {
		msg.Params = append(msg.Params, server)
	}
	n.queueOut <- msg
	//TODO: replies:
	//ERR_NONICKNAMEGIVEN             ERR_WASNOSUCHNICK
	//RPL_WHOWASUSER                  RPL_WHOISSERVER
	//RPL_ENDOFWHOWAS
	return
}

func (n *Network) PingNick(nick string) {
	n.queueOut <- &IrcMessage{"", "PING", []string{nick}}
	//TODO: replies:
	//ERR_NOORIGIN                    ERR_NOSUCHSERVER
	return
}

func (n *Network) Ping() (int64, os.Error) {
	myreplies := []string{"ERR_NOORIGIN", "ERR_NOSUCHSERVER"}
	t := strconv.Itoa64(time.Nanoseconds())
	repch := make(chan *IrcMessage, 10)
	ticker := time.NewTicker(timeout(n.lag))
	defer ticker.Stop()
	defer func(myreplies []string, t string, n *Network) {
		for _, rep := range myreplies {
			n.Listen.DelListener(replies[rep], t)
		}
		n.Listen.DelListener("PONG", t)
		return
	}(myreplies, t, n)
	for _, rep := range myreplies {
		n.Listen.RegListener(replies[rep], t, repch)
	}
	n.Listen.RegListener("PONG", t, repch)
	var rep *IrcMessage
	n.queueOut <- &IrcMessage{"", "PING", []string{strconv.Itoa64(time.Nanoseconds())}}
	select {
	case <-ticker.C:
		return 0, os.NewError("Timeout in receiving reply")
	case rep = <-repch:
	}
	if rep.Cmd == "PONG" {
		origtime, err := strconv.Atoi64(rep.Params[len(rep.Params)-1])
		if err == nil {
			n.lag = time.Nanoseconds() - origtime
			return n.lag, err
		} else {
			return 0, err
		}
	} else {
		switch rep.Cmd {
		case replies["ERR_NOORIGIN"]:
			return 0, os.NewError("ERR_NOORIGIN")
		case replies["ERR_NOSUCHSERVER"]:
			return 0, os.NewError("ERR_NOSUCHSERVER")
		default:
			return 0, os.NewError("Unknown error")
		}
	}
	return 0, os.NewError("Unknown error")
}

func (n *Network) Pong(msg string) {
	n.queueOut <- &IrcMessage{"", "PONG", []string{msg}}
	//TODO: numeric replies? PingNick?
	return
}

func (n *Network) Away(reason string) {
	msg := &IrcMessage{"", "AWAY", []string{}}
	if reason != "" {
		msg.Params = append(msg.Params, reason)
	}
	n.queueOut <- msg
	//TODO: replies:
	//RPL_UNAWAY                      RPL_NOWAWAY
	return
}

func (n *Network) Users(server string) {
	msg := &IrcMessage{"", "USERS", []string{}}
	if server != "" {
		msg.Params = append(msg.Params, server)
	}
	n.queueOut <- msg
	return
}

func (n *Network) Userhost(users []string) {
	if len(users) > 5 {
		//todo cycle them 5-by-5?
		return
	}
	n.queueOut <- &IrcMessage{"", "USERHOST", []string{strings.Join(users, " ")}}
	//TODO: replies
	//RPL_USERHOST                    ERR_NEEDMOREPARAMS
	return
}

func (n *Network) Ison(users []string) {
	if len(users) > 53 { //maximum number of nicks: 512/9 9 is max length of a nick
		return
	}
	n.queueOut <- &IrcMessage{"", "ISON", []string{strings.Join(users, " ")}}
	//TODO: replies
	//RPL_ISON                ERR_NEEDMOREPARAMS
	return
}

func (n *Network) SendRaw(raw string) {
	msg, err := PackMsg(raw)
	if err == nil {
		n.queueOut <- &msg
	}
}

func (n *Network) SetPort(port string) {
	n.port = port
	n.Reconnect("Changing server.")
}

func (n *Network) SetNetwork(net string) {
	n.network = net
	n.Reconnect("Changing server.")
}

func (n *Network) SetVersion(newversion string) {
	IRCVERSION = newversion
}

func (n *Network) GetVersion() string {
	return IRCVERSION
}
