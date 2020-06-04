package main

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber"
	"github.com/gofiber/template/html"
	"socket/socket"
)

// Basic chat message object
type MessageObject struct {
	Data string `json:"data"`
	From string `json:"from"`
	To   string `json:"to"`
}


func main() {
	// The key for the map is message.to
	clients := make(map[string]string)

	// Start a new Fiber application
	app := fiber.New()

	// Setup the middleware to retrieve the data sent in first GET request
	app.Use(func(c *fiber.Ctx) {
		c.Locals("user_id", c.Query("user_id"))
		c.Next()
	})

	// Multiple event handling supported
	socket.On(socket.EventConnect, func(ep *socket.EventPayload) {
		fmt.Println("fired connect 1")
	})

	socket.On(socket.EventConnect, func(ep *socket.EventPayload) {
		fmt.Println("fired connect 2")
	})

	socket.On(socket.EventMessage, func(ep *socket.EventPayload) {
		fmt.Println("fired message: " + string(ep.Data))
	})

	socket.On(socket.EventDisconnect, func(ep *socket.EventPayload) {
		fmt.Println("fired disconnect" + ep.Error.Error())
	})

	// Websocket route init
	app.Get("/ws", socket.New(func(kws *socket.Websocket) {
		// Retrieve user id from the middleware (optional)
		userId :=  fmt.Sprintf("%v", kws.Locals("user_id"))
		// Every websocket connection has an optional session key => value storage
		kws.SetAttribute("user_id", userId)

		// On connect event. Notify when comes a new connection
		kws.OnConnect = func() {
			// Add the connection to the list of the connected clients
			// The UUID is generated randomly
			clients[userId] = kws.UUID
			//Broadcast to all the connected users the newcomer
			kws.Broadcast([]byte("New user connected: "+userId+" and UUID: "+kws.UUID), true)
			//Write welcome message
			kws.Emit([]byte("Hello user: " + userId + " and UUID: " + kws.UUID))
		}

		// On message event
		kws.OnMessage = func(data []byte) {

			message := MessageObject{}
			json.Unmarshal(data, &message)
			// Emit the message directly to specified user
			err := kws.EmitTo(clients[message.To], data)
			if err != nil {
				fmt.Println(err)
			}
		}
	}))

	socket.On("close", func(payload *socket.EventPayload) {
		fmt.Println("fired close " + payload.SocketAttributes["user_id"])
	})
	/*app.Static("/", "./assets", fiber.Static{
		Compress:  true,
		ByteRange: false,
		Browse:    false,
		Index:     "index.html",
	})*/
	app.Settings.Templates = html.New("./views", ".html")

	app.Get("/", func(c *fiber.Ctx) {
		userId := fmt.Sprintf("%v", c.Locals("user_id"))
		fmt.Println(userId)
		c.Render("index", fiber.Map{
			"id": userId,
		})
	})
	// Start the application on port 3000
	err := app.Listen(3021)
	if err != nil {
		panic(err)
	}
}
