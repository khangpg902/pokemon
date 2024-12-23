package main

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

func main() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano()) // Seed for randomness

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

		// Extract player ID from the received message (based on the client's address)
		clientAddr := strings.Replace(addr.String(), ".", "", -1) // Remove periods from IP and port
		clientAddr = strings.Replace(clientAddr, ":", "", -1)     // Remove colon from IP and port
		idStr := clientAddr

		commands := string(buffer[:n])
		parts := strings.Split(commands, " ")

		// Handle the different commands
		switch strings.ToUpper(parts[0]) {
		case "CONNECT":
			// Get player name from the command
			playerName := parts[1]

			// Generate random coordinates for the new player
			x := rand.Intn(sizeX)
			y := rand.Intn(sizeY)

			// Call HandlePlayerLogin to load or create player data
			player := HandlePlayerLogin(idStr, playerName, x, y, conn, addr)

			// Send the connection message back to the client
			connectMessage := fmt.Sprintf("Client connected: %s %s ID: %s", playerName, addr, idStr)
			_, err := conn.WriteToUDP([]byte(connectMessage), addr)
			if err != nil {
				fmt.Println("Error sending connect message to client:", err)
			}

			// Send the updated world information (player's coordinates)
			worldMessage := fmt.Sprintf("World: Player at X: %d, Y: %d", player.PlayerCoordinateX, player.PlayerCoordinateY)
			_, err = conn.WriteToUDP([]byte(worldMessage), addr)
			if err != nil {
				fmt.Println("Error sending world map to client:", err)
			}

		case "INFO":
			// Send player info
			playerInfo := fmt.Sprintf("Player Info: %s", idStr)
			_, err := conn.WriteToUDP([]byte(playerInfo), addr)
			if err != nil {
				fmt.Println("Error sending player info to client:", err)
			}

		case "DISCONNECT":
			// Handle disconnect
			fmt.Println("Disconnected from server.")
			return

		case "UP", "DOWN", "LEFT", "RIGHT":
			// Handle player movement
			moveMessage := movePlayer(idStr, strings.ToUpper(parts[0]), conn, addr)
			_, err := conn.WriteToUDP([]byte(moveMessage), addr)
			if err != nil {
				fmt.Println("Error sending movement message to client:", err)
			}

		case "INVENTORY":
			// Send inventory details to the player
			inventoryDetails := getPlayerInventory(idStr)
			_, err := conn.WriteToUDP([]byte(inventoryDetails), addr)
			if err != nil {
				fmt.Println("Error sending inventory to client:", err)
			}
		}
	}
}
