package router

import (
	"encoding/json"
	"log"
	"os"

	"onql/config"
	"onql/engine"
)

var Registries map[string]interface{}
var nats = engine.NatsClient{}

type Message struct {
	ID      string `json:"id"`      // who is sending (your keyword)
	RID     string `json:"rid"`     // unique request ID
	Target  string `json:"target"`  // which keyword to route to
	Payload string `json:"payload"` // arbitrary JSON payload
	Type    string `json:"type"`    // "request" or "response"
	// ConnId  string `json:"connId"`  // connection unique id
}

func init() {
	err := nats.Connect(config.Env("NATS_URL"), false)
	if err != nil {
		log.Printf("⚠️  could not connect to NATS: %v", err)
		return
	}
	
	// Determine the path to the registry JSON (env overrides default)
	registryPath := config.Env("EXTENSION_REGISTRY")

	// Initialize the map
	Registries = make(map[string]interface{})

	// Read the file
	data, err := os.ReadFile(registryPath)
	if err != nil {
		log.Printf("⚠️  could not read extension registry %q: %v", registryPath, err)
		return
	}

	// Parse JSON into the map
	if err := json.Unmarshal(data, &Registries); err != nil {
		log.Printf("⚠️  could not parse extension registry JSON: %v", err)
		return
	}

	log.Printf("✅  loaded %d registry entries from %s", len(Registries), registryPath)
	// connect with extensions channel
	setupExtensions()
	// StartStreamer()
}
