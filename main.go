package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/hypebeast/go-osc/osc"
)

// SchemeFunc is a function type that takes an OSC address & args
type SchemeFunc func(addr string, args []interface{}) map[string]interface{}

var schemes = map[string]SchemeFunc{
	"dirt_basic": func(addr string, args []interface{}) map[string]interface{} {
		// Take first element only
		if len(args) > 0 {
			return map[string]interface{}{
				"address": addr,
				"data":    args[0],
			}
		}
		return map[string]interface{}{
			"address": addr,
			"data":    nil,
		}
	},
	"dirt_strip": func(addr string, args []interface{}) map[string]interface{} {
		// Return only odd-indexed elements
		var oddArgs []interface{}
		for i, v := range args {
			if i%2 == 1 {
				oddArgs = append(oddArgs, v)
			}
		}
		return map[string]interface{}{
			"address": addr,
			"data":    oddArgs,
		}
	},
	"basic": func(addr string, args []interface{}) map[string]interface{} {
		// Keep everything
		return map[string]interface{}{
			"address": addr,
			"data":    args,
		}
	},
	"only_numbers": func(addr string, args []interface{}) map[string]interface{} {
		// Filter only numeric values
		var numeric []interface{}
		for _, v := range args {
			switch v.(type) {
			case int32, float32, float64, int64, int:
				numeric = append(numeric, v)
			}
		}
		return map[string]interface{}{
			"address": addr,
			"data":    numeric,
		}
	},
}

// RecordedMessage is what we'll store for each incoming OSC message.
type RecordedMessage struct {
	Time    float64     `json:"time"`
	Address string      `json:"address"`
	Data    interface{} `json:"data"`
}

var (
	schemeFn   SchemeFunc
	messages   []RecordedMessage
	startTime  time.Time
	quantized  bool
	fileOutput string
)

func main() {
	// Parse CLI flags
	addressFlag := flag.String("address", "", "IP address to listen on (required)")
	portFlag := flag.Int("port", 0, "Port to listen on (required)")
	fileFlag := flag.String("file", "", "Path to the output JSON file (required)")
	schemeFlag := flag.String("scheme", "", "Scheme for processing incoming OSC messages (required)")
	repeatersFlag := flag.String("repeaters", "", "Comma-separated list of ports to forward messages to (optional)")
	quantizedFlag := flag.Bool("quantized", false, "Quantize timing so first message is at time=0 (optional)")
	flag.Parse()

	// Validate required flags
	if *addressFlag == "" || *portFlag == 0 || *fileFlag == "" || *schemeFlag == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Grab the scheme function
	sf, found := schemes[*schemeFlag]
	if !found {
		log.Fatalf("Unknown scheme: %s\n", *schemeFlag)
	}
	schemeFn = sf

	// Save other flag values
	fileOutput = *fileFlag
	quantized = *quantizedFlag

	// Set up repeaters
	var repeaterClients []*osc.Client
	if *repeatersFlag != "" {
		ports := strings.Split(*repeatersFlag, ",")
		for _, p := range ports {
			p = strings.TrimSpace(p)
			portNum, err := strconv.Atoi(p)
			if err != nil {
				log.Printf("Invalid repeater port: %s\n", p)
				continue
			}
			client := osc.NewClient(*addressFlag, portNum)
			repeaterClients = append(repeaterClients, client)
		}
	}

	// Create an OSC dispatcher
	dispatcher := osc.NewStandardDispatcher()
	// Catch-all handler for any OSC address: "*"
	dispatcher.AddMsgHandler("*", func(msg *osc.Message) {
		handleOSCMessage(msg, repeaterClients)
	})

	// Create an OSC server
	serverAddr := fmt.Sprintf("%s:%d", *addressFlag, *portFlag)
	server := &osc.Server{
		Addr:       serverAddr,
		Dispatcher: dispatcher,
	}

	// Start time
	startTime = time.Now()

	// Handle Ctrl+C / SIGINT to gracefully save
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		finalizeAndExit()
	}()

	// Start listening
	log.Printf("Listening for OSC on %s ...\n", serverAddr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error listening for OSC: %v", err)
	}
}

// handleOSCMessage processes each incoming OSC message
func handleOSCMessage(msg *osc.Message, repeaters []*osc.Client) {
	t := time.Since(startTime).Seconds()

	// If quantized and this is the first message, reset startTime
	if quantized && len(messages) == 0 {
		startTime = time.Now()
		t = 0
	}

	// Apply the chosen scheme
	processed := schemeFn(msg.Address, msg.Arguments)

	record := RecordedMessage{
		Time:    t,
		Address: processed["address"].(string),
		Data:    processed["data"],
	}
	messages = append(messages, record)

	// Print out for debugging
	log.Printf("[%.4f] %s => %v\n", t, record.Address, record.Data)

	// Forward to repeaters (if any)
	for _, client := range repeaters {
		fwdMsg := osc.NewMessage(msg.Address)
		for _, arg := range msg.Arguments {
			fwdMsg.Append(arg)
		}
		go client.Send(fwdMsg)
	}
}

// finalizeAndExit writes messages to a JSON file and exits
func finalizeAndExit() {
	outFile, err := os.Create(fileOutput)
	if err != nil {
		log.Printf("Failed to create file %s: %v\n", fileOutput, err)
		os.Exit(1)
	}
	defer outFile.Close()

	enc := json.NewEncoder(outFile)
	enc.SetIndent("", "  ")
	if err := enc.Encode(messages); err != nil {
		log.Printf("Failed to write JSON: %v\n", err)
		os.Exit(1)
	}

	log.Printf("Saved %d messages to %s\n", len(messages), fileOutput)
	os.Exit(0)
}
