package main

import (
	"bufio"
	"net"
	"strings"
	"sync"
	"time"

	"golang.org/x/text/encoding/charmap"
)

const (
	LEVEL_ANY    = -1000
	LEVEL_NEWBIE = 0
	LEVEL_PLAYER = 1

	LEVEL_BUILDER     = LEVEL_OWNER - 40
	LEVEL_MODERATOR   = LEVEL_OWNER - 30
	LEVEL_ADMIN       = LEVEL_OWNER - 20
	LEVEL_IMPLEMENTER = LEVEL_OWNER - 10
	LEVEL_OWNER       = 1000
)

type levelNameData struct {
	Name        string
	Short       string
	Description string
}

var levelToName map[int]*levelNameData = map[int]*levelNameData{
	LEVEL_ANY:         {Name: "Anyone", Short: "any", Description: "Anyone"},
	LEVEL_NEWBIE:      {Name: "Newbie", Short: "newb", Description: "New players"},
	LEVEL_PLAYER:      {Name: "Players", Short: "play", Description: "Normal players"},
	LEVEL_BUILDER:     {Name: "Builders", Short: "build", Description: "People working on areas, objects, mobs, etc."},
	LEVEL_MODERATOR:   {Name: "Moderator", Short: "mod", Description: "People who moderate the MUD"},
	LEVEL_ADMIN:       {Name: "Administrators", Short: "admin", Description: "Mud staff, administrators"},
	LEVEL_IMPLEMENTER: {Name: "Implementer", Short: "imp", Description: "People writing code for the MUD."},
	LEVEL_OWNER:       {Name: "Owner", Short: "own", Description: "Owner of the MUD."},
}

func nameToLevel(input string) (int, bool) {
	input = strings.ToLower(input)
	for lvl, item := range levelToName {
		if strings.HasPrefix(input, item.Short) {
			return lvl, true
		}
	}
	return LEVEL_ANY, false
}

var (
	topID    uint64
	descList []*descData
	descLock sync.Mutex
	charList []*characterData
)

type descData struct {
	id        uint64
	conn      net.Conn
	reader    *bufio.Reader
	state     int
	dns, ip   string
	idleTime  time.Time
	processed bool

	tls bool

	telnet telnetData

	inputLock sync.Mutex
	inBufLen  int
	inBuf     []byte
	outBuf    []byte
	haveOut   bool

	numInputLines int
	inputLines    []string

	account   *accountData
	character *characterData

	connectTime time.Time
	valid       bool
}

type IgnoreData struct {
	Name   string
	UUID   uuidData
	Silent bool
	Added  time.Time
}

type ConfigValue struct {
	Format int
	Name   string
	ValInt int
	ValStr string
}

type draftNoteData struct {
	Editing bool

	DraftNotes []*noteData
}

type characterData struct {
	Version int
	UUID    uuidData
	desc    *descData

	Name  string
	room  *roomData
	Loc   LocData `json:",omitempty"`
	Level int

	Vitals VitalsData

	Prompt string `json:",omitempty"`

	Channels   Bitmask              `json:",omitempty"`
	Config     Bitmask              `json:",omitempty"`
	ConfigVals map[int]*ConfigValue `json:",omitempty"`
	LastHide   time.Time            `json:",omitempty"`

	OLCEditor  olcEditorData `json:",omitempty"`
	Ignores    []IgnoreData  `json:",omitempty"`
	NoteRead   map[string]time.Time
	DraftNotes draftNoteData `json:",omitempty"`

	NumReports int `json:",omitempty"`

	SaveTime time.Time
	CreDate  time.Time
	idleTime time.Time

	Tells  []tellData `json:",omitempty"`
	Banned []banData  `json:",omitempty"`

	loginTime time.Time

	dirty bool
	valid bool
}

type VitalsData struct {
	Heal, Move, Mana         int
	Dead, Immobile, Silenced bool
}

type UndoData struct {
	OLCMode int
	Name    string
	From    string
	To      string
	Loc     LocData
}

type olcEditorData struct {
	OLCMode int `json:",omitempty"`

	Location LocData
	area     *areaData
	room     *roomData

	Undo []UndoData `json:",omitempty"`

	EditText []string `json:",omitempty"`
}

type tellData struct {
	SenderName string
	SenderUUID uuidData
	Message    string
	Sent       time.Time
}

type telnetData struct {
	Charset, termType string

	charMap *charmap.Charmap
	Options *termSettings

	subSeqType byte
	subSeqMode bool
	subSeqData []byte
	subLength  int

	hideEcho bool
}

type accountData struct {
	Version int
	UUID    uuidData
	Level   int

	Login    string
	PassHash []byte

	tempString string

	CreDate time.Time
	ModDate time.Time

	Characters []accountIndexData
	Banned     []banData `json:",omitempty"`

	dirty bool
}

type banData struct {
	Reason string `json:",omitempty"`
	Date   time.Time
	BanBy  string

	Temporary bool
	Until     time.Time

	Revoked bool
}
