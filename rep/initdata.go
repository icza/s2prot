/*

Types describing the init data (the inital lobby).

*/

package rep

import "github.com/icza/s2prot"

// InitData describes the init data (the initial lobby).
type InitData struct {
	s2prot.Struct

	GameDescription GameDescription `json:"-"` // Game description
	LobbyState      LobbyState      `json:"-"` // Lobby state
	UserInitDatas   []UserInitData  `json:"-"` // Array User init data structs
}

// newInitData creates a new init data from the specified Struct.
func newInitData(s s2prot.Struct) InitData {
	// Init data is a struct with 1 field only which is a struct. Use that as the root struct.
	i := InitData{Struct: s.Structv("syncLobbyState")}

	i.GameDescription = GameDescription{Struct: i.Structv("gameDescription")}
	i.GameDescription.GameOptions = GameOptions{Struct: i.GameDescription.Structv("gameOptions")}

	i.LobbyState = LobbyState{Struct: i.Structv("lobbyState")}
	slots := i.LobbyState.Array("slots")
	i.LobbyState.Slots = make([]Slot, len(slots))
	for j, s := range slots {
		i.LobbyState.Slots[j] = Slot{Struct: s.(s2prot.Struct)}
	}

	uids := i.Array("userInitialData")
	i.UserInitDatas = make([]UserInitData, len(uids))
	for j, uid := range uids {
		u := UserInitData{Struct: uid.(s2prot.Struct)}
		if cl := u.Stringv("clanLogo"); cl != "" {
			u.ClanLogo = newCacheHandle(cl)
		}
		i.UserInitDatas[j] = u
	}

	return i
}

// GameDescription is the game description
type GameDescription struct {
	s2prot.Struct

	GameOptions      GameOptions
	cacheHandles     []*CacheHandle    // Lazily initialized cache handles
	slotDescriptions []SlotDescription // Lazily initialized slot descriptions
}

// Region returns the region of the replay.
func (g *GameDescription) Region() *Region {
	if chs := g.CacheHandles(); len(chs) > 0 {
		return chs[0].Region
	}
	return RegionUnknown
}

// GameSpeed returns the game speed.
func (g *GameDescription) GameSpeed() *GameSpeed {
	return GameSpeeds[g.Int("gameSpeed")]
}

// HasExtensionMod returns if the game has extension mod.
func (g *GameDescription) HasExtensionMod() bool {
	return g.Bool("hasExtensionMod")
}

// HasNonBlizzardExtensionMod returns if the game has non-Blizzard extension mod.
func (g *GameDescription) HasNonBlizzardExtensionMod() bool {
	return g.Bool("hasNonBlizzardExtensionMod")
}

// IsBlizzardMap tells if the map is an official Blizzard map.
func (g *GameDescription) IsBlizzardMap() bool {
	return g.Bool("isBlizzardMap")
}

// GameType returns the game type.
func (g *GameDescription) GameType() int64 {
	return g.Int("gameType")
}

// IsCoopMode tells if the game is coop mode.
func (g *GameDescription) IsCoopMode() bool {
	return g.Bool("isCoopMode")
}

// IsPremadeFFA tells if the game is pre-made FFA.
func (g *GameDescription) IsPremadeFFA() bool {
	return g.Bool("isPremadeFFA")
}

// MapAuthorName returns the name of the map author.
func (g *GameDescription) MapAuthorName() string {
	return g.Stringv("mapAuthorName")
}

// MapFileName returns the name of the map file.
func (g *GameDescription) MapFileName() string {
	return g.Stringv("mapFileName")
}

// MapFileSyncChecksum returns the map file sync checksum.
func (g *GameDescription) MapFileSyncChecksum() int64 {
	return g.Int("mapFileSyncChecksum")
}

// MapSizeX returns the map width.
func (g *GameDescription) MapSizeX() int64 {
	return g.Int("mapSizeX")
}

// MapSizeY returns the map height.
func (g *GameDescription) MapSizeY() int64 {
	return g.Int("mapSizeY")
}

// MaxColors returns the max colors.
func (g *GameDescription) MaxColors() int64 {
	return g.Int("maxColors")
}

// MaxControls returns the max controls.
func (g *GameDescription) MaxControls() int64 {
	return g.Int("maxControls")
}

// MaxObservers returns the max observers.
func (g *GameDescription) MaxObservers() int64 {
	return g.Int("maxObservers")
}

// MaxPlayers returns the max players.
func (g *GameDescription) MaxPlayers() int64 {
	return g.Int("maxPlayers")
}

