package commands

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type move struct {
	user   discordgo.User
	chanID string
	x, y   int
}

var games = make(map[string]*ticTacToe)

// discordgo message event handler to parse commands
func Parse(session *discordgo.Session, message *discordgo.MessageCreate) {
	if me, _ := session.User("@me"); me.ID == message.Author.ID {
		return
	}

	if strings.Index(message.Content, "+") != 0 {
		return
	}
	args := strings.Fields(message.Content)
	command, args := args[0][1:], args[1:]
	log.Printf("%s: %v", command, args)

	switch command {
	case "say":
		log.Printf("COMMAND: +say")
		del := " "
		if args[0] == "-d" && len(args) >= 2 {
			del, args = args[1], args[2:]
		}
		session.ChannelMessageSend(message.ChannelID, strings.Join(args, del))
	case "ping":
		log.Printf("COMMAND: +ping")
		session.ChannelMessageSend(message.ChannelID, "Pong!")
	case "tictactoe":
		log.Printf("COMMAND: +tictactoe")
		if _, ok := games[message.ChannelID]; ok {
			log.Printf("TICTACTOE: Invalid: ongoing game")
			session.ChannelMessageSend(message.ChannelID, "There's already a game "+
				"of tic tac toe being played in this channel.")
			return
		} else if len(args) == 0 {
			log.Printf("TICTACTOE: Invalid: No opponent specified")
			session.ChannelMessageSend(message.ChannelID, "You must specify an opponent.")
			return
		}

		c, _ := session.Channel(message.ChannelID)
		u, err := session.User(strings.Trim(args[0], "@<>"))
		if err != nil {
			log.Printf("TICTACTOE: Invalid: UserID not found")
			session.ChannelMessageSend(message.ChannelID, "Invalid opponent.")
			return
		}

		p, _ := session.State.Presence(c.GuildID, u.ID)
		if p.Status != discordgo.StatusOnline {
			log.Printf("TICTACTOE: Invalid: Opponent offline")
			session.ChannelMessageSend(message.ChannelID, "Opponent is offline.")
			return
		} else if u.ID == message.Author.ID {
			log.Printf("TICTACTOE: Invalid: Opponent is self")
			session.ChannelMessageSend(message.ChannelID, "Cannot play with yourself.")
			return
		} else if u.Bot {
			log.Printf("TICTACTOE: Invalid: Opponent is bot")
			session.ChannelMessageSend(message.ChannelID, "Cannot play with a bot.")
			return
		}

		games[message.ChannelID] = newTicTacToe(*u, *message.Author, message.ChannelID)
		session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Starting game of "+
			"tic tac toe: %s vs %s!\n"+
			"Type '+move [x] [y]' to move, type '+tictactoe cancel' to quit.\n"+
			"%s goes first!",
			message.Author.Mention(), u.Mention(), games[message.ChannelID].player.Mention()))
		session.ChannelMessageSend(message.ChannelID, games[message.ChannelID].ToString())
	case "move":
		log.Printf("COMMAND: +move")
		if len(args) != 2 {
			log.Printf("MOVE: Invalid move")
			session.ChannelMessageSend(message.ChannelID, "Invalid move.")
			return
		}
		x, err := strconv.Atoi(args[0])
		if err != nil || x > 2 || x < 0 {
			log.Printf("MOVE: Invalid move")
			session.ChannelMessageSend(message.ChannelID, "Invalid move.")
			return
		}
		y, err := strconv.Atoi(args[1])
		if err != nil || y > 2 || y < 0 {
			log.Printf("MOVE: Invalid move")
			session.ChannelMessageSend(message.ChannelID, "Invalid move.")
			return
		}
		t := games[message.ChannelID]
		if t == nil {
			log.Printf("MOVE: Invalid: No game in progress")
			session.ChannelMessageSend(message.ChannelID, "No game in progress.")
			return
		}
		if t.player.ID != message.Author.ID {
			log.Printf("MOVE: Invalid: Played out of turn")
			session.ChannelMessageSend(message.ChannelID, "It's not your turn.")
			return
		}
		if t.board[x][y] != 0 {
			log.Printf("MOVE: Invalid move")
			session.ChannelMessageSend(message.ChannelID, "Invalid move.")
			return
		}
		m := move{*message.Author, message.ChannelID, x, y}
		handleMove(m, session)
	case "cancel":
		if _, ok := games[message.ChannelID]; ok {
			delete(games, message.ChannelID)
			log.Printf("TICTACTOE: Game canclled")
			session.ChannelMessageSend(message.ChannelID, "Game cancelled.")
		} else {
			log.Printf("TICTACTOE: No game to cancel")
			session.ChannelMessageSend(message.ChannelID, "No ongoing game.")
		}
	}
}

func handleMove(m move, session *discordgo.Session) {
	t := games[m.chanID]
	t.board[m.x][m.y] = t.p
	t.player, t.other = t.other, t.player
	t.p, t.o = t.o, t.p
	session.ChannelMessageSend(m.chanID, games[m.chanID].ToString())
	if t.checkVictory(m.x, m.y) {
		log.Printf("TICTACTOE: %s wins", m.user.ID)
		delete(games, m.chanID)
		session.ChannelMessageSend(m.chanID, fmt.Sprintf("Congratulations, %s wins!",
			m.user.Mention()))
	} else if t.checkFull() {

		delete(games, m.chanID)
		session.ChannelMessageSend(m.chanID, "Draw!")
	}
}

func newTicTacToe(u1, u2 discordgo.User, chanID string) *ticTacToe {
	rand.Seed(time.Now().UnixNano())
	var player, other discordgo.User
	if rand.Float32() < 0.5 {
		player, other = u1, u2
	} else {
		player, other = u2, u1
	}
	return &ticTacToe{player, other, 1, 2, *new([3][3]int), chanID}
}

type ticTacToe struct {
	player, other discordgo.User
	p, o          int
	board         [3][3]int
	chanID        string
}

func (t ticTacToe) ToString() string {
	buf := new(bytes.Buffer)
	for i := 0; i < 3; i++ {
		buf.WriteString("|")
		for j := 0; j < 2; j++ {
			switch games[t.chanID].board[j][2-i] {
			case 0:
				buf.WriteString("    ")
			case 1:
				buf.WriteString("O ")
			case 2:
				buf.WriteString("X  ")
			}
		}
		switch games[t.chanID].board[2][2-i] {
		case 0:
			buf.WriteString("   |")
		case 1:
			buf.WriteString("O|")
		case 2:
			buf.WriteString("X|")
		}
		buf.WriteString(" " + strconv.Itoa(2-i) + "\n")
	}
	buf.WriteString(" 0   1   2")
	return buf.String()
}

func (t ticTacToe) checkVictory(x, y int) bool {
	if t.board[0][y] == t.board[1][y] && t.board[1][y] == t.board[2][y] {
		return true
	}

	if t.board[x][0] == t.board[x][1] && t.board[x][1] == t.board[x][2] {
		return true
	}

	if x == y && t.board[0][0] == t.board[1][1] && t.board[1][1] == t.board[2][2] {
		return true
	}

	if x+y == 2 && t.board[0][2] == t.board[1][1] && t.board[1][1] == t.board[2][0] {
		return true
	}

	return false
}

func (t ticTacToe) checkFull() bool {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if t.board[i][j] == 0 {
				return false
			}
		}
	}
	return true
}
