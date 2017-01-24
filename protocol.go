/*

The exported Protocol type.

*/

package s2prot

import (
	"bufio"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/icza/s2prot/build"
)

// Default base build to be used to decode replay headers (any can be used).
// Practical to specify the latest as that is the one most likely always needed.
var defBaseBuild int

func init() {
	// Find highest base build and set it as the default:
	for k := range build.Builds {
		if defBaseBuild < k {
			defBaseBuild = k
		}
	}
}

// EvtType describes a named event data structure type.
type EvtType struct {
	Id     int    // Id of the event
	Name   string // Name of the event
	typeid int    // Type id of the event data structure
}

// The Protocol type which implements the data structures and their decoding
// from SC2Replay files defined by s2protocol.
type Protocol struct {
	baseBuild int // Base build

	typeInfos []typeInfo // Type info slice, decoding instructions for all the types

	hasTrackerEvents bool // Tells if this protocol has/handles tracker events

	gameEvtTypes         []EvtType // Game event type descriptors; index is event id
	gameEventidTypeid    int       // The typeid of the NNet.Game.EEventId enum
	messageEvtTypes      []EvtType // Message event type descriptors; index is event id
	messageEventidTypeid int       // The typeid of the NNet.Game.EMessageId enum
	trackerEvtTypes      []EvtType // Tracker event type descriptors; index is event id
	trackerEventidTypeid int       // The typeid of the NNet.Replay.Tracker.EEventId enum

	svaruint32Typeid int // The typeid of NNet.SVarUint32 (the type used to encode gameloop deltas)

	replayUseridTypeid int // The typeid of NNet.Replay.SGameUserId (the type used to encode player ids) [from base build 24764, before that player id is stored instead of user id!]

	replayHeaderTypeid   int // The typeid of NNet.Replay.SHeader (the type used to store replay game version and length)
	gameDetailsTypeid    int // The typeid of NNet.Game.SDetails (the type used to store overall replay details)
	replayInitdataTypeid int // The typeid of NNet.Replay.SInitData (the type used to store the initial lobby)
}

var (
	// Holds the already parsed Protocols mapped from base build.
	protocols = make(map[int]*Protocol)
	// Mutex protecting access of the protocols map
	protMux = &sync.Mutex{}
)

// GetProtocol returns the Protocol for the specified base build.
// nil return value indicates unknown/unsupported base build.
func GetProtocol(baseBuild int) *Protocol {
	protMux.Lock()
	defer protMux.Unlock()

	return getProtocol(baseBuild)
}

// getProtocol returns the Protocol for the specified base build.
// nil return value indicates unknown/unsupported base build.
// protMux must be locked when this function is called.
func getProtocol(baseBuild int) *Protocol {
	// Check if protocol is already parsed:
	p, ok := protocols[baseBuild]
	if ok {
		// Note that ok only means a value exists for baseBuild but it might be nil
		// in case we didn't find it or failed to parse it in an earlier call.
		return p
	}

	// Not yet parsed, check if an original base build (not duplicate):
	src, ok := build.Builds[baseBuild]
	if ok {
		p = parseProtocol(src, baseBuild)
		protocols[baseBuild] = p
		return p
	}

	// Either a duplicate or an Unknown base build. Check for duplicate:
	origBaseBuild, ok := build.Duplicates[baseBuild]
	if ok {
		// It's a duplicate. Get the original (will load original if needed).
		// origBasebuild surely exists (build.Duplicates contains valid entries, ensured by test!)
		// but parsing it may (still) fail, so check for nil:
		if op := getProtocol(origBaseBuild); op != nil {
			// Copy / clone protocol with proper base build:
			p = new(Protocol)
			*p = *op
			p.baseBuild = baseBuild
		}
	}
	// (else it's not a duplicate: it's an Unknown base build; p remains nil)

	// Even if p is nil: still store nil value so we'll know this earlier next time
	protocols[baseBuild] = p
	return p
}

