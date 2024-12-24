package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"os"
	"strconv"
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

func printWorld(x, y int) string {
	world := "" // Initialize the world as an empty string
	for i := 0; i < sizeX; i++ {
		for j := 0; j < sizeY; j++ {
			// If the current position matches the player's coordinates
			if i == x && j == y {
				world += "P"

			} else if Pokeworld[i][j] == "E" {
				world += "E" // Append "E" (Entity) for Pokemon
			} else if Pokeworld[i][j] != "" {
				world += "O" // Append "O" for Other player space
			} else {
				world += "." // Append "." for Empty space
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

func movePlayer(id string, direction string, step string, conn *net.UDPConn, addr *net.UDPAddr) string {
	// Get the player based on the ID
	player, exists := players[id]
	if !exists {
		return fmt.Sprintf("Player with ID %s does not exist", id)
	}

	// Determine the movement step based on direction
	var deltaX, deltaY int
	stepSize, _ := strconv.Atoi(step)

	switch direction {
	case "UP":
		deltaX = -1 * stepSize
	case "DOWN":
		deltaX = 1 * stepSize
	case "LEFT":
		deltaY = -1 * stepSize
	case "RIGHT":
		deltaY = 1 * stepSize
	default:
		return fmt.Sprintf("Invalid direction: %s", direction)
	}

	// Calculate new position
	newX := player.PlayerCoordinateX + deltaX
	newY := player.PlayerCoordinateY + deltaY

	// Ensure the new position is within the world bounds
	if newX < 0 {
		newX = 0
	} else if newX >= sizeX {
		newX = sizeX - 1
	}

	if newY < 0 {
		newY = 0
	} else if newY >= sizeY {
		newY = sizeY - 1
	}

	// Clear the old position
	Pokeworld[player.PlayerCoordinateX][player.PlayerCoordinateY] = ""

	pokeCat := PokeCat(id, player.Name, newX, newY, conn, addr)
	return pokeCat
}

func positionofPok(pokedex *Pokedex) {

	x := rand.IntN(sizeX)

	y := rand.IntN(sizeY)

	pokedex.CoordinateX = int(x)
	pokedex.CoordinateY = int(y)
}

// PokeCat handles player placement, world updates, and encounters with other players or Pokemon
func PokeCat(Id string, playername string, x int, y int, conn *net.UDPConn, Addr *net.UDPAddr) string {
	// Check if the coordinates are within the bounds of Pokeworld
	if x >= 0 && x < sizeX && y >= 0 && y < sizeY {
		// Check if the position is already occupied
		if Pokeworld[x][y] == "" || Pokeworld[x][y] == "E" {
			// Place the player at the specified coordinates
			Pokeworld[x][y] = Id
			if player, exists := players[Id]; exists {
				// Player exists, update the existing player's fields
				player.Name = playername
				player.PlayerCoordinateX = x
				player.PlayerCoordinateY = y
				player.Addr = Addr
				// Check for Pokemon encounters
				for _, pokedex := range pokeDexWorld {
					CheckForPokemonEncounter(players[Id], pokedex)
				}
				// Save the updated player data to the player-specific file
				if err := SavePlayerData(player); err != nil {
					fmt.Println("Error saving player data:", err)
					return "Error"
				}
			} else {
				// Player does not exist, create a new one
				players[Id] = &Player{
					Name:              playername,
					ID:                Id,
					PlayerCoordinateX: x,
					PlayerCoordinateY: y,
					Addr:              Addr,
				}
				// Check for Pokemon encounters
				for _, pokedex := range pokeDexWorld {
					CheckForPokemonEncounter(players[Id], pokedex)
				}

				// Save the new player data to their specific file
				if err := SavePlayerData(players[Id]); err != nil {
					fmt.Println("Error saving new player data:", err)
					return "Error"
				}
			}

			fmt.Println("Player placed at", x, y)
			world := printWorld(x, y)
			return world
		} else {
			// If the position is occupied, start a battle with the other player
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

// SavePlayerData saves the data of a single player to their specific JSON file
func SavePlayerData(player *Player) error {
	// Marshal the player data into JSON format
	updatedData, err := json.MarshalIndent(player, "", "  ")
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s.json", player.Name)
	// Write the player data to their specific file
	err = os.WriteFile(filename, updatedData, 0666)
	if err != nil {
		return err
	}
	return nil
}

// LoadPlayerData loads the player data from their specific JSON file
func LoadPlayerData(playerID string) (*Player, error) {
	// Check if the player’s JSON file exists
	filename := fmt.Sprintf("%s.json", playerID)
	data, err := os.ReadFile(filename)
	if err != nil {
		// If the file doesn't exist, return an error
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("player %s does not exist", playerID)
		}
		return nil, err
	}

	// Unmarshal the JSON data into a Player struct
	var player Player
	err = json.Unmarshal(data, &player)
	if err != nil {
		return nil, err
	}

	return &player, nil
}

func HandlePlayerLogin(idStr, playerName string, x, y int, conn *net.UDPConn, addr *net.UDPAddr) *Player {
	// File name based on player's name
	playerFileName := playerName + ".json"
	player := &Player{ID: idStr, Name: playerName, PlayerCoordinateX: x, PlayerCoordinateY: y, Addr: addr}

	// Check if player file exists and load data if present
	if _, err := os.Stat(playerFileName); err == nil {
		// Load existing player data from the JSON file
		data, err := os.ReadFile(playerFileName)
		if err != nil {
			fmt.Println("Error reading player file:", err)
			return nil
		}
		err = json.Unmarshal(data, player)
		if err != nil {
			fmt.Println("Error unmarshalling player data:", err)
			return nil
		}
		fmt.Println("Loaded existing player:", playerName)
	} else {
		// New player, save to a JSON file
		data, err := json.Marshal(player)
		if err != nil {
			fmt.Println("Error marshalling player data:", err)
			return nil
		}
		err = os.WriteFile(playerFileName, data, 0644)
		if err != nil {
			fmt.Println("Error saving player file:", err)
			return nil
		}
		fmt.Println("Created new player file:", playerName)
	}

	// Save player data to global players map
	players[idStr] = player
	return player
}
