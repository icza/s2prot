/*

Defines the map which contains the python sources of different builds.

If there are identical build specs, the Builds map will contain entry only for the oldest base build number.

The Duplicates map should be checked to get the oldest base build number (if there is any).

*/

package build

// Holds the python sources mapped from base build.
// In case of identical build specs,
// this only contains entry for the oldest base build number.
var Builds = make(map[int]string)

// Holds duplicates / identical build specs.
// Key is a (newer) base build number, value is an older build number.
// In case of duplicates, Builds only contains entry for the oldest base build number.
var Duplicates = make(map[int]int)
