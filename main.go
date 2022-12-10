package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver
)

const sendChannel = 0

// Input CCs
const (
	inPlay1 = 1
	inPlay2 = 2
	inPlay3 = 3
	inPlayDrums = 4
	inRec1 = 5
	inRec2 = 6
	inRec3 = 7
	inStopAll = 8
)

// Output CCs
const (
	outPlayDrums = 1
	outStopAll = 2
	outStopTrack1 = 5
	outSelectTrack1 = 6
	outPlayTrack1 = 7
	outRecordTrack1 = 8
	outStopTrack2 = 9
	outSelectTrack2 = 10
	outPlayTrack2 = 11
	outRecordTrack2 = 12
	outStopTrack3 = 13
	outSelectTrack3 = 14
	outPlayTrack3 = 15
	outRecordTrack3 = 16
)

func main() {
	defer midi.CloseDriver()

	fmt.Printf("outports: %v\n",  midi.GetOutPorts())
	fmt.Printf("inports: %v\n",  midi.GetInPorts())

	mpkIn, err := midi.FindInPort("MPKmini2")
	if err != nil {
		fmt.Println("can't find MPKmini2")
		return
	}

	iacOut, err := midi.FindOutPort("IAC Driver Bus 1")
	if err != nil {
		fmt.Println("can't find IAC Out Driver on Bus 1")
		return
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	iacSend, err := midi.SendTo(iacOut)
	if err != nil {
		fmt.Printf("Cant send to IAC Out: %v\n", err)
		return
	}

	sendCC := func(cc uint8) error {
		fmt.Printf("sending: %v\n", cc)
		return iacSend(midi.ControlChange(sendChannel, cc, 127))
	}

	stop, err := midi.ListenTo(mpkIn, func(msg midi.Message, timestampms int32) {
		var ch, cc, val uint8
		switch {
		case msg.GetControlChange(&ch, &cc, &val):
			if (val > 0) {
				fmt.Printf("received control change %v\n", cc)
				mapCC(cc, sendCC)
			}
		default:
			// ignore
		}
	})
	if err != nil {
		fmt.Printf("Cant listen to MPK in: %v\n", err)
		return
	}

	fmt.Println("connected to IAC and MPKmini and listening for CCs...")

	<- done

	fmt.Println("shutting down...")
	stop()
}

func mapCC(cc uint8, send func(uint8) error) {
	var err1, err2, err3, err4 error
	switch cc {
	case inPlay1:
		err1 = send(outStopTrack2)
		err2 = send(outStopTrack3)
		err3 = send(outSelectTrack1)
		err4 = send(outPlayTrack1)
	case inPlay2:
		err1 = send(outStopTrack1)
		err2 = send(outStopTrack3)
		err3 = send(outSelectTrack2)
		err4 = send(outPlayTrack2)
	case inPlay3:
		err1 = send(outStopTrack1)
		err2 = send(outStopTrack2)
		err3 = send(outSelectTrack3)
		err4 = send(outPlayTrack3)
	case inRec1:
		err1 = send(outStopTrack2)
		err2 = send(outStopTrack3)
		err3 = send(outSelectTrack1)
		err4 = send(outRecordTrack1)
	case inRec2:
		err1 = send(outStopTrack1)
		err2 = send(outStopTrack3)
		err3 = send(outSelectTrack2)
		err4 = send(outRecordTrack2)
	case inRec3:
		err1 = send(outStopTrack1)
		err2 = send(outStopTrack2)
		err3 = send(outSelectTrack3)
		err4 = send(outRecordTrack3)
	case inPlayDrums:
		err1 = send(outStopTrack1)
		err2 = send(outStopTrack2)
		err3 = send(outStopTrack3)
		err4 = send(outPlayDrums)
	case inStopAll:
		err1 = send(outStopAll)
	default:
		// Ignore the rest
	}

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		fmt.Printf("send err: %v, %v, %v, %v\n", err1, err2, err3, err4)
	}
} 