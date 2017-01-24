/*
This example shows how to use the rep package to easily extract information
from a StarCraft II (*.SC2Replay) file.
*/
package main

import (
	"fmt"

	"github.com/icza/s2prot/rep"
)

func main() {
	//r, err := rep.NewFromFileEvts("../../mpq/reps/automm.SC2Replay", true, true, true)
	r, err := rep.NewFromFile("../../mpq/reps/lotv.SC2Replay")
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		return
	}
	defer r.Close()

	fmt.Printf("Version:        %v\n", r.Header.VersionString())
	fmt.Printf("Loops:          %d\n", r.Header.Loops())
	fmt.Printf("Length:         %v\n", r.Header.Duration())
	fmt.Printf("Map:            %s\n", r.Details.Title())
	fmt.Printf("Game events:    %d\n", len(r.GameEvts))
	fmt.Printf("Message events: %d\n", len(r.MessageEvts))
	fmt.Printf("Tracker events: %d\n", len(r.TrackerEvts))

	fmt.Println("Players:")
	for _, p := range r.Details.Players() {
		fmt.Printf("\tName: %-20s, Race: %c, Team: %d, Result: %v\n",
			p.Name, p.Race().Letter, p.TeamId()+1, p.Result())
	}
	fmt.Printf("Full Header:\n%v\n", r.Header)

	//fmt.Printf("%s\n", r.Details.String())
	//fmt.Printf("%s\n", r.InitData.String())
	//fmt.Printf("%s\n", r.AttrEvts.String())
}
