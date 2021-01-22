package distance

import (
	"math"
	"net/url"
	"time"
)

// Summary is the structure of a summary response from /summary.
type Summary struct {
	Server       Server
	Level        Level
	ChatLog      []ChatMessage
	Players      []Player
	AutoServer   AutoServer
	VoteCommands VoteCommands
}

// Server describes the server in Summary.
type Server struct {
	CurrentLevelID               int64 `json:"CurrentLevelId"`
	MaxPlayers                   int64
	Port                         int64
	ReportToMasterServer         bool
	MasterServerGameModeOverride string
	DistanceVersion              int64
	IsInLobby                    bool
	HasModeStarted               bool
	ModeStartTime                float64
}

// AutoServer is part of Summary.
type AutoServer struct {
	IdleTimeout                      int64
	LevelTimeout                     int64
	AdvanceWhenStartingPlayersFinish bool
	WelcomeMessage                   string
	LevelEndTime                     float64
	StartingPlayerGuids              []string
}

// ChatMessage is part of Summary. It describes a message in the ChatLog.
type ChatMessage struct {
	Sender      string
	GUID        string `json:"Guid"`
	Timestamp   float64
	Chat        string
	Type        ChatMessageType
	Description string
}

// Time returns the timestamp in time.Time.
func (msg ChatMessage) Time() time.Time {
	sec, f := math.Modf(msg.Timestamp)
	return time.Unix(int64(sec), int64(float64(time.Second)*f))
}

// ChatMessageType is the enumerated type for a message's type.
type ChatMessageType string

const (
	PlayerCustomMessage  ChatMessageType = "PlayerCustom"
	ServerCustomMessage  ChatMessageType = "ServerCustom"
	ServerVanillaMessage ChatMessageType = "ServerVanilla"
	PlayerActionMessage  ChatMessageType = "PlayerAction"
	PlayerChatMessage    ChatMessageType = "PlayerChatMessage"
)

// Level describes a level in Summary.
type Level struct {
	Index             int
	Name              string
	RelativeLevelPath string
	WorkshopFileID    string `json:"WorkshopFileId"`
	GameMode          string
	Difficulty        string
}

// WorkshopURL returns the workshop URL to this level.
func (lvl Level) WorkshopURL() string {
	return "https://steamcommunity.com/sharedfiles/filedetails/?id=" + lvl.WorkshopFileID
}

// Player describes a player in Summary.
type Player struct {
	UnityPlayerGUID        string `json:"UnityPlayerGuid"`
	State                  PlayerState
	Stuck                  bool
	LevelID                int `json:"LevelId"`
	ReceivedInfo           bool
	Index                  int
	Name                   string
	JoinedAt               float64
	ValidatedAt            float64
	Ready                  bool
	Car                    Car
	LevelCompatibilityInfo LevelCompatibilityInfo
	LevelCompatibility     string
	Valid                  bool
	IPAddress              string `json:"IpAddress"`
	Port                   int
}

// PlayerState describes the current state of a Player in Summary.
type PlayerState string

const (
	PlayerInitializing                  PlayerState = "Initializing"
	PlayerInitialized                   PlayerState = "Initialized"
	PlayerLoadingLobbyScene             PlayerState = "LoadingLobbyScene"
	PlayerLoadedLobbyScene              PlayerState = "LoadedLobbyScene"
	PlayerSubmittedLobbyInfo            PlayerState = "SubmittedLobbyInfo"
	PlayerWaitingForCompatibilityStatus PlayerState = "WaitingForCompatibilityStatus"
	PlayerLoadingGameModeScene          PlayerState = "LoadingGameModeScene"
	PlayerLoadedGameModeScene           PlayerState = "LoadedGameModeScene"
	PlayerSubmittedGameModeInfo         PlayerState = "SubmittedGameModeInfo"
	PlayerStartedMode                   PlayerState = "StartedMode"
	PlayerCantLoadLevelSoInLobby        PlayerState = "CantLoadLevelSoInLobby"
)

// LevelCompatibilityInfo describes the level compatibility information from
// Summary.
type LevelCompatibilityInfo struct {
	LevelCompatibilityID int `json:"LevelCompatibilityId"`
	HasLevel             bool
	LevelVersion         string
}

// Car describes a Player's car in Summary.
type Car struct {
	Colors          [][4]float32 `json:"CarColors"`
	Name            string       `json:"CarName"`
	Points          int
	Finished        bool
	FinishData      int
	Spectator       bool
	Alive           bool
	WingsOpen       bool
	Position        []float32
	Rotation        []float32
	Velocity        []float32
	AngularVelocity []float32

	// TODO: FinishType
}

// VoteCommands describes the vote commands in Summary.
type VoteCommands struct {
	SkipThreshold   float64
	HasSkipped      bool
	ExtendThreshold float64
	ExtendTime      float64
	LeftAt          map[string]float64
	PlayerVotes     map[string]Level
	AgainstVotes    map[string]int
	SkipVotes       []string
	ExtendVotes     []string
}

// Summary gets the server summary.
func (c *Client) Summary() (*Summary, error) {
	var s *Summary
	return s, c.getJSON(url.URL{Path: "/summary"}, &s)
}
