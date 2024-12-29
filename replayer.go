package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/hypebeast/go-osc/osc"
)

// RecordedMessage matches the JSON structure created by the recorder
type RecordedMessage struct {
	Time    float64     `json:"time"`
	Address string      `json:"address"`
	Data    interface{} `json:"data"`
}

func main() {
	// This replayer expects a JSON file of RecordedMessage objects and an OSC target to send to.

	jsonFile := flag.String("file", "", "JSON file with recorded OSC messages (required)")
	address := flag.String("address", "127.0.0.1", "IP address to send OSC messages to")
	port := flag.Int("port", 8000, "Port to send OSC messages to")
	speedFactor := flag.Float64("speed", 1.0, "Adjust playback speed: 2.0 = double speed, 0.5 = half speed, etc.")
	flag.Parse()

	if *jsonFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	// 1) Open and read the JSON file
	file, err := os.Open(*jsonFile)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// 2) Parse the JSON into a slice of RecordedMessage
	var recordings []RecordedMessage
	rawData, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	if err := json.Unmarshal(rawData, &recordings); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(recordings) == 0 {
		log.Println("No messages found in JSON file. Exiting.")
		return
	}

	// 3) Sort by Time, just in case
	//    (If your recorder always generates sorted data, you can skip this step)
	//    But let's assume it is already sorted.

	// 4) Create an OSC Client to send messages
	client := osc.NewClient(*address, *port)
	log.Printf("Replaying %d messages to %s:%d at speed factor: %.2f\n",
		len(recordings), *address, *port, *speedFactor)

	// 5) Replay each message in order
	startTime := time.Now()
	firstMsgTime := recordings[0].Time

	for i := 0; i < len(recordings); i++ {
		msg := recordings[i]
		// The next line ensures we preserve relative timing
		// If firstMsgTime is not zero, we measure offset from the first message
		scheduledTime := (msg.Time - firstMsgTime) / *speedFactor

		// Sleep until it's time to send this message
		// We subtract how long we've been running so far
		elapsed := time.Since(startTime).Seconds()
		toWait := scheduledTime - elapsed
		if toWait > 0 {
			time.Sleep(time.Duration(toWait * float64(time.Second)))
		}

		// Build an OSC message
		oscMsg := osc.NewMessage(msg.Address)
		// msg.Data might be a slice, float64, int, string, etc.
		// We'll do some type-assertion logic to handle slices vs. single values
		switch dataVal := msg.Data.(type) {
		case []interface{}:
			for _, v := range dataVal {
				appendArg(oscMsg, v)
			}
		default:
			// If it wasn't a slice, we assume it's a single value
			appendArg(oscMsg, dataVal)
		}

		// Send the message
		err := client.Send(oscMsg)
		if err != nil {
			log.Printf("Failed to send OSC message: %v", err)
		} else {
			log.Printf("[%.4f] Sent -> %s %v\n", msg.Time, msg.Address, msg.Data)
		}
	}

	log.Println("Replay completed.")
}

// appendArg tries to append a typed value to the OSC message properly
func appendArg(msg *osc.Message, val interface{}) {
	switch v := val.(type) {
	case float64:
		// Check if integer disguised as float
		if float64(int64(v)) == v {
			// cast to int32 if you prefer that
			msg.Append(int32(v))
		} else {
			msg.Append(float32(v))
		}
	case string:
		msg.Append(v)
	case int:
		// convert to int32
		msg.Append(int32(v))
	case int32:
		msg.Append(v)
	case int64:
		// if small enough for int32, cast it, or just use float
		msg.Append(int32(v))
	case float32:
		msg.Append(v)
	case bool:
		msg.Append(v)
	case []interface{}:
		// If your replay logic hits this branch, handle each element
		for _, elem := range v {
			appendArg(msg, elem)
		}
	default:
		// fallback: treat as string if it's truly unknown
		msg.Append(fmt.Sprintf("%v", v))
	}
}
