package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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

func movePlayer(idStr string, direction string, step string, conn *net.UDPConn) string {
	fmt.Println(idStr)
	player, exists := players[idStr]
	if !exists {
		fmt.Println("Player does not exist.")

	}
	stepSize, _ := strconv.Atoi(step)
	deltaX := map[string]int{"UP": -1 * stepSize, "DOWN": 1 * stepSize}[direction]
	newX := player.PlayerCoordinateX + deltaX
	if newX < 0 {
		newX = 0
	} else if newX >= sizeX {
		newX = sizeX - 1
	}
	deltaY := map[string]int{"LEFT": -1 * stepSize, "RIGHT": 1 * stepSize}[direction]
	newY := player.PlayerCoordinateY + deltaY
	if newY < 0 {
		newY = 0
	} else if newY >= sizeY {
		newY = sizeY - 1
	}
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
				// Check for Pokemon encounters
				for _, pokedex := range pokeDexWorld {
					CheckForPokemonEncounter(players[Id], pokedex)
				}
				// Save the updated list of all players
				playersList, err := LoadAllPlayerData()
				if err != nil {
					fmt.Println("Error loading player data:", err)
				} else {
					// Update the player list with the new data
					for i := range playersList {
						if playersList[i].ID == Id {
							playersList[i] = *player
							break
						}
					}
					// Save updated player data back to JSON file
					if err := SaveAllPlayerData(playersList); err != nil {
						fmt.Println("Error saving player data:", err)
					}
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

				// Load existing player data and append the new player
				playersList, err := LoadAllPlayerData()
				if err != nil {
					fmt.Println("Error loading player data:", err)
				} else {
					// Append the new player to the list
					playersList = append(playersList, *players[Id])
					// Save the updated player data back to JSON
					if err := SaveAllPlayerData(playersList); err != nil {
						fmt.Println("Error saving player data:", err)
					}
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
func LoadAllPlayerData() ([]Player, error) {
	// Read the JSON file containing all players' data
	data, err := ioutil.ReadFile("playersData.json")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// If file does not exist, return empty slice
			return []Player{}, nil
		}
		return nil, err
	}

	// Unmarshal the JSON data into a slice of Player structs
	var players []Player
	err = json.Unmarshal(data, &players)
	if err != nil {
		return nil, err
	}

	return players, nil
}
func SaveAllPlayerData(players []Player) error {
	// Marshal the players data into JSON format
	updatedData, err := json.MarshalIndent(players, "", "  ")
	if err != nil {
		return err
	}

	// Write the updated player data back to the file
	err = os.WriteFile("playersData.json", updatedData, 0666)
	if err != nil {
		return err
	}
	return nil
}
func HandlePlayerLogin(nameOrID string, x int, y int, conn *net.UDPConn, addr *net.UDPAddr) string {
	// Load all player data from the JSON file
	players, err := LoadAllPlayerData()
	if err != nil {
		fmt.Println("Error loading player data:", err)
		return "Error"
	}

	// Check if the player already exists
	var existingPlayer *Player
	for i := range players {
		if players[i].ID == nameOrID {
			existingPlayer = &players[i]
			break
		}
	}

	// If player doesn't exist, create a new player
	if existingPlayer == nil {
		fmt.Println("No existing player data found. Creating a new player...")
		existingPlayer = &Player{
			Name:              nameOrID,
			ID:                nameOrID,
			PlayerCoordinateX: x,
			PlayerCoordinateY: y,
			Addr:              addr,
		}
		// Append the new player to the player list
		players = append(players, *existingPlayer)
	} else {
		fmt.Println("Player data loaded successfully.")
		// Update the player's position and network address
		existingPlayer.PlayerCoordinateX = x
		existingPlayer.PlayerCoordinateY = y
		existingPlayer.Addr = addr
	}

	// Save the updated player data back to the JSON file
	if err := SaveAllPlayerData(players); err != nil {
		fmt.Println("Error saving player data:", err)
		return "Error"
	}

	// Display the game world after the player logs in
	return PokeCat(nameOrID, existingPlayer.Name, x, y, conn, addr)
}
