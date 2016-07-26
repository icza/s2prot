/*

Package s2prot is a decoder/parser of Blizzard's StarCraft II replay file format (*.SC2Replay).

s2prot processes the "raw" data that can be decoded from replay files using an MPQ parser
such as https://github.com/icza/mpq.

The package is safe for concurrent use.

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
	fmt.Printf("Version: %d.%d.%d.%d\n", ver.Int("major"), ver.Int("minor"), ver.Int("revision"), ver.Int("build"))
	// Output: "Version: 2.1.9.34644"

Base build is part of the replay header:

	baseBuild := int(ver.Int("baseBuild"))
	fmt.Printf("Base build: %d\n", baseBuild)
	// Output: "Base build: 32283"

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
	// Output: "Map name: Hills of Peshkov"

Tip: use the encoding/json package to nicely format Struct values, e.g.:

	data, _ := json.MarshalIndent(header, "", "  ")
	fmt.Printf("Full Header:\n%s\n", data)

Output:

	Full Header:
	{
	  "elapsedGameLoops": 25811,
	  "signature": "StarCraft II replay\u001b11",
	  "type": 2,
	  "useScaledTime": false,
	  "version": {
	    "baseBuild": 32283,
	    "build": 34644,
	    "flags": 1,
	    "major": 2,
	    "minor": 1,
	    "revision": 9
	  }
	}

Information sources

- s2protocol: Blizzard's reference implementation in python: https://github.com/Blizzard/s2protocol

- s2protocol implementation of the Scelight project: https://github.com/icza/scelight/tree/master/src-app/hu/scelight/sc2/rep/s2prot


*/
package s2prot
