/*

Types describing the game details (overall replay details).

*/

package rep

import (
	"fmt"
	"strings"
	"time"

	"github.com/icza/s2prot"
)

// Details describles the game details (overall replay details).
type Details struct {
	s2prot.Struct

	players      []Player       // Lazily initialized players
	cacheHandles []*CacheHandle // Lazily initialized cache handles
}

// Title returns the map name.
func (d *Details) Title() string {
	return d.Stringv("title")
}

// IsBlizzardMap tells if the map is an official Blizzard map.
func (d *Details) IsBlizzardMap() bool {
	return d.Bool("isBlizzardMap")
}

// GameSpeed returns the game speed.
func (d *Details) GameSpeed() *GameSpeed {
	return gameSpeedByID(d.Int("gameSpeed"))
}

// ThumbnailFile returns the map thumbnail file name.
func (d *Details) ThumbnailFile() string {
	return d.Stringv("thumbnail", "file")
}

// Time returns the replay date+time.
func (d *Details) Time() time.Time {
	// timeUTC is in 10 microsecond unit
	return time.Unix(0, (d.Int("timeUTC")-116444736000000000)*100)
}

// TimeLocalOffset returns the local time offset of the playing who saved the replay.
func (d *Details) TimeLocalOffset() time.Duration {
	// timeLocalOffset is in 10 microsecond unit
	return time.Duration(d.Int("timeLocalOffset") * 100)
}

// CacheHandles returns the array of cache handles.
func (d *Details) CacheHandles() []*CacheHandle {
	if d.cacheHandles == nil {
		chs := d.Array("cacheHandles")
		d.cacheHandles = make([]*CacheHandle, len(chs))
		for i, ch := range chs {
			d.cacheHandles[i] = newCacheHandle(ch.(string))
		}
	}

	return d.cacheHandles
}

// CampaignIndex returns the campaign index.
func (d *Details) CampaignIndex() int64 {
	return d.Int("campaignIndex")
}

// DefaultDifficulty returns the default difficulty.
func (d *Details) DefaultDifficulty() int64 {
	return d.Int("defaultDifficulty")
}

// Difficulty returns the difficulty.
func (d *Details) Difficulty() string {
	return d.Stringv("difficulty")
}

// Description returns the description.
func (d *Details) Description() string {
	return d.Stringv("description")
}

// ImageFilePath returns the image file path.
func (d *Details) ImageFilePath() string {
	return d.Stringv("imageFilePath")
}

// MapFileName returns the name of the map file.
func (d *Details) MapFileName() string {
	return d.Stringv("mapFileName")
}

// MiniSave returns if mini save.
func (d *Details) MiniSave() bool {
	return d.Bool("miniSave")
}

// ModPaths returns the mod paths.
func (d *Details) ModPaths() interface{} {
	return d.Value("modPaths")
}

// RestartAsTransitionMap returns if restart as transition map.
func (d *Details) RestartAsTransitionMap() bool {
	return d.Bool("restartAsTransitionMap")
}

// Players returns the list of players.
func (d *Details) Players() []Player {
	if d.players == nil {
		players := d.Array("playerList")
		d.players = make([]Player, len(players))
		for i, pl := range players {
			p := Player{Struct: pl.(s2prot.Struct)}
			p.Name = strings.Replace(p.Stringv("name"), "<sp/>", "", -1)
			p.Toon = Toon{Struct: p.Structv("toon")}
			c := p.Structv("color")
			p.Color = [4]byte{byte(c.Int("a")), byte(c.Int("r")), byte(c.Int("g")), byte(c.Int("b"))}
			d.players[i] = p
		}
	}

	return d.players
}

// Player (participant of the game). Includes computers players but excludes observers.
type Player struct {
	s2prot.Struct

	Name  string  // Name of the player. Contains optional clan tag.
	Toon  Toon    // Toon of the player. This is a unique identifier.
	Color [4]byte // Color of the player, ARGB components. A=255 means completely opaque, A=0 means completely transparent.
	race  *Race   // Lazily initialized race.
}

// RaceString returns the localized race string.
func (p *Player) RaceString() string {
	return p.Stringv("race")
}

// Race returns the race.
func (p *Player) Race() *Race {
	if p.race == nil {
		p.race = raceFromLocalString(p.Stringv("race"))
	}
	return p.race
}

// TeamID returns the team ID.
// Not always accurate! Team ID from slot (init data) should be used instead!
func (p *Player) TeamID() int64 {
	return p.Int("teamId")
}

// Result returns the game result.
func (p *Player) Result() *Result {
	return resultByID(p.Int("result"))
}

// Handicap returns the handicap.
func (p *Player) Handicap() int64 {
	return p.Int("handicap")
}

// WorkingSetSlotID returns the working set slot ID.
func (p *Player) WorkingSetSlotID() int64 {
	return p.Int("workingSetSlotId")
}

// Control returns the control.
func (p *Player) Control() *Control {
	return controlByID(p.Int("control"))
}

// Observe returns the observe.
// Not always accurate! Observe from slot (init data) should be used instead!
func (p *Player) Observe() *Observe {
	return observeByID(p.Int("observe"))
}

// Hero returns the hero.
func (p *Player) Hero() string {
	return p.Stringv("hero")
}

// Toon - a unique identifier (of a player)
type Toon struct {
	s2prot.Struct
}

// ID returns the ID.
func (t *Toon) ID() int64 {
	return t.Int("id")
}

// ProgramID returns the program ID, always "S2", leading zeros will be stripped.
func (t *Toon) ProgramID() string {
	// always "\0\0S2", strip leading zeros
	s := t.Stringv("programId")
	for i := 0; i < len(s); i++ {
		if s[i] != 0 {
			return s[i:]
		}
	}
	return s
}

// RealmID returns the realm ID.
func (t *Toon) RealmID() int64 {
	return t.Int("realm")
}

// Realm returns the realm.
func (t *Toon) Realm() *Realm {
	return t.Region().Realm(t.RealmID())
}

// RegionID returns the region ID.
func (t *Toon) RegionID() int64 {
	return t.Int("region")
}

// Region returns the region.
func (t *Toon) Region() *Region {
	return regionByID(t.RegionID())
}

// String returns a string representation of the Toon, the same format as used in
// InitData["lobbyState"]["slots"]["toonHandle"]:
//
//     regionId-programId-reamId-playerId
// Using value receiver as Player.Toon is not a pointer (and so printing Player.Toon will call this method).
func (t Toon) String() string {
	return fmt.Sprintf("%d-%s-%d-%d", t.RegionID(), t.ProgramID(), t.RealmID(), t.ID())
}
