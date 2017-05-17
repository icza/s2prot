/*

Type describing the game metadata (calculated, confirmed results).

*/

package rep

import "github.com/icza/s2prot"

// Metadata describes the game metadata (calculated, confirmed results).
type Metadata struct {
	s2prot.Struct

	players []MetaPlayer // Lazily initialized meta players
}

// Title returns the map name.
func (m *Metadata) Title() string {
	return m.Stringv("Title")
}

// GameVersion returns the game version string.
func (m *Metadata) GameVersion() string {
	return m.Stringv("GameVersion")
}

// DataBuild returns the data build version string.
func (m *Metadata) DataBuild() string {
	return m.Stringv("DataBuild")
}

// BaseBuild returns the base build version string.
// This has a "Base" prefix to the base build number.
func (m *Metadata) BaseBuild() string {
	return m.Stringv("BaseBuild")
}

// DurationSec returns the game duraiton in seconds.
func (m *Metadata) DurationSec() float64 {
	return m.Float("Duration")
}

// Players returns the list of meta players.
func (m *Metadata) Players() []MetaPlayer {
	if m.players == nil {
		players := m.Array("Players")
		m.players = make([]MetaPlayer, len(players))
		for i, pl := range players {
			// Metadata is a result of JSON unmarshaling (and not protocol decoding)
			// So Players will not be of type s2prot.Struct but a simple map:
			p := MetaPlayer{Struct: s2prot.Struct(pl.(map[string]interface{}))}
			m.players[i] = p
		}
	}

	return m.players
}

// MetaPlayer describes a player in the metadata section.
type MetaPlayer struct {
	s2prot.Struct
}

// PlayerID returns the player ID.
func (m *MetaPlayer) PlayerID() int64 {
	return m.Int("PlayerID")
}

// MMR returns the player's (race-specific) MMR value.
func (m *MetaPlayer) MMR() float64 {
	return m.Float("MMR")
}

// APM returns the player's APM value.
func (m *MetaPlayer) APM() float64 {
	return m.Float("APM")
}

// Result returns the result string.
// "Win" for win, "Loss" for loss, "Undecided" if unknown.
func (m *MetaPlayer) Result() string {
	return m.Stringv("Result")
}

// SelectedRace returns the player's selected race string.
// It's a 4-letter prefix of the race, e.g. "Rand", "Prot", "Terr", "Zerg".
func (m *MetaPlayer) SelectedRace() string {
	return m.Stringv("SelectedRace")
}

// AssignedRace returns the race string that was assigned to the player.
// It's a 4-letter prefix of the race (no Random), e.g. "Prot", "Terr", "Zerg".
func (m *MetaPlayer) AssignedRace() string {
	return m.Stringv("AssignedRace")
}
