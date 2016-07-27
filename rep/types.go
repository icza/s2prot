/*

Common types and constants used in decoded replay data.

*/

package rep

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"path"
	"strings"
)

// Base of enum-like types.
type Enum struct {
	Name string
}

// String returns the string representation of the enum (the name).
func (e *Enum) String() string {
	return e.Name
}

// Game speed type
type GameSpeed struct {
	Enum
	attrValue string // Game speed value used in attributes events
	RelSpeed  int    // Relative speed compared to Normal
}

// Slice of all game speeds, index is used in Details["gameSpeed"]
var GameSpeeds = []*GameSpeed{
	&GameSpeed{Enum{"Slower"}, "Slor", 60},
	&GameSpeed{Enum{"Slow"}, "Slow", 45},
	&GameSpeed{Enum{"Normal"}, "Norm", 36},
	&GameSpeed{Enum{"Fast"}, "Fast", 30},
	&GameSpeed{Enum{"Faster"}, "Fasr", 26},
	&GameSpeed{Enum{"Unknown"}, "", 26},
}

// Named game speeds.
var (
	GameSpeedSlower  = GameSpeeds[0]
	GameSpeedSlow    = GameSpeeds[1]
	GameSpeedNormal  = GameSpeeds[2]
	GameSpeedFast    = GameSpeeds[3]
	GameSpeedFaster  = GameSpeeds[4]
	GameSpeedUnknown = GameSpeeds[5]
)

// gameSpeedById returns the GameSpeed specified by its id.
// GameSpeedUnknown is returned if id is unknown.
func gameSpeedById(gameSpeedId int64) *GameSpeed {
	if id := int(gameSpeedId); id >= 0 && id < len(GameSpeeds) {
		return GameSpeeds[id]
	}
	return GameSpeedUnknown
}

// Race type.
type Race struct {
	Enum
	Letter rune // Race letter (first character of the English name)
}

// Slice of all races.
var Races = []*Race{
	&Race{Enum{"Terran"}, 'T'},
	&Race{Enum{"Zerg"}, 'Z'},
	&Race{Enum{"Protoss"}, 'P'},
	&Race{Enum{"Random"}, 'R'},
	&Race{Enum{"Unknown"}, '-'},
}

// Named races.
var (
	RaceTerran  = Races[0]
	RaceZerg    = Races[1]
	RaceProtoss = Races[2]
	RaceRandom  = Races[3]
	RaceUnknown = Races[4]
)

// Map of localized race names, maps from localized name to Race, used in Details["playerList"]["race"]
var localRaceNames = make(map[string]*Race)

func init() {
	// Build the localRaceNames map
	// English, German, Portuguese, Korean, Chinese, Russian, Polish, Mandarin (Chinese)
	for _, s := range []string{"Terran", "Terraner", "Terrano", "테란", "人類", "Терран", "Terrani", "人类"} {
		localRaceNames[s] = RaceTerran
	}
	// English, Korean, Chinese, Russian, Polish, Mandarin (Chinese)
	for _, s := range []string{"Zerg", "저그", "蟲族", "Зерг", "Zergi", "异虫"} {
		localRaceNames[s] = RaceZerg
	}
	// English, Korean, Chinese, Russian, Polish, Mandarin (Chinese)
	for _, s := range []string{"Protoss", "프로토스", "神族", "Протосс", "Protosi", "星灵"} {
		localRaceNames[s] = RaceProtoss
	}
}

// RaceFromLocalString returns the race specified by a localized name.
func raceFromLocalString(s string) *Race {
	if r, ok := localRaceNames[s]; ok {
		return r
	}

	// Could not find the localized value, let's try to find out
	switch {
	case strings.HasPrefix(s, "Pr"):
		return RaceProtoss
	case strings.HasPrefix(s, "Te"):
		return RaceTerran
	case strings.HasPrefix(s, "Ze"):
		return RaceZerg
	default:
		return RaceUnknown
	}
}

// raceById returns the Race specified by its id.
// RaceUnknown is returned if id is unknown.
func raceById(raceId int64) *Race {
	if id := int(raceId); id >= 0 && id < len(Races) {
		return Races[id]
	}
	return RaceUnknown
}