// MaxRaces returns the max races.
func (g *GameDescription) MaxRaces() int64 {
	return g.Int("maxRaces")
}

// MaxTeams returns the max teams.
func (g *GameDescription) MaxTeams() int64 {
	return g.Int("maxTeams")
}

// MaxUsers returns the max users.
func (g *GameDescription) MaxUsers() int64 {
	return g.Int("maxUsers")
}

// ModFileSyncChecksum returns the mod file sync checksum.
func (g *GameDescription) ModFileSyncChecksum() int64 {
	return g.Int("modFileSyncChecksum")
}

// RandomValue returns the random value.
func (g *GameDescription) RandomValue() int64 {
	return g.Int("randomValue")
}

// GameCacheName returns the game cache name.
func (g *GameDescription) GameCacheName() string {
	return g.Stringv("gameCacheName")
}

// DefaultAIBuild returns the default AI build.
func (g *GameDescription) DefaultAIBuild() int64 {
	return g.Int("defaultAIBuild")
}

// DefaultDifficulty returns the default difficulty.
func (g *GameDescription) DefaultDifficulty() int64 {
	return g.Int("defaultDifficulty")
}

// CacheHandles returns the array of cache handles.
func (g *GameDescription) CacheHandles() []*CacheHandle {
	if g.cacheHandles == nil {
		chs := g.Array("cacheHandles")
		g.cacheHandles = make([]*CacheHandle, len(chs))
		for i, ch := range chs {
			g.cacheHandles[i] = newCacheHandle(ch.(string))
		}
	}

	return g.cacheHandles
}

// GameOptions is the game options
type GameOptions struct {
	s2prot.Struct
}

// AdvancedSharedControl returns if advanced shared control.
func (g *GameOptions) AdvancedSharedControl() bool {
	return g.Bool("advancedSharedControl")
}

// Amm returns if AMM (AutoMM - Automated Match Making).
func (g *GameOptions) Amm() bool {
	return g.Bool("amm")
}

// BattleNet returns if game was played on Battle.net.
func (g *GameOptions) BattleNet() bool {
	return g.Bool("battleNet")
}

// ClientDebugFlags returns the client debug flags.
func (g *GameOptions) ClientDebugFlags() int64 {
	return g.Int("clientDebugFlags")
}

// CompetitiveOrRanked returns if game is competitive or if that property is not present if ranked.
// Competitive means either ranked or unranked. Before competitive there was no unraked type (so ranked="ladder").
func (g *GameOptions) CompetitiveOrRanked() bool {
	// competitive is present from base version 24674, replaces ranked
	if v, ok := g.Value("competitive").(bool); ok {
		return v
	}
	return g.Bool("ranked")
}

// Fog returns the fog.
func (g *GameOptions) Fog() int64 {
	return g.Int("fog")
}

// LockTeams returns if teams are locked.
func (g *GameOptions) LockTeams() bool {
	return g.Bool("lockTeams")
}

// NoVictoryOrDefeat returns if no victory or defeat.
func (g *GameOptions) NoVictoryOrDefeat() bool {
	return g.Bool("noVictoryOrDefeat")
}

// Observers returns the observers.
func (g *GameOptions) Observers() int64 {
	return g.Int("observers")
}

// RandomRaces returns if random races.
func (g *GameOptions) RandomRaces() bool {
	return g.Bool("randomRaces")
}

// TeamsTogether returns if teams together.
func (g *GameOptions) TeamsTogether() bool {
	return g.Bool("teamsTogether")
}

// UserDifficulty returns the user difficulty.
func (g *GameOptions) UserDifficulty() int64 {
	return g.Int("userDifficulty")
}

// Practice returns if practice.
func (g *GameOptions) Practice() bool {
	return g.Bool("practice")
}

// Cooperative returns if cooperative.
func (g *GameOptions) Cooperative() bool {
	return g.Bool("cooperative")
}

// HeroDuplicatesAllowed returns if hero duplicates are allowed.
func (g *GameOptions) HeroDuplicatesAllowed() bool {
	return g.Bool("heroDuplicatesAllowed")
}

// SlotDescriptions returns the array of slot descriptions.
func (g *GameDescription) SlotDescriptions() []SlotDescription {
	if g.slotDescriptions == nil {
		sds := g.Array("slotDescriptions")
		g.slotDescriptions = make([]SlotDescription, len(sds))
		for i, sd := range sds {
			g.slotDescriptions[i] = SlotDescription{Struct: sd.(s2prot.Struct)}
		}
	}

	return g.slotDescriptions
}

