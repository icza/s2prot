/*

Type describing the replay header (replay game version and length).

*/

package rep

import (
	"fmt"
	"github.com/icza/s2prot"
	"time"
)

// Replay Header (replay game version and length).
type Header struct {
	s2prot.Struct

	version string // Lazily initialized full version string
}

// BaseBuild returns the base build.
func (h *Header) BaseBuild() int64 {
	return h.Int("version", "baseBuild")
}

// VersionString returns the full version string in the form of "major.minor.revision.build".
func (h *Header) VersionString() string {
	if h.version == "" {
		v := h.Version()
		h.version = fmt.Sprintf("%d.%d.%d.%d", v.Major(), v.Minor(), v.Revision(), v.Build())
	}
	return h.version
}

// Loops returns the elapsed game loops (game length in loops).
func (h *Header) Loops() int64 {
	return h.Int("elapsedGameLoops")
}

// Duration returns the game duration.
func (h *Header) Duration() time.Duration {
	// 1 second = 16 loops => 1 loop = 1/16 second = 62,500,000 ns
	return time.Duration(h.Loops() * 62500000)
}

// Signature returns the header signature.
// Should always be "StarCratII replay\u001b11".
func (h *Header) Signature() string {
	return h.Stringv("signature")
}

// UseScaledTime returns whether scaled time is used.
func (h *Header) UseScaledTime() bool {
	return h.Bool("useScaledTime")
}

// Type returns the type.
func (h *Header) Type() int64 {
	return h.Int("type")
}

// DataBuildNum returns the data build number.
func (h *Header) DataBuildNum() int64 {
	return h.Int("dataBuildNum")
}

// NgdpRootKey returns the data ngdp root key.
func (h *Header) NgdpRootKey() string {
	return h.Stringv("ngdpRootKey", "data")
}

// ReplayCompatibilityHash returns the replay compatibility hash.
func (h *Header) ReplayCompatibilityHash() string {
	return h.Stringv("replayCompatibilityHash", "data")
}

// Version returns the version of the replay.
func (h *Header) Version() Version {
	return Version{Struct: h.Structv("version")}
}

// Version of the replay.
type Version struct {
	s2prot.Struct
}

// Major returns the major part of the version.
func (v *Version) Major() int64 {
	return v.Int("major")
}

// Minor returns the minor part of the version.
func (v *Version) Minor() int64 {
	return v.Int("minor")
}

// Revision returns the revision part of the version.
func (v *Version) Revision() int64 {
	return v.Int("revision")
}

// Build returns the build part of the version.
func (v *Version) Build() int64 {
	return v.Int("build")
}

// BaseBuild returns the base build part of the version.
func (v *Version) BaseBuild() int64 {
	return v.Int("baseBuild")
}

// Flags returns the flags part of the version.
func (v *Version) Flags() int64 {
	return v.Int("flags")
}
