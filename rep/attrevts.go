/*

Type describing the attributes events.

*/

package rep

import "github.com/icza/s2prot"

// Attribute ID constants
const (
	// attrGameMode is the game mode attribute
	attrGameMode = "3009"
)

// scopeGlobal is the global scope.
const scopeGlobal = "16"

// AttrEvts contains game attributes.
type AttrEvts struct {
	s2prot.Struct

	// Scopes
	scopes s2prot.Struct
}

// NewAttrEvts creates a new attributes events from the specified Struct.
func NewAttrEvts(s s2prot.Struct) AttrEvts {
	a := AttrEvts{
		Struct: s,
		scopes: s.Structv("scopes"),
	}
	return a
}

// Source returns the source.
func (a *AttrEvts) Source() string {
	return a.Stringv("source")
}

// MapNamespace returns the map namespace.
func (a *AttrEvts) MapNamespace() string {
	return a.Stringv("mapNamespace")
}

// GameMode returns the game mode
func (a *AttrEvts) GameMode() *GameMode {
	if a.scopes == nil {
		return GameModeUnknown
	}
	return gameModeByAttrValue(a.scopes.Stringv(scopeGlobal, attrGameMode, "value"))
}