// SlotDescription is the slot description
type SlotDescription struct {
	s2prot.Struct
}

// AllowedAIBuilds returns the allowed AI builds bitmap.
func (s *SlotDescription) AllowedAIBuilds() s2prot.BitArr {
	return s.BitArr("allowedAIBuilds")
}

// AllowedColors returns the allowed colors bitmap.
func (s *SlotDescription) AllowedColors() s2prot.BitArr {
	return s.BitArr("allowedColors")
}

// AllowedControls returns the allowed controls bitmap.
func (s *SlotDescription) AllowedControls() s2prot.BitArr {
	return s.BitArr("allowedControls")
}

// AllowedDifficulty returns the allowed difficulty bitmap.
func (s *SlotDescription) AllowedDifficulty() s2prot.BitArr {
	return s.BitArr("allowedDifficulty")
}

// AllowedObserveTypes returns the allowed observe types bitmap.
func (s *SlotDescription) AllowedObserveTypes() s2prot.BitArr {
	return s.BitArr("allowedObserveTypes")
}

// AllowedRaces returns the allowed races bitmap.
func (s *SlotDescription) AllowedRaces() s2prot.BitArr {
	return s.BitArr("allowedRaces")
}

// LobbyState is the lobby state
type LobbyState struct {
	s2prot.Struct

	Slots []Slot
}

// DefaultAIBuild returns the default AI build.
func (l *LobbyState) DefaultAIBuild() int64 {
	return l.Int("defaultAIBuild")
}

// DefaultDifficulty returns the default difficulty.
func (l *LobbyState) DefaultDifficulty() int64 {
	return l.Int("defaultDifficulty")
}

// GameDuration returns the game duration.
func (l *LobbyState) GameDuration() int64 {
	return l.Int("gameDuration")
}

// HostUserID returns the host user ID.
func (l *LobbyState) HostUserID() int64 {
	return l.Int("hostUserId")
}

// IsSinglePlayer tells if game is single player.
func (l *LobbyState) IsSinglePlayer() bool {
	return l.Bool("isSinglePlayer")
}

// MaxObservers returns the max observers.
func (l *LobbyState) MaxObservers() int64 {
	return l.Int("maxObservers")
}

// MaxUsers returns the max users.
func (l *LobbyState) MaxUsers() int64 {
	return l.Int("maxUsers")
}

// Phase returns the phase.
func (l *LobbyState) Phase() int64 {
	return l.Int("phase")
}

// RandomSeed returns the random seed.
func (l *LobbyState) RandomSeed() int64 {
	return l.Int("randomSeed")
}

// PickedMapTag returns the picked map tag.
func (l *LobbyState) PickedMapTag() int64 {
	return l.Int("pickedMapTag")
}

// Slot describes a slot
type Slot struct {
	s2prot.Struct
}

// AIBuild returns the AI build.
func (s *Slot) AIBuild() int64 {
	return s.Int("aiBuild")
}

// ColorPrefColor returns the color preference color.
func (s *Slot) ColorPrefColor() *Color {
	return colorByID(s.Int("colorPref", "color"))
}

// Control returns the control.
func (s *Slot) Control() *Control {
	return controlByID(s.Int("control"))
}

// Difficulty returns the difficulty.
func (s *Slot) Difficulty() int64 {
	return s.Int("difficulty")
}

// Handicap returns the handicap.
func (s *Slot) Handicap() int64 {
	return s.Int("handicap")
}

// Licenses returns the array of licenses.
// The array has elements of type int64.
func (s *Slot) Licenses() []interface{} {
	return s.Array("licenses")
}

// LogoIndex returns the logo index.
func (s *Slot) LogoIndex() int64 {
	return s.Int("logoIndex")
}

// TandemID returns the tandem ID.
func (s *Slot) TandemID() int64 {
	return s.Int("tandemId")
}

// TandemLeaderUserID returns the tandem leader user ID (in case of Archon mode games).
func (s *Slot) TandemLeaderUserID() int64 {
	return s.Int("tandemLeaderUserId")
}

// Observe returns the observe.
func (s *Slot) Observe() *Observe {
	return observeByID(s.Int("observe"))
}

// RacePrefRace returns the race preference race. This may be RaceRandom.
func (s *Slot) RacePrefRace() *Race {
	if rp := s.Structv("racePref"); rp != nil {
		r := rp.Value("race")
		if r == nil {
			return RaceRandom
		}
		if i, ok := r.(int64); ok {
			return raceByID(i)
		}
	}

	return RaceUnknown
}

