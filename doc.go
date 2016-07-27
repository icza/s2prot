/*

Package s2prot is a decoder/parser of Blizzard's StarCraft II replay file format (*.SC2Replay).

s2prot processes the "raw" data that can be decoded from replay files using an MPQ parser
such as https://github.com/icza/mpq.

The package is safe for concurrent use.

High-level Usage

The package s2prot/rep provides enumerations and types to model data structures
of StarCraft II replays (*.SC2Replay) decoded by the s2prot package. These provide a higher level overview
and much easier to use.

The below example code can be found in https://github.com/icza/s2prot/blob/master/_example/rep.go.

To open and parse a replay:

	import "github.com/icza/s2prot/rep"

	r, err := rep.NewFromFile("../../mpq/reps/lotv.SC2Replay")
	if err != nil {
		fmt.Println("%v\n", err)
		return
	}
	defer r.Close()

And that's all! We now have all the info from the replay! Printing some of it:

	fmt.Printf("Version:        %v\n", r.Header.VersionString())
	fmt.Printf("Loops:          %d\n", r.Header.Loops())
	fmt.Printf("Length:         %v\n", r.Header.Duration())
	fmt.Printf("Map:            %s\n", r.Details.Title())
	fmt.Printf("Game events:    %d\n", len(r.GameEvts))
	fmt.Printf("Message events: %d\n", len(r.MessageEvts))
	fmt.Printf("Tracker events: %d\n", len(r.TrackerEvts))

	fmt.Println("Players:")
	for _, p := range r.Details.Players() {
		fmt.Printf("\tName: %-20s, Race: %c, Team: %d, Result: %s\n",
			p.Name, p.Race().Letter, p.TeamId()+1, p.Result())
	}

Output:

	Version:        3.2.2.42253
	Loops:          13804
	Length:         14m22.75s
	Map:            Magma Mines
	Game events:    10461
	Message events: 32
	Tracker events: 1758
	Players:
		Name: <NoGy>IMBarabba     , Race: P, Team: 1, Result: Defeat
		Name: <NoGy>Nova          , Race: T, Team: 1, Result: Defeat
		Name: <9KingS>BiC         , Race: T, Team: 2, Result: Victory
		Name: <9KingS>DakotaFannin, Race: P, Team: 2, Result: Victory

Tip: the Struct type defines a String() method which returns a nicely formatted JSON representation;
this is what most type are "made of": 

	fmt.Printf("Full Header:\n%v\n", r.Header)

Output:

	Full Header:
	{
	  "dataBuildNum": 42253,
	  "elapsedGameLoops": 13804,
	  "ngdpRootKey": {
	    "data": "\ufffd \ufffd\ufffd\ufffd\ufffd]\ufffd\ufffd\ufffd\ufffd\ufffd\ufffd..."
	  },
	  "replayCompatibilityHash": {
	    "data": "\ufffd\ufffd\ufffd'⌂\u001fv\ufffd%\rEĪѓX"
	  },
	  "signature": "StarCraft II replay\u001b11",
	  "type": 2,
	  "useScaledTime": true,
	  "version": {
	    "baseBuild": 42253,
	    "build": 42253,
	    "flags": 1,
	    "major": 3,
	    "minor": 2,
	    "revision": 2
	  }
	}


Low-level Usage

The below example code can be found in https://github.com/icza/s2prot/blob/master/_example/s2prot.go.

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
	fmt.Printf("Version: %d.%d.%d.%d\n",
		ver.Int("major"), ver.Int("minor"), ver.Int("revision"), ver.Int("build"))
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

Tip: We can of course print the whole decoded header which is a Struct: 

	fmt.Printf("Full Header:\n%v\n", header)

Which yields a JSON text similar to the one posted above (at High-level Usage).


Information sources

- s2protocol: Blizzard's reference implementation in python: https://github.com/Blizzard/s2protocol

- s2protocol implementation of the Scelight project: https://github.com/icza/scelight/tree/master/src-app/hu/scelight/sc2/rep/s2prot


*/
package s2prot
