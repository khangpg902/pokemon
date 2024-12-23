package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	sizeX int = 10
	sizeY int = 10
)

type Stats struct {
	HP         float32 `json:"HP"`
	Attack     float32 `json:"Attack"`
	Defense    float32 `json:"Defense"`
	Speed      int     `json:"Speed"`
	Sp_Attack  float32 `json:"Sp_Attack"`
	Sp_Defense float32 `json:"Sp_Defense"`
}

type GenderRatio struct {
	MaleRatio   float32 `json:"MaleRatio"`
	FemaleRatio float32 `json:"FemaleRatio"`
}

type Profile struct {
	Height      float32     `json:"Height"`
	Weight      float32     `json:"Weight"`
	CatchRate   float32     `json:"CatchRate"`
	GenderRatio GenderRatio `json:"GenderRatio"`
	EggGroup    string      `json:"EggGroup"`
	HatchSteps  int         `json:"HatchSteps"`
	Abilities   string      `json:"Abilities"`
}

type DamegeWhenAttacked struct {
	Element     string  `json:"Element"`
	Coefficient float32 `json:"Coefficient"`
}

type Moves struct {
	Name    string   `json:"Name"`
	Element []string `json:"Element"`
	Acc     int      `json:"Acc"`
}

type Pokemon struct {
	Id                 int                  `json:"Id"`
	Name               string               `json:"Name"`
	Elements           []string             `json:"Elements"`
	EV                 int                  `json:"EV"`
	Stats              Stats                `json:"Stats"`
	Profile            Profile              `json:"Profile"`
	DamegeWhenAttacked []DamegeWhenAttacked `json:"DamegeWhenAttacked"`
	EvolutionLevel     int                  `json:"EvolutionLevel"`
	NextEvolution      string               `json:"NextEvolution"`
	Moves              []Moves              `json:"Moves"`
	Exp                int                  `json:"Exp"`
	Level              int                  `json:"Level"`
}

type Pokedex struct {
	Pokemon     []Pokemon `json:"Pokemon"`
	CoordinateX int
	CoordinateY int
}

type Player struct {
	Name              string    `json:"Name"`
	ID                string    `json:"ID"`
	PlayerCoordinateX int       `json:"PlayerCoordinateX"`
	PlayerCoordinateY int       `json:"PlayerCoordinateY"`
	Inventory         []Pokemon `json:"Inventory"`
	IsTurn            bool
	Addr              *net.UDPAddr
	sync.Mutex
}

var players = make(map[string]*Player)
var Pokeworld [sizeX][sizeY]string
var pokeDexWorld = make(map[string]*Pokedex)
var inventory1 = make(map[string]*Pokemon)
var inventory []Pokemon