// Result type.
type Result struct {
	Enum
	Letter rune // Result letter (first character of the name)
}

// Slice of all results, index used in Details["playerList"]["result"]
var Results = []*Result{
	&Result{Enum{"Unknown"}, '-'},
	&Result{Enum{"Victory"}, 'V'},
	&Result{Enum{"Defeat"}, 'D'},
	&Result{Enum{"Tie"}, 'T'},
}

// Named results.
var (
	ResultUnknown = Results[0]
	ResultVictory = Results[1]
	ResultDefeat  = Results[2]
	ResultTie     = Results[3]
)

// resultById returns the Result specified by its id.
// ResultUnknown is returned if id is unknown.
func resultById(resultId int64) *Result {
	if id := int(resultId); id >= 0 && id < len(Results) {
		return Results[id]
	}
	return ResultUnknown
}

// Control type.
type Control struct {
	Enum
	attrValue string // Control value used in attributes events
}

// Slice of all control, index used in InitData["lobbyState"]["slots"]["control"] and in Details["playerList"]["control"]
var Controls = []*Control{
	&Control{Enum{"Open"}, "Open"},
	&Control{Enum{"Closed"}, "Clsd"},
	&Control{Enum{"Human"}, "Humn"},
	&Control{Enum{"Computer"}, "Comp"},
	&Control{Enum{"Unknown"}, ""},
}

// Named controls.
var (
	ControlOpen     = Controls[0]
	ControlClosed   = Controls[1]
	ControlHuman    = Controls[2]
	ControlComputer = Controls[3]
	ControlUnknown  = Controls[4]
)

// controlById returns the Control specified by its id.
// ControlUnknown is returned if id is unknown.
func controlById(controlId int64) *Control {
	if id := int(controlId); id >= 0 && id < len(Controls) {
		return Controls[id]
	}
	return ControlUnknown
}

// Observe type.
type Observe struct {
	Enum
}

// Slice of all observes, index used in InitData["lobbyState"]["slots"]["observe"] and in Details["playerList"]["observe"]
var Observes = []*Observe{
	&Observe{Enum{"Participant"}},
	&Observe{Enum{"Spectator"}},
	&Observe{Enum{"Referee"}},
	&Observe{Enum{"Unknown"}},
}

// Named observes.
var (
	ObserveParticipant = Observes[0]
	ObserveSpectator   = Observes[1] // Can only talk to other observers.
	ObserveReferee     = Observes[2] // Can talk to players as well.
	ObserveUnknown     = Observes[3]
)

// observeById returns the Observe specified by its id.
// ObserveUnknown is returned if id is unknown.
func observeById(observeId int64) *Observe {
	if id := int(observeId); id >= 0 && id < len(Observes) {
		return Observes[id]
	}
	return ObserveUnknown
}

// Color type.
type Color struct {
	Enum
	RGB       [3]byte // Color value, RGB components.
	Darker    [3]byte // Darker version of the color's RGB values.
	Lighter   [3]byte // Lighter versions of the color's RGB values.
	attrValue string  // Color value used in attributes events
}

// Slice of all colors, index used in InitData["lobbyState"]["slots"]["colorPref"]["color"]
var Colors = []*Color{
	&Color{Enum: Enum{"Unknown"}, RGB: [3]byte{0, 0, 0}},
	&Color{Enum: Enum{"Red"}, RGB: [3]byte{180, 20, 30}},
	&Color{Enum: Enum{"Blue"}, RGB: [3]byte{0, 66, 255}},
	&Color{Enum: Enum{"Teal"}, RGB: [3]byte{28, 167, 234}},
	&Color{Enum: Enum{"Purple"}, RGB: [3]byte{84, 0, 129}},
	&Color{Enum: Enum{"Yellow"}, RGB: [3]byte{235, 225, 41}},
	&Color{Enum: Enum{"Orange"}, RGB: [3]byte{254, 138, 14}},
	&Color{Enum: Enum{"Green"}, RGB: [3]byte{22, 128, 0}},
	&Color{Enum: Enum{"Light Pink"}, RGB: [3]byte{204, 166, 252}},
	&Color{Enum: Enum{"Violet"}, RGB: [3]byte{31, 1, 201}},
	&Color{Enum: Enum{"Light Gray"}, RGB: [3]byte{82, 84, 148}},
	&Color{Enum: Enum{"Dark Green"}, RGB: [3]byte{16, 98, 70}},
	&Color{Enum: Enum{"Brown"}, RGB: [3]byte{78, 42, 4}},
	&Color{Enum: Enum{"Light Green"}, RGB: [3]byte{150, 255, 145}},
	&Color{Enum: Enum{"Dark Gray"}, RGB: [3]byte{35, 35, 35}},
	&Color{Enum: Enum{"Pink"}, RGB: [3]byte{229, 91, 176}},
}