// parseProtocol parses a Protocol from its python source.
// nil is returned if parsing error occurs.
func parseProtocol(src string, baseBuild int) *Protocol {
	// Protect the parsing logic:
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Failed to parse protocol source %d: %v\n", baseBuild, r)
		}
		// nil will be returned by parseProtocol()
	}()

	p := Protocol{baseBuild: baseBuild, hasTrackerEvents: baseBuild >= 24944}

	scanner := bufio.NewScanner(strings.NewReader(src))

	var line string

	// Helper function to seek to a line with a given prefix:
	seek := func(prefix string) {
		for scanner.Scan() {
			line = scanner.Text()
			if strings.HasPrefix(line, prefix) {
				return
			}
		}
		panic(fmt.Sprintf(`Couldn't find "%s"`, prefix))
	}

	// Helper function to parse the last integer number from the current line with form: "some_name = int_value"
	parseInt := func() int {
		i := strings.LastIndex(line, "=")
		if i < 0 {
			panic("Can't find '=' in line")
		}
		n, err := strconv.Atoi(strings.TrimSpace(line[i+1:]))
		if err != nil {
			panic(err)
		}
		return n
	}

	// Helper function to parse an event types slice
	parseEvtTypes := func(stripPref, stripPost string) []EvtType {
		var err error
		em := make(map[int]EvtType) // First build it in a map
		maxEid := -1                // Max Event id
		for scanner.Scan() {
			line = scanner.Text()
			if line == "}" {
				break
			}
			e := EvtType{}
			i := strings.IndexByte(line, ':')
			e.Id, err = strconv.Atoi(strings.TrimSpace(line[:i]))
			if err != nil {
				panic(err)
			}
			line = line[i+1:]
			i = strings.IndexByte(line, '(') + 1
			j := strings.IndexByte(line, ',')
			e.typeid, err = strconv.Atoi(strings.TrimSpace(line[i:j]))
			if err != nil {
				panic(err)
			}
			i = strings.IndexByte(line, '\'') + 1
			line = line[i:]
			i = strings.IndexByte(line, '\'')
			e.Name = line[len(stripPref) : i-len(stripPost)]
			em[e.Id] = e
			if e.Id > maxEid {
				maxEid = e.Id
			}
		}

		// And now create a slice from the map:
		es := make([]EvtType, maxEid+1)
		for k, v := range em {
			es[k] = v
		}
		return es
	}
	_ = parseEvtTypes

	// Decode typeinfos
	seek("typeinfos")
	// Use a large local variable
	typeInfos := make([]typeInfo, 0, 256)
	for scanner.Scan() {
		line = scanner.Text()
		if line == "]" {
			break
		}
		typeInfos = append(typeInfos, parseTypeInfo(line))
	}
	// And now copy a trimmed version of this to Protocol (typeInfo is a relatively large struct):
	p.typeInfos = make([]typeInfo, len(typeInfos))
	copy(p.typeInfos, typeInfos)

	// Decode game event types
	seek("game_event_types")
	p.gameEvtTypes = parseEvtTypes("NNet.Game.S", "Event")

	seek("game_eventid_typeid")
	p.gameEventidTypeid = parseInt()

	// Decode message event types
	seek("message_event_types")
	p.messageEvtTypes = parseEvtTypes("NNet.Game.S", "Message")

	seek("message_eventid_typeid")
	p.messageEventidTypeid = parseInt()

	if p.hasTrackerEvents {
		// Decode track event types
		seek("tracker_event_types")
		p.trackerEvtTypes = parseEvtTypes("NNet.Replay.Tracker.S", "Event")

		seek("tracker_eventid_typeid")
		p.trackerEventidTypeid = parseInt()
	}

	seek("svaruint32_typeid")
	p.svaruint32Typeid = parseInt()

	// From basebuild 24764 user id is present, before that player id
	if baseBuild >= 24764 {
		seek("replay_userid_typeid")
	} else {
		seek("replay_playerid_typeid")
	}
	p.replayUseridTypeid = parseInt()

	seek("replay_header_typeid")
	p.replayHeaderTypeid = parseInt()

	seek("game_details_typeid")
	p.gameDetailsTypeid = parseInt()

	seek("replay_initdata_typeid")
	p.replayInitdataTypeid = parseInt()

	return &p
}

// DecodeHeader decodes and returns the replay header.
// Panics if decoding fails.
func DecodeHeader(contents []byte) Struct {
	p := GetProtocol(defBaseBuild)
	if p == nil {
		panic("Default protocol is not available!")
	}

	contents = contents[4:] // 3c 00 00 00 (might be part of the MPQ header and not the user data)

	d := newVersionedDec(contents, p.typeInfos)

	v, ok := d.instance(p.replayHeaderTypeid).(Struct)
	if !ok {
		return nil
	}

	return v
}

// DecodeDetails decodes and returns the game details.
// Panics if decoding fails.
func (p *Protocol) DecodeDetails(contents []byte) Struct {
	d := newVersionedDec(contents, p.typeInfos)

	v, ok := d.instance(p.gameDetailsTypeid).(Struct)
	if !ok {
		return nil
	}

	return v
}

// DecodeInitData decodes and returns the replay init data.
// Panics if decoding fails.
func (p *Protocol) DecodeInitData(contents []byte) Struct {
	d := newBitPackedDec(contents, p.typeInfos)

	v, ok := d.instance(p.replayInitdataTypeid).(Struct)
	if !ok {
		return nil
	}

	return v
}

