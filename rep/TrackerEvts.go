/*

Type describing the tracker events.

*/

package rep

import (
	"math"

	"github.com/icza/s2prot"
)

const (
	// TrackerEvtIDPlayerStats is the ID of the Player Stats tracker event
	TrackerEvtIDPlayerStats = 0
	// TrackerEvtIDUnitBorn is the ID of the Unit Born tracker event
	TrackerEvtIDUnitBorn = 1
)

// TrackerEvts contains tracker events and some metrics and data calculated from them.
type TrackerEvts struct {
	// Evts contains the tracker events
	Evts []s2prot.Event

	// PIDPlayerDescMap is a PlayerDesc map mapped from player ID.
	PIDPlayerDescMap map[int64]*PlayerDesc
}

// PlayerDesc contains calculated, derived data from tracker events.
type PlayerDesc struct {
	// PlayerID is the ID of the player this PlayerDesc belongs to.
	PlayerID int64

	// Start location of the player
	StartLocX, StartLocY int64

	// StartDir is the start direction of the player, expressed in clock,
	// e.g. 1 o'clock, 3 o'clock etcc, in range of 1..12
	StartDir int
}

// init initializes / preprocesses the tracker events.
func (t *TrackerEvts) init(rep *Rep) {
	pidPlayerDescMap := make(map[int64]*PlayerDesc)
	t.PIDPlayerDescMap = pidPlayerDescMap

	getPD := func(pid int64) *PlayerDesc {
		pd := pidPlayerDescMap[pid]
		if pd == nil {
			pd = &PlayerDesc{PlayerID: pid}
			pidPlayerDescMap[pid] = pd
		}
		return pd
	}

	cx := rep.InitData.GameDescription.MapSizeX()
	cy := rep.InitData.GameDescription.MapSizeY()

	for _, e := range t.Evts {
		eid := e.Int("ID")
		if e.Loop() == 0 && eid == TrackerEvtIDUnitBorn {
			if isMainBuilding(e.Stringv("unitTypeName")) {
				pd := getPD(e.Int("controlPlayerId"))
				pd.StartLocX = e.Int("x")
				pd.StartLocX = e.Int("x")
				pd.StartDir = angleToClock(math.Atan2(float64(pd.StartLocY-cy), float64(pd.StartLocX-cx)))
			}
		}

		if eid == TrackerEvtIDPlayerStats {
			// TODO
		}
	}
}

// isMainBuilding tells if the unit type name denots a main building, that is
// one of Nexus, Command Center and Hatchery.
func isMainBuilding(unitTypeName string) bool {
	return unitTypeName == "Nexus" || unitTypeName == "CommandCenter" || unitTypeName == "Hatchery"
}

// angleToClock converts an angle given in radian to an hour clock value
// in the range of 1..12.
//
// Examples:
//  - PI/2 => 12 (o'clock)
//  - 0 => 3 (o'clock)
//  - PI => 9 (o'clock)
func angleToClock(angle float64) int {
	// The algorithm below computes clock value in the range of 0..11 where
	// 0 corresponds to 12.

	// 1 hour is PI/6 angle range
	const oneHour = math.Pi / 6

	// Shift by 3:30 (0 or 12 o-clock starts at 11:30)
	// and invert direction (clockwise):
	angle = -angle + oneHour*3.5

	// Put in range of 0..2*PI
	for angle < 0 {
		angle += oneHour * 12
	}
	for angle >= oneHour*12 {
		angle -= oneHour * 12
	}

	// And convert to a clock value:
	hour := int(angle / oneHour)
	if hour == 0 {
		return 12
	}
	return hour
}