func init() {
	// Init calculated / derivative fields of Color.
	for i, c := range Colors {
		c.attrValue = fmt.Sprintf("tc%02d", i+1)
		c.Darker = [3]byte{c.RGB[0] / 2, c.RGB[1] / 2, c.RGB[2] / 2}
		c.Lighter = [3]byte{128 + c.Darker[0], 128 + c.Darker[1], 128 + c.Darker[2]}
	}
}

// Named colors.
var (
	ColorUnknown    = Colors[0]
	ColorRed        = Colors[1]
	ColorBlue       = Colors[2]
	ColorTeal       = Colors[3]
	ColorPurple     = Colors[4]
	ColorYellow     = Colors[5]
	ColorOrange     = Colors[6]
	ColorGreen      = Colors[7]
	ColorLightPink  = Colors[8]
	ColorViolet     = Colors[9]
	ColorLightGray  = Colors[10]
	ColorDarkGreen  = Colors[11]
	ColorBrown      = Colors[12]
	ColorLightGreen = Colors[13]
	ColorDarkGray   = Colors[14]
	ColorPink       = Colors[15]
)

// colorById returns the Color specified by its id.
// ColorUnknown is returned if id is unknown.
func colorById(colorId int64) *Color {
	if id := int(colorId); id >= 0 && id < len(Colors) {
		return Colors[id]
	}
	return ColorUnknown
}

// League type.
type League struct {
	Enum
	Letter rune // League letter (first character of the English name except 'R' for LeagueGrandmaster and '-' for Unknown)
}

// Slice of all leagues.
var Leagues = []*League{
	&League{Enum{"Unknown"}, '-'},
	&League{Enum{"Bronze"}, 'B'},
	&League{Enum{"Silver"}, 'S'},
	&League{Enum{"Gold"}, 'G'},
	&League{Enum{"Platinum"}, 'P'},
	&League{Enum{"Diamond"}, 'D'},
	&League{Enum{"Master"}, 'M'},
	&League{Enum{"Grandmaster"}, 'R'},
	&League{Enum{"Unranked"}, 'U'},
}

// Named leagues.
var (
	LeagueUnknown     = Leagues[0]
	LeagueBronze      = Leagues[1]
	LeagueSilver      = Leagues[2]
	LeagueGold        = Leagues[3]
	LeaguePlatinum    = Leagues[4]
	LeagueDiamond     = Leagues[5]
	LeagueMaster      = Leagues[6]
	LeagueGrandmaster = Leagues[7]
	LeagueUnranked    = Leagues[8]
)

// leagueById returns the Leage specified by its id.
// LeagueUnknown is returned if id is unknown.
func leagueById(leagueId int64) *League {
	if id := int(leagueId); id >= 0 && id < len(Leagues) {
		return Leagues[id]
	}
	return LeagueUnknown
}

// Battle.net website language.
type BnetLang struct {
	Enum
	Code string // 2-letter language code, the way it appears in URLs.
}

// Slice of all Battle.net languages.
var BnetLangs = []*BnetLang{
	&BnetLang{Enum{"English"}, "en"},
	&BnetLang{Enum{"Chinese (Traditional)"}, "zn"},
	&BnetLang{Enum{"French"}, "fr"},
	&BnetLang{Enum{"German"}, "de"},
	&BnetLang{Enum{"Italian"}, "it"},
	&BnetLang{Enum{"Korean"}, "ko"},
	&BnetLang{Enum{"Polish"}, "pl"},
	&BnetLang{Enum{"Portuguese"}, "pt"},
	&BnetLang{Enum{"Russian"}, "ru"},
	&BnetLang{Enum{"Spanish"}, "es"},
}