// DecodeAttributesEvts decodes and returns the attributes events.
// Panics if decoding fails.
func (p *Protocol) DecodeAttributesEvts(contents []byte) Struct {
	s := Struct{}

	if len(contents) == 0 {
		return s
	}

	bb := &bitPackedBuff{
		contents:  contents,
		bigEndian: false, // Note: the only place where little endian order is used.
	}

	// Source is only present from 1.2 and onward (base build 17326)
	if p.baseBuild >= 17326 {
		s["source"] = bb.readBits(8)
	}
	s["mapNamespace"] = bb.readBits(32)

	bb.readBits(32) // Attributes count

	scopes := Struct{}
	for !bb.EOF() {
		attr := Struct{}
		attr["namespace"] = bb.readBits(32)
		attrid := bb.readBits(32)
		attr["attrid"] = attrid
		attrscope := bb.readBits(8)

		// SIDENOTE: My feeling is that since this (decoding attributes events) is the only place
		// where little endian order is used, readAligned() implementation should will the slice backwards.
		// That way no reverse would be needed.
		vb := bb.readAligned(4)
		// Reverse and strip leading zeros
		vb[0], vb[3] = vb[3], vb[0]
		vb[1], vb[2] = vb[2], vb[1]
		for i := 3; i >= 0; i-- {
			if vb[i] == 0 {
				vb = vb[i+1:]
				break
			}
		}
		attr["value"] = string(vb)

		sattrscope := strconv.FormatInt(attrscope, 10)

		scope, ok := scopes[sattrscope].(Struct)
		if !ok {
			scope = Struct{}
			scopes[sattrscope] = scope
		}
		scope[strconv.FormatInt(attrid, 10)] = attr
	}
	s["scopes"] = scopes

	return s
}

// Type decoder defines the most basic methods a decoder must support.
type decoder interface {
	EOF() bool
	byteAlign()
	instance(typeid int) interface{}
}

// DecodeGameEvts decodes and returns the game events.
// In case of a decoding error, successfully decoded events are still returned along with an error.
func (p *Protocol) DecodeGameEvts(contents []byte) ([]Event, error) {
	return p.decodeEvts(newBitPackedDec(contents, p.typeInfos), p.gameEventidTypeid, p.gameEvtTypes, true)
}

// DecodeMessageEvts decodes and returns the message events.
// In case of a decoding error, successfully decoded events are still returned along with an error.
func (p *Protocol) DecodeMessageEvts(contents []byte) ([]Event, error) {
	return p.decodeEvts(newBitPackedDec(contents, p.typeInfos), p.messageEventidTypeid, p.messageEvtTypes, true)
}

// DecodeTrackerEvts decodes and returns the tracker events.
// In case of a decoding error, successfully decoded events are still returned along with an error.
func (p *Protocol) DecodeTrackerEvts(contents []byte) ([]Event, error) {
	return p.decodeEvts(newVersionedDec(contents, p.typeInfos), p.trackerEventidTypeid, p.trackerEvtTypes, false)
}

// decodeEvts decodes a series of events.
// In case of a decoding error, successfully decoded events are still returned along with an error.
func (p *Protocol) decodeEvts(d decoder, evtidTypeid int, etypes []EvtType, decUserId bool) (events []Event, err error) {
	deltaTypeid := p.svaruint32Typeid    // Local var for efficiency
	useridTypeid := p.replayUseridTypeid // Local var for efficiency

	events = make([]Event, 0, 256) // This is most likely overestimation for messages events but underestimation for all other even types

	// Protect the events decoding:
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Failed to decode events: %v", r)
			log.Println(err)
		}
		// Successfully decoded events will be returned
	}()

	var (
		loop   int64
		userid interface{}
	)

	for !d.EOF() {
		delta := d.instance(deltaTypeid).(Struct)
		// delta has one key-value pair:
		for _, v := range delta {
			loop += v.(int64)
		}

		if decUserId {
			userid = d.instance(useridTypeid)
		}

		evtid := d.instance(evtidTypeid).(int64)
		evtType := &etypes[evtid]

		// Decode the event data structure:
		e := Event{Struct: d.instance(evtType.typeid).(Struct), EvtType: evtType}
		// Copy to / duplicate data in Struct so Struct.String() includes them too
		e.Struct["id"] = evtid
		e.Struct["name"] = evtType.Name
		e.Struct["loop"] = loop
		if decUserId {
			e.Struct["userid"] = userid
		}

		events = append(events, e)

		// The next event is byte-aligned:
		d.byteAlign()
	}

	return
}
