/*

The Rep type that models a replay (and everything in it).

*/

package rep

import (
	"errors"
	"io"

	"github.com/icza/mpq"
	"github.com/icza/s2prot"
)

var (
	// ErrInvalidRepFile means invalid replay file.
	ErrInvalidRepFile = errors.New("Invalid SC2Replay file")

	// ErrUnsupportedRepVersion means the replay file is valid but its version is not supported.
	ErrUnsupportedRepVersion = errors.New("Unsupported replay version")

	// ErrDecoding means decoding the replay file failed,
	// Most likely because replay file is invalid, but also might be due to an implementation bug
	ErrDecoding = errors.New("Decoding error")
)

// Rep describes a replay.
type Rep struct {
	m *mpq.MPQ // MPQ parser for reading the file

	protocol *s2prot.Protocol // Protocol to decode the replay

	Header   Header        // Replay header (replay game version and length)
	Details  Details       // Game details (overall replay details)
	InitData InitData      // Replay init data (the initial lobby)
	AttrEvts s2prot.Struct // Attributes events

	GameEvts    []s2prot.Event // Game events
	MessageEvts []s2prot.Event // Message events
	TrackerEvts []s2prot.Event // Tracker events

	GameEvtsErr    bool // Tells if decoding game events had errors
	MessageEvtsErr bool // Tells if decoding message events had errors
	TrackerEvtsErr bool // Tells if decoding tracker events had errors
}

// NewFromFile returns a new Rep constructed from a file.
// All types of events are decoded from the replay.
// The returned Rep must be closed with the Close method!
//
// ErrInvalidRepFile is returned if the specified name does not denote a valid SC2Replay file.
//
// ErrUnsupportedRepVersion is returned if the file exists and is a valid SC2Replay file but its version is not supported.
//
// ErrDecoding is returned if decoding the replay fails. This is most likely because the replay file is invalid, but also might be due to an implementation bug.
func NewFromFile(name string) (*Rep, error) {
	return NewFromFileEvts(name, true, true, true)
}

// NewFromFileEvts returns a new Rep constructed from a file, only the specified types of events decoded.
// The game, message and tracker tells if game events, message events and tracker events are to be decoded.
// Replay header, init data, details and attributes events are always decoded.
// The returned Rep must be closed with the Close method!
//
// ErrInvalidRepFile is returned if the specified name does not denote a valid SC2Replay file.
//
// ErrUnsupportedRepVersion is returned if the file exists and is a valid SC2Replay file but its version is not supported.
//
// ErrDecoding is returned if decoding the replay fails. This is most likely because the replay file is invalid, but also might be due to an implementation bug.
func NewFromFileEvts(name string, game, message, tracker bool) (*Rep, error) {
	m, err := mpq.NewFromFile(name)
	if err != nil {
		return nil, ErrInvalidRepFile
	}
	return newRep(m, game, message, tracker)
}

// New returns a new Rep using the specified io.ReadSeeker as the SC2Replay file source.
// All types of events are decoded from the replay.
// The returned Rep must be closed with the Close method!
//
// ErrInvalidRepFile is returned if the input is not a valid SC2Replay file content.
//
// ErrUnsupportedRepVersion is returned if the input is a valid SC2Replay file but its version is not supported.
//
// ErrDecoding is returned if decoding the replay fails. This is most likely because the input is invalid, but also might be due to an implementation bug.
func New(input io.ReadSeeker) (*Rep, error) {
	return NewEvts(input, true, true, true)
}