// Named Battle.net languages.
var (
	BnetLangEnglish            = BnetLangs[0]
	BnetLangChineseTraditional = BnetLangs[1]
	BnetLangFrench             = BnetLangs[2]
	BnetLangGerman             = BnetLangs[3]
	BnetLangItalian            = BnetLangs[4]
	BnetLangKorean             = BnetLangs[5]
	BnetLangPolish             = BnetLangs[6]
	BnetLangPortuguese         = BnetLangs[7]
	BnetLangRussian            = BnetLangs[8]
	BnetLangSpanish            = BnetLangs[9]
)

// SC2 Realm (sub-region).
type Realm struct {
	Enum
}

// Slice of all realms.
var Realms = []*Realm{
	&Realm{Enum{"North America"}},
	&Realm{Enum{"Latin America"}},
	&Realm{Enum{"China"}},
	&Realm{Enum{"Europe"}},
	&Realm{Enum{"Russia"}},
	&Realm{Enum{"Korea"}},
	&Realm{Enum{"Taiwan"}},
	&Realm{Enum{"SEA"}},
	&Realm{Enum{"Unknown"}},
}

// Named realms.
var (
	RealmNorthAmerica = Realms[0]
	RealmLatinAmerica = Realms[1]
	RealmChina        = Realms[2]
	RealmEurope       = Realms[3]
	RealmRussia       = Realms[4]
	RealmKorea        = Realms[5]
	RealmTaiwan       = Realms[6]
	RealmSEA          = Realms[7]
	RealmUnknown      = Realms[8]
)

// SC2 Region.
type Region struct {
	Enum
	Code      string      // 2-letter region code
	DepotURL  *url.URL    // Region's depot server URL
	BnetURL   *url.URL    // Region's Battle.net website
	Realms    []*Realm    // Realms of the region, index+1 used in Details["playerList"]["toon"]["realm"]
	BnetLangs []*BnetLang // Available languages of the region's web page, first is the default language
}

// Slice of all regions, index used in Details["playerList"]["toon"]["region"]
var Regions = []*Region{
	&Region{Enum{"Unknown"}, "", mustPU("http://unknown.depot.battle.net:1119/"), mustPU("http://unknown.battle.net/"),
		[]*Realm{},
		[]*BnetLang{BnetLangEnglish}},
	&Region{Enum{"US"}, "US", mustPU("http://usb.depot.battle.net:1119/"), mustPU("http://us.battle.net/"),
		[]*Realm{RealmNorthAmerica, RealmLatinAmerica},
		[]*BnetLang{BnetLangEnglish, BnetLangSpanish, BnetLangPortuguese}},
	&Region{Enum{"Europe"}, "EU", mustPU("http://eub.depot.battle.net:1119/"), mustPU("http://eu.battle.net/"),
		[]*Realm{RealmEurope, RealmRussia},
		[]*BnetLang{BnetLangEnglish, BnetLangGerman, BnetLangFrench, BnetLangSpanish, BnetLangRussian, BnetLangItalian, BnetLangPolish}},
	&Region{Enum{"Korea"}, "KR", mustPU("http://krb.depot.battle.net:1119/"), mustPU("http://kr.battle.net/"),
		[]*Realm{RealmKorea, RealmTaiwan},
		[]*BnetLang{BnetLangKorean, BnetLangChineseTraditional}},
	&Region{Enum{"China"}, "CN", mustPU("http://cnb.depot.battle.net:1119/"), mustPU("http://www.battlenet.com.cn/"),
		[]*Realm{RealmChina},
		[]*BnetLang{BnetLangChineseTraditional}},
	&Region{Enum{"SEA"}, "SG", mustPU("http://sg.depot.battle.net:1119/"), mustPU("http://sea.battle.net/"),
		[]*Realm{RealmSEA},
		[]*BnetLang{BnetLangEnglish}},
	&Region{Enum{"Public Test"}, "XX", mustPU("http://xx.depot.battle.net:1119/"), mustPU("http://us.battle.net/"),
		[]*Realm{},
		[]*BnetLang{BnetLangEnglish}},
}

