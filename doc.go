/*

Package s2prot is a decoder/parser of Blizzard's StarCraft II replay file format (*.SC2Replay).

s2prot processes the "raw" data that can be decoded from replay files using an MPQ parser.
https://github.com/icza/mpq is such an MPQ parser.

Usage

The below example code can be found in _example/repexample.go.

To use s2prot, we need an MPQ parser to get content from a replay.

	import "github.com/icza/mpq"
	
	m, err := mpq.NewFromFile("../../mpq/reps/automm.SC2Replay")
	if err != nil {
		panic(err)
	}
	defer m.Close()
	
Replay header (which is the MPQ User Data) can be decoded by s2prot.DecodeHeader(). Printing replay version:

	header := s2prot.DecodeHeader(m.UserData())
	ver := header.Structv("version")
	fmt.Printf("Version: %d.%d.%d.%d\n", ver.Int("major"), ver.Int("minor"), ver.Int("revison"), ver.Int("build"))

Base build is part of the replay header:

	baseBuild := int(ver.Int("baseBuild"))
	fmt.Printf("Base build: %d\n", baseBuild)

Which can be used to obtain the proper instance of Protocol:

	p := s2prot.GetProtocol(baseBuild)
	if p == nil {
		panic("Unknown base build!")
	}

Which can now be used to decode all other info in the replay. To decode the Details and print the map name:

	detailsData, err := m.FileByName("replay.details")
	if err != nil {
		panic(err)
	}
	details := p.DecodeDetails(detailsData)
	fmt.Println("Map name:", details.Stringv("title"))

Information sources

- s2protocol: Blizzard's reference implementation in python: https://github.com/Blizzard/s2protocol

- s2protocol implementation of the Scelight project: https://github.com/icza/scelight/tree/master/src-app/hu/scelight/sc2/rep/s2prot


*/
package s2prot