// Rewards returns the array of rewards.
// The array has elements of type int64.
func (s *Slot) Rewards() []interface{} {
	return s.Array("rewards")
}

// TeamID returns the team ID.
func (s *Slot) TeamID() int64 {
	return s.Int("teamId")
}

// ToonHandle returns the toon handle.
func (s *Slot) ToonHandle() string {
	return s.Stringv("toonHandle")
}

// UserID returns the user ID.
func (s *Slot) UserID() int64 {
	return s.Int("userId")
}

// WorkingSetSlotID returns the working set slot ID.
func (s *Slot) WorkingSetSlotID() int64 {
	return s.Int("workingSetSlotId")
}

// Hero returns the hero.
func (s *Slot) Hero() string {
	return s.Stringv("hero")
}

// Skin returns the skin.
func (s *Slot) Skin() string {
	return s.Stringv("skin")
}

// Mount returns the mount.
func (s *Slot) Mount() string {
	return s.Stringv("mount")
}

// Artifacts returns the array of artifacts.
// The array has elements of type string.
func (s *Slot) Artifacts() []interface{} {
	return s.Array("artifacts")
}

// Commander returns the commander.
func (s *Slot) Commander() string {
	return s.Stringv("commander")
}

// CommanderLevel returns the commander level.
func (s *Slot) CommanderLevel() int64 {
	return s.Int("commanderLevel")
}

// CommanderMasteryLevel returns the commander mastery level.
func (s *Slot) CommanderMasteryLevel() int64 {
	return s.Int("commanderMasteryLevel")
}

// CommanderMasteryTalents returns the array of commander mastery talents.
// The array has elements of type int64.
func (s *Slot) CommanderMasteryTalents() []interface{} {
	return s.Array("commanderMasteryTalents")
}

// HasSilencePenalty returns if there is slience penalty.
func (s *Slot) HasSilencePenalty() bool {
	return s.Bool("hasSilencePenalty")
}

// UserInitData describes user initial data
type UserInitData struct {
	s2prot.Struct

	ClanLogo *CacheHandle // Cache handle of the clan logo image resource
}

// ClanTag returns the clan tag.
func (u *UserInitData) ClanTag() string {
	return u.Stringv("clanTag")
}

// CombinedRaceLevels returns the combined race levels.
func (u *UserInitData) CombinedRaceLevels() int64 {
	return u.Int("combinedRaceLevels")
}

// CustomInterface tells if custom interface.
func (u *UserInitData) CustomInterface() bool {
	return u.Bool("customInterface")
}

// Examine tells if examine.
func (u *UserInitData) Examine() bool {
	return u.Bool("examine")
}

// HighestLeague returns the highest league.
func (u *UserInitData) HighestLeague() *League {
	// If property doesn't exist, zero value 0 is returned which is LeagueUnknown
	// which is exactly we would return anyway, so simply:
	return leagueByID(u.Int("highestLeague"))
}

// Name returns the name.
func (u *UserInitData) Name() string {
	return u.Stringv("name")
}

// Observe returns the observe.
func (u *UserInitData) Observe() *Observe {
	return observeByID(u.Int("observe"))
}

// RacePreferenceRace returns the race preference race.
func (u *UserInitData) RacePreferenceRace() int64 {
	return u.Int("racePreference", "race")
}

// RandomSeed returns the random seed.
func (u *UserInitData) RandomSeed() int64 {
	return u.Int("randomSeed")
}

// TeamPreferenceTeam returns the team preference team.
func (u *UserInitData) TeamPreferenceTeam() int64 {
	return u.Int("teamPreference", "team")
}

// TestAuto tells if test auto.
func (u *UserInitData) TestAuto() bool {
	return u.Bool("testAuto")
}

// TestMap tells if test map.
func (u *UserInitData) TestMap() bool {
	return u.Bool("testMap")
}

// TestType returns the test type.
func (u *UserInitData) TestType() int64 {
	return u.Int("testType")
}

// Hero returns the hero.
func (u *UserInitData) Hero() string {
	return u.Stringv("hero")
}

// Skin returns the skin.
func (u *UserInitData) Skin() string {
	return u.Stringv("skin")
}

// Mount returns the mount.
func (u *UserInitData) Mount() string {
	return u.Stringv("mount")
}

// ToonHandle returns the toon handle.
func (u *UserInitData) ToonHandle() string {
	return u.Stringv("toonHandle")
}
