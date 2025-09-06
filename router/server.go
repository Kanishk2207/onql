package router

import (
	"strings"
)

var connQueryMap = make(map[string]*func(string))

// var connMap = make(map[string]string)

func HandleServerRequest(query string, connUniqueId string, responder *func(string)) {
	//split query and create Message object
	parts := strings.Split(query, "\x1E")
	msg := &Message{
		ID:      "server",
		RID:     parts[0],
		Target:  parts[1],
		Payload: parts[2],
		Type:    "request",
		// ConnId:  connUniqueId,
	}
	// store the responder function in the map
	connQueryMap[msg.RID] = responder
	// connMap[connUniqueId] = msg.RID
	// connQueryMap[connUniqueId] = responder
	// route the message
	Route(msg)
}

// responsible for taking queries from tcp server and response them
func handleServerResponse(msg *Message) {
	rid := msg.RID
	//convert message to server response string
	response := msg.RID + "\x1E" + msg.ID + "\x1E" + string(msg.Payload)
	// print(response)
	//send back to user and it will be handled by responder function
	(*connQueryMap[rid])(response)
	// if msg.ID != "subscribe" {
	delete(connQueryMap, rid) // Remove the entry after responding
	// }
}
