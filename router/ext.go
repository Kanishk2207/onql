package router

import (
	"encoding/json"
	"fmt"
	"log"

	"onql/engine"
)

func setupExtensions() {
	// grab your NATS client singleton

	for topic := range Registries {
		// shadow topic so the closure captures this iteration’s value
		topic := topic
		subj := fmt.Sprintf("%s.send", topic)

		// subscribe using the engine’s Subscribe method
		_, err := nats.Subscribe(subj, func(m *engine.Msg) {
			handleMessageFromExtension(topic, m)
		})
		if err != nil {
			log.Printf("⚠️ subscribe %q failed: %v", subj, err)
		} else {
			log.Printf("✅ subscribed to %q", subj)
		}
	}
}

func handleMessageFromExtension(keyword string, msg *engine.Msg) {
	log.Printf("🔔 received message from %s: %s", keyword, msg.Subject)
	// Unmarshal the message payload into a Message struct
	var message Message
	if err := json.Unmarshal(msg.Data, &message); err != nil {
		log.Printf("failed to unmarshal message: %v", err)
		return
	}
	// call route functions here
	Route(&message)
}

// publish to listen channel
func handleExtensionMessage(msg *Message) {
	subj := fmt.Sprintf("%s.listen", msg.Target)
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("failed to marshal response: %v", err)
		return
	}
	if err := nats.Publish(subj, data); err != nil {
		log.Printf("failed to publish to %q: %v", subj, err)
	}
}

