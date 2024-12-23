package main

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
)

func main() {

	// Initialize pokedex with random Pokémon
	for k := 0; k < 20; k++ {
		pokemon, err := getRandomPokemon()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		pokedex := Pokedex{Pokemon: []Pokemon{*pokemon}}
		key := strconv.Itoa(k)
		pokeDexWorld[key] = &pokedex
		positionofPok(&pokedex)
		Pokeworld[pokedex.CoordinateX][pokedex.CoordinateY] = "E"
		fmt.Println("Pokemon:", pokemon.Name, "X:", pokedex.CoordinateX, "Y:", pokedex.CoordinateY)
	}

	// Set up the UDP server
	addr, err := net.ResolveUDPAddr("udp", ":8080")
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	buffer := make([]byte, 1024)

	for {
		// Read the incoming command from the client
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			panic(err)
		}

		// Extract player ID and command from the received message
		clientAddr := (strings.Replace(addr.String(), ".", "", -1)) // Remove periods from IP and port
		clientAddr = (strings.Replace(clientAddr, ":", "", -1))     // Remove colon from IP and port
		clientAddr = (strings.Replace(clientAddr, " ", "", -1))     // Remove spaces from IP and port
		idStr := clientAddr

		commands := string(buffer[:n])
		parts := strings.Split(commands, " ")

		switch strings.ToUpper(parts[0]) {
		case "CONNECT":
			// Call HandlePlayerLogin for new or existing players
			fmt.Println("Unique ID Int:", idStr)

			// Get player name from the command
			playerName := parts[1]

			// Generate random coordinates for the new player
			x := int(rand.Intn(sizeX))
			y := int(rand.Intn(sizeY))

			// Call HandlePlayerLogin to load or create player data
			PokeC := HandlePlayerLogin(playerName, x, y, conn, addr)

			// Send the connection message back to the client
			connectclient := fmt.Sprintf("Client connected: %s %s ID: %s", playerName, addr, idStr)
			_, err := conn.WriteToUDP([]byte(connectclient), addr)
			if err != nil {
				fmt.Println("Error sending connect message to client:", err)
			}

			// Send the updated world map to the player
			_, err = conn.WriteToUDP([]byte(PokeC), addr)
			if err != nil {
				fmt.Println("Error sending connect message to client:", err)
			}

		case "INFO":
			// Send player info
			Info := fmt.Sprintf("Player Info:%s", idStr)
			_, err := conn.WriteToUDP([]byte(Info), addr)
			if err != nil {
				fmt.Println("Error sending connect message to client:", err)
			}

		case "DISCONNECT":
			// Handle disconnect
			fmt.Println("Disconnected from server.")
			return

		case "UP", "DOWN", "LEFT", "RIGHT":
			// Handle player movement
			PokeK := movePlayer(idStr, strings.ToUpper(parts[0]), conn)
			fmt.Println(PokeK)
			_, err := conn.WriteToUDP([]byte(PokeK), addr)
			if err != nil {
				fmt.Println("Error sending connect message to client:", err)
			}

		case "INVENTORY":
			// Send inventory details to the player
			for _, inv := range players[idStr].Inventory {
				inventoryDetails := fmt.Sprintf("Player Inventory: Name: %s, Level: %d", inv.Name, inv.Level)
				_, err := conn.WriteToUDP([]byte(inventoryDetails), addr)
				if err != nil {
					fmt.Println("Error sending connect message to client:", err)
				}
			}
		}
	}
}