func main() {

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

		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			panic(err)
		}

		clientAddr := (strings.Replace(addr.String(), ".", "", -1)) // This removes all periods      // This includes IP and port
		clientAddr = (strings.Replace(clientAddr, ":", "", -1))     // This removes the colon         // This includes IP and port
		clientAddr = (strings.Replace(clientAddr, " ", "", -1))     // This removes all spaces        // This includes IP and port                // This converts the length to a string // This includes IP and port

		idStr := clientAddr

		commands := string(buffer[:n])
		parts := strings.Split(commands, " ")

		switch strings.ToUpper(parts[0]) {
		case "CONNECT":
			fmt.Println("Unique ID Int:", idStr)
			players[idStr] = &Player{Name: parts[1], ID: idStr}
			x := int(rand.IntN(sizeX))
			y := int(rand.IntN(sizeY))

			PokeC := PokeCat(idStr, parts[1], x, y, conn, addr)
			connectclient := fmt.Sprintf("Client connected: %s %s ID: %s", parts[1], addr, idStr)
			_, err := conn.WriteToUDP([]byte(connectclient), addr)
			if err != nil {
				fmt.Println("Error sending connect message to client:", err)
			}
			_, err = conn.WriteToUDP([]byte(PokeC), addr)
			if err != nil {
				fmt.Println("Error sending connect message to client:", err)
			}
			// Handle connection...
		case "INFO":
			Info := fmt.Sprintf("Player Info:%s", idStr)
			_, err := conn.WriteToUDP([]byte(Info), addr)
			if err != nil {
				fmt.Println("Error sending connect message to client:", err)
			}
			// Display player info...
		case "DISCONNECT":
			fmt.Println("Disconnected from server.")
			return
		case "UP", "DOWN", "LEFT", "RIGHT":
			PokeK := movePlayer(idStr, strings.ToUpper(parts[0]), conn)
			fmt.Println(PokeK)
			_, err := conn.WriteToUDP([]byte(PokeK), addr)
			if err != nil {
				fmt.Println("Error sending connect message to client:", err)
			}
		case "INVENTORY":
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

func printWorld(x, y int) string {
	world := "" // Initialize the world as an empty string
	for i := 0; i < sizeX; i++ {
		for j := 0; j < sizeY; j++ {
			// If the current position matches the player's coordinates
			if i == x && j == y {
				world += "P"

			} else if Pokeworld[i][j] == "E" {
				world += "E" // Append "E" (Entity) for Pokemon
			} else {
				world += "-" // Append "." for Empty space
			}
		}
		world += "\n" // New line after each row
	}
	return world
}

func CheckForPokemonEncounter(player *Player, pokemon *Pokedex) {
	for _, pokedex := range pokemon.Pokemon {
		if player.PlayerCoordinateX == pokemon.CoordinateX && player.PlayerCoordinateY == pokemon.CoordinateY {
			PassingPokemontoInventory(&pokedex, player)
			fmt.Println("Pokemon encountered:", pokedex.Name)

		}
	}
}

func movePlayer(idStr string, direction string, conn *net.UDPConn) string {
	player, exists := players[idStr]
	if !exists {
		fmt.Println("Player does not exist.")

	}
	deltaX := map[string]int{"UP": -1, "DOWN": 1}[direction]
	newX := player.PlayerCoordinateX + deltaX
	deltaY := map[string]int{"LEFT": -1, "RIGHT": 1}[direction]
	newY := player.PlayerCoordinateY + deltaY
	Pokeworld[player.PlayerCoordinateX][player.PlayerCoordinateY] = ""

	PokeK := PokeCat(idStr, player.Name, newX, newY, conn, player.Addr)
	return PokeK

}
func positionofPok(pokedex *Pokedex) {

	x := rand.IntN(sizeX)

	y := rand.IntN(sizeY)

	pokedex.CoordinateX = int(x)
	pokedex.CoordinateY = int(y)
}

func PokeCat(Id string, playername string, x int, y int, conn *net.UDPConn, Addr *net.UDPAddr) string {
	// Check if the coordinates are within the bounds of Pokeworld.
	if x >= 0 && x < sizeX && y >= 0 && y < sizeY {
		// Check if the position is already occupied.
		if Pokeworld[x][y] == "" || Pokeworld[x][y] == "E" {
			// Place the player at the specified coordinates.
			Pokeworld[x][y] = Id
			if player, exists := players[Id]; exists {
				// Player exists, update the existing player's fields
				player.Name = playername
				player.PlayerCoordinateX = x
				player.PlayerCoordinateY = y
				player.Addr = Addr
				for _, pokedex := range pokeDexWorld {
					CheckForPokemonEncounter(players[Id], pokedex)
				}
				fileName := fmt.Sprintf("pokeInventory%s.json", Id)
				PassingPlayertoJson(fileName, players[Id])
			} else {
				// Player does not exist, create a new one
				players[Id] = &Player{
					Name:              playername,
					ID:                Id,
					PlayerCoordinateX: x,
					PlayerCoordinateY: y,
					Addr:              Addr,
				}
				for _, pokedex := range pokeDexWorld {
					CheckForPokemonEncounter(players[Id], pokedex)
				}
				fileName := fmt.Sprintf("pokeInventory%s.json", Id)
				PassingPlayertoJson(fileName, players[Id])
			}

			fmt.Println("Player placed at", x, y)
			world := printWorld(x, y)
			return world
		} else {
			battleScene(players[Id], players[Pokeworld[x][y]], conn, Addr, players[Pokeworld[x][y]].Addr)
			return "Battle"
		}
	}
	return ""
}

func PassingPokemontoInventory(pokemon *Pokemon, player *Player) {
	player.Lock() // Lock the player instance
	defer player.Unlock()
	player.Inventory = append(player.Inventory, *pokemon)
}
func PassingPlayertoJson(filename string, player *Player) {

	updatedData, err := json.MarshalIndent(player, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
	}
	if err := os.WriteFile(filename, updatedData, 0666); err != nil {
		fmt.Println("Error:", err)
	}

}
func getRandomPokemon() (*Pokemon, error) {
	// Read the JSON file
	data, err := os.ReadFile("../lib/pokedex.json")
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON data into a slice of Pokemon structs
	var pokemons []Pokemon
	err = json.Unmarshal(data, &pokemons)
	if err != nil {
		return nil, err
	}

	// Generate a random index
	index := rand.IntN(len(pokemons))

	// Return the randomly selected Pokemon
	return &pokemons[index], nil
}