// mustPU parses the specified raw url string and panics if it is invalid.
func mustPU(rawurl string) *url.URL {
	if u, err := url.Parse(rawurl); err != nil {
		panic(err)
	} else {
		return u
	}
}

// Realm returns the realm of the region specified by its code.
func (r *Region) Realm(realmId int64) *Realm {
	if id := int(realmId) - 1; id >= 0 && id < len(r.Realms) {
		return r.Realms[id]
	}
	return RealmUnknown
}

// Named regions.
var (
	RegionUnknown    = Regions[0]
	RegionUS         = Regions[1]
	RegionEU         = Regions[2]
	RegionKR         = Regions[3]
	RegionCN         = Regions[4]
	RegionSEA        = Regions[5]
	RegionPublicTest = Regions[6]
)

// Map of regions, mapped from the 2-letter region code.
var regionMap = make(map[string]*Region)

func init() {
	// Build the regionMap map
	for _, r := range Regions {
		regionMap[r.Code] = r
	}
}

// regionByCode returns the Region specified by its 2-letter code.
// RegionUnknown is returned if code is unknown.
func regionByCode(code string) *Region {
	if r, ok := regionMap[code]; ok {
		return r
	}
	return RegionUnknown
}

// regionById returns the Region specified by its id.
// RegionUnknown is returned if id is unknown.
func regionById(regionId int64) *Region {
	if id := int(regionId); id >= 0 && id < len(Regions) {
		return Regions[id]
	}
	return RegionUnknown
}

// Cache handle, the identifier of a remote resource. A cache hande is a depedency.
type CacheHandle struct {
	Type   string  // Type of the resource, file extension.
	Region *Region // Region the resouce poins to.
	Digest string  // Hexadecimal representation of the SHA-256 digest of the content of the denoted resource.
}

// newCacheHandle parses the specified source string and returns a new CacheHandle.
func newCacheHandle(s string) *CacheHandle {
	c := &CacheHandle{Type: s[:4], Digest: hex.EncodeToString([]byte(s[8:]))}

	// Strip off leading zeros
	regionCode := ""
	for i := 4; i <= 8; i++ {
		if s[i] != 0 {
			regionCode = s[i:8]
			break
		}
	}
	c.Region = regionByCode(regionCode)

	return c
}

// FileName returns the file name denoted by the cache handle (with extension).
func (c *CacheHandle) FileName() string {
	return c.Digest + "." + c.Type
}

// RelativeFile returns the file denoted by the cache handle relative to the local cache folder.
func (c *CacheHandle) RelativeFile() string {
	return path.Join(c.Digest[0:2], c.Digest[2:4], c.FileName())
}

// StandardData returns the content of the resouce denoted by the cache handle if this is a standard data.
func (c *CacheHandle) StandardData() string {
	return standardCHData[c.Digest]
}