// NewEvts returns a new Rep using the specified io.ReadSeeker as the SC2Replay file source, only the specified types of events decoded.
// The game, message and tracker tells if game events, message events and tracker events are to be decoded.
// Replay header, init data, details and attributes events are always decoded.
// The returned Rep must be closed with the Close method!
//
// ErrInvalidRepFile is returned if the input is not a valid SC2Replay file content.
//
// ErrUnsupportedRepVersion is returned if the input is a valid SC2Replay file but its version is not supported.
//
// ErrDecoding is returned if decoding the replay fails. This is most likely because the input is invalid, but also might be due to an implementation bug.
func NewEvts(input io.ReadSeeker, game, message, tracker bool) (*Rep, error) {
	m, err := mpq.New(input)
	if err != nil {
		return nil, ErrInvalidRepFile
	}
	return newRep(m, game, message, tracker)
}

// newRep returns a new Rep constructed using the specified mpq.MPQ handler of the SC2Replay file, only the specified types of events decoded.
// The game, message and tracker tells if game events, message events and tracker events are to be decoded.
// Replay header, init data, details and attributes events are always decoded.
// The returned Rep must be closed with the Close method!
//
// ErrInvalidRepFile is returned if the specified name does not denote a valid SC2Replay file.
//
// ErrUnsupportedRepVersion is returned if the input is a valid SC2Replay file but its version is not supported.
//
// ErrDecoding is returned if decoding the replay fails. This is most likely because the input is invalid, but also might be due to an implementation bug.
func newRep(m *mpq.MPQ, game, message, tracker bool) (parsedRep *Rep, errRes error) {
	closeMPQ := true
	defer func() {
		// If returning due to an error, MPQ must be closed!
		if closeMPQ {
			m.Close()
		}

		// The input is completely untrusted and the decoding implementation omits error checks for efficiency:
		// Protect replay decoding:
		if r := recover(); r != nil {
			errRes = ErrDecoding
		}
	}()

	rep := Rep{m: m}

	rep.Header = Header{Struct: s2prot.DecodeHeader(m.UserData())}
	if rep.Header.Struct == nil {
		return nil, ErrInvalidRepFile
	}

	bb := rep.Header.BaseBuild()
	p := s2prot.GetProtocol(int(bb))
	if p == nil {
		return nil, ErrUnsupportedRepVersion
	}
	rep.protocol = p

	data, err := m.FileByHash(620083690, 3548627612, 4013960850) // "replay.details"
	if err != nil {
		return nil, ErrInvalidRepFile
	}
	rep.Details = Details{Struct: p.DecodeDetails(data)}

	data, err = m.FileByHash(3544165653, 1518242780, 4280631132) // "replay.initData"
	if err != nil {
		return nil, ErrInvalidRepFile
	}
	rep.InitData = newInitData(p.DecodeInitData(data))

	data, err = m.FileByHash(1306016990, 497594575, 2731474728) // "replay.attributes.events"
	if err != nil {
		return nil, ErrInvalidRepFile
	}
	rep.AttrEvts = p.DecodeAttributesEvts(data)

	if game {
		data, err = m.FileByHash(496563520, 2864883019, 4101385109) // "replay.game.events"
		if err != nil {
			return nil, ErrInvalidRepFile
		}
		rep.GameEvts, err = p.DecodeGameEvts(data)
		rep.GameEvtsErr = err != nil
	}

	if message {
		data, err = m.FileByHash(1089231967, 831857289, 1784674979) // "replay.message.events"
		if err != nil {
			return nil, ErrInvalidRepFile
		}
		rep.MessageEvts, err = p.DecodeMessageEvts(data)
		rep.MessageEvtsErr = err != nil
	}

	if tracker {
		data, err = m.FileByHash(1501940595, 4263103390, 1648390237) // "replay.tracker.events"
		if err != nil {
			return nil, ErrInvalidRepFile
		}
		rep.TrackerEvts, err = p.DecodeTrackerEvts(data)
		rep.TrackerEvtsErr = err != nil
	}

	// Everything went well, Rep is about to be returned, do not close MPQ
	// (it will be the caller's responsibility, done via Rep.Close()).
	closeMPQ = false

	return &rep, nil
}

// Close closes the Rep and its resources.
func (r *Rep) Close() error {
	return r.m.Close()
}
