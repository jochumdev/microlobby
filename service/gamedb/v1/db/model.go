package db

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"

	sdb "wz2100.net/microlobby/shared/db"
)

type GamePlayer struct {
	bun.BaseModel `bun:"game_players,alias:p"`
	Id            int       `bun:"id,pk,autoincrement,type:bigserial" json:"id" yaml:"id"`
	GameID        uuid.UUID `bun:"game_id,type:uuid" json:"game_id" yaml:"game_id"`
	Game          *Game     `bun:"rel:belongs-to" json:"game" yaml:"game"`
	UUID          uuid.UUID `bun:"uuid,type:uuid" json:"uuid" yaml:"uuid"`
	Name          string    `bun:"name" json:"name" yaml:"name"`
	IpAddress     string    `bun:"ip_address" json:"ip_address" yaml:"ip_address"`
	IsHost        bool      `bun:"is_host" json:"is_host" yaml:"is_host"`

	sdb.Timestamps
	sdb.SoftDelete
}

type Game struct {
	bun.BaseModel `bun:"games,alias:g"`
	Id            uuid.UUID     `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" yaml:"id"`
	Description   string        `bun:"description" json:"description" yaml:"description"`
	Map           string        `bun:"map" json:"map" yaml:"map"`
	HostIp        string        `bun:"host_ip" json:"host_ip" yaml:"host_ip"`
	Port          uint32        `bun:"port" json:"port" yaml:"port"`
	Players       []*GamePlayer `bun:"rel:has-many,join:id=game_id" json:"players" yaml:"players"`
	MaxPlayers    uint32        `bun:"max_players" json:"max_players" yaml:"max_players"`
	Version       string        `bun:"version" json:"version" yaml:"version"`
	VerMajor      uint32        `bun:"ver_major" json:"ver_major" yaml:"ver_major"`
	VerMinor      uint32        `bun:"ver_minor" json:"ver_minor" yaml:"ver_minor"`
	IsPure        bool          `bun:"is_pure" json:"is_pure" yaml:"is_pure"`
	IsPrivate     bool          `bun:"is_private" json:"is_private" yaml:"is_private"`
	LobbyVersion  uint32        `bun:"lobby_version" json:"lobby_version" yaml:"lobby_version"`
	V3GameId      uint32        `bun:"v3_game_id" json:"v3_game_id" yaml:"v3_game_id"`
	Mods          []string      `bun:"mods,array" json:"mods" yaml:"mods"`

	sdb.Timestamps
	sdb.SoftDelete
}