// Standard Cache Handle data. Maps from digest to the content of the denoted resource.
var standardCHData = map[string]string{
	"6de41503baccd05656360b6f027db88169fa1989bb6357b1b215a2547939f5fb": "Standard Data: Core.SC2Mod",
	"421c8aa0f3619b652d23a2735dfee812ab644228235e7a797edecfe8b67da30e": "Standard Data: Liberty.SC2Mod",
	"5c673e6cd2f1bf6e068fa59e2f9421f5debb91cb516aca3237d3b05fe7c7e9fa": "Standard Data: LibertyMulti.SC2Mod",
	"29198eca59d0f326f06c90c106348469415c08f9bd76da8413a7f9cd3bde8694": "Standard Data: Liberty.SC2Campaign",
	"1767383aa0f5b2eec7a1ec0eec8c98f10377fe2104c38a7e4fce44555aac07c7": "Standard Data: LibertyStory.SC2Campaign",
	"66093832128453efffbb787c80b7d3eec1ad81bde55c83c930dea79c4e505a04": "Standard Data: Swarm.SC2Mod",
	"881585946366c3c9d1499f38aba954216d3213de69554b9bee6b07311fb38336": "Standard Data: SwarmMulti.SC2Mod",
	"d92dfc48c484c59154270b924ad7d57484f2ab9a47621c7ab16431bf66c53b40": "Standard Data: Void.SC2Mod",
	"af23fed12efa6c496166dcf9441f802f33ab15172a87133dfae41ed603e3de16": "Standard Data: VoidMulti.SC2Mod",

	"d2b6f3851f19812ab614544137b896bb046c6a278e75d196604d6fdbbc69f42a": "Standard Data: Teams01.SC2Mod",
	"7f41411aa597f4b46440d42a563348bf53822d2a68112f0104f9b891f6f05ae1": "Standard Data: Teams02.SC2Mod",
	"6062b70f1485cf2320631d0f32a649c6b24af534ce087166d07cb7c4babdc92a": "Standard Data: Teams03.SC2Mod",
	"658e520aa5deb48866dc2b21b023daa9a291be4cf22fd9d785ca67f178132a87": "Standard Data: Teams04.SC2Mod",
	"bdf8a39d80f9d26947251efa9f29a4f5b98f6a190651f03051c7f11857d99b4c": "Standard Data: Teams05.SC2Mod",
	"b6ccab9e1dca6e10b65a4cf87569ace66c5743dd42cf30113f2b83c59ce7f1a2": "Standard Data: Teams06.SC2Mod",
	"c870fdaaf8f381a907f2ae8b195c4a472875428daab03145d3c678f62dd5f1b3": "Standard Data: Teams07.SC2Mod",
	"26b1b27647947a0f05ffe9a64f089b487a052de985a17310bea2041832a3dd85": "Standard Data: Teams08.SC2Mod",
	"0305203b64d6d35c80bf58030b0f497555cf7e31849726fd800853bb602415f3": "Standard Data: Teams09.SC2Mod",
	"b1c834f48b618b17caae9d1d174625bab89b84da581d94ef6ce7f5a6e8344802": "Standard Data: Teams10.SC2Mod",
	"8af9900ddeb1416a2619460124603198ae5bceda6387e0374f216a54955982a0": "Standard Data: Teams11.SC2Mod",
	"eaceeb172ee73b9650789d2fee249bac54ecee8f2b2204980929d69aeb135a44": "Standard Data: Teams12.SC2Mod",
	"e233a7b9e0e1ce10d2cba194fef783927df6ac128c5f73db881d64201e9ead0b": "Standard Data: Teams13.SC2Mod",
	"0e639dfeb6bbe18f5a859b5059dd6e296a7a19d1c902f538c250545fc7dd5658": "Standard Data: Teams14.SC2Mod",
	"1f720f0a950a29e6a77bddd4d3e4986faef7c6773066f61e9e4688242ec2538a": "Standard Data: Teams15.SC2Mod",
	"d50705d5859b6c52aead440f2a0bcedbfd811f06b259cc1733e5cefdf38aed82": "Standard Data: Teams16.SC2Mod",
}

// Game event ids
const (
	GmEIdPlayerLeave     = 25  // PlayerLeave game event id [ONLY UP TO BASEBUILD 23260; REPLACED BY USERLEAVE]
	GmEIdCmd             = 27  // CmdEvent game event id
	GmEIdSelDelta        = 28  // SelectionDelta game event id
	GmEIdCtrlGroupUpdate = 29  // ControlGroupUpdate game event id
	GmEIdCamUpdate       = 49  // CameraUpdate game event id
	GmEIdUsrLeave        = 101 // UserLeave game event id [ONLY FROM BASEBUILD 24764; REPLACES PLAYERLEAVE]
)

// Message event ids
const (
	MsgEIdChat = 0 // ChatMessage message event id
	MsgEIdPing = 1 // PingMessage message event id
)

// Tracker event ids
const (
	TrEIdPlayerSetup = 9 // PlayerSetup tracker event id [ONLY FROM BASEBUILD 27950]
)
