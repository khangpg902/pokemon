package main

import (
	"fmt"
	"math/rand/v2"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func battleScene(player1 *Player, player2 *Player, conn *net.UDPConn, addr1, addr2 *net.UDPAddr) {

	if player1 == nil {
		fmt.Println("Error: player1 is nil")
		return
	}
	if conn == nil {
		fmt.Println("Error: conn is nil")
		return
	}
	if addr1 == nil {
		fmt.Println("Error: addr1 is nil")
		return
	}
	if len(player1.Inventory) < 3 {
		fmt.Println("Player 1 has less than 3 pokemons")
		conn.WriteToUDP([]byte("You have less than 3 pokemons"), addr1)
		return
	} else if len(player2.Inventory) < 3 {
		fmt.Println("Player 2 has less than 3 pokemons")
		conn.WriteToUDP([]byte("You have less than 3 pokemons"), addr2)
		return
	}

	// Player 1 select 3 Pokemons
	fmt.Println("Player 1 please select 3 pokemons from:")
	conn.WriteToUDP([]byte("Player 1 please select 3 pokemons from:\n"), addr1)
	for i := range player1.Inventory {
		printPokemonInfo(i, player1.Inventory[i])
		conn.WriteToUDP([]byte(fmt.Sprintf("%d: %s\n", i, player1.Inventory[i].Name)), addr1)
	}
	player1Pokemons := selectPokemon(player1, conn, addr1)

	// Player 2 select 3 Pokemons
	fmt.Println("Player 2 please select 3 pokemons from:")
	conn.WriteToUDP([]byte("Player 2 please select 3 pokemons from:\n"), addr2)
	for i := range player2.Inventory {
		printPokemonInfo(i, player2.Inventory[i])
		conn.WriteToUDP([]byte(fmt.Sprintf("%d: %s\n", i, player2.Inventory[i].Name)), addr2)
	}
	player2Pokemons := selectPokemon(player2, conn, addr2)

	allBattlingPokemons := append(*player1Pokemons, *player2Pokemons...)
	firstAttacker := getFirstAttacker(allBattlingPokemons)
	var firstDefender *Pokemon

	fmt.Println("Battle start!")
	conn.WriteToUDP([]byte("Battle start!\n"), addr1)
	conn.WriteToUDP([]byte("Battle start!\n"), addr2)
	var winner, loser *Player
	if isContain(*player1Pokemons, *firstAttacker) {
		firstDefender = getFirstDefender(*player2Pokemons)
		fmt.Println("Player 1 goes first")
		conn.WriteToUDP([]byte("Player 1 goes first\n"), addr1)
		conn.WriteToUDP([]byte("Player 1 goes first\n"), addr2)
		player1.IsTurn = true
		player2.IsTurn = false
	} else {
		firstDefender = getFirstDefender(*player1Pokemons)
		fmt.Println("Player 2 goes first")
		conn.WriteToUDP([]byte("Player 2 goes first\n"), addr1)
		conn.WriteToUDP([]byte("Player 2 goes first\n"), addr2)
		player1.IsTurn = false
		player2.IsTurn = true
	}

	// The battle loop
	var player1Pokemon = firstAttacker
	var player2Pokemon = firstDefender
	for {
		if player1.IsTurn {
			if !isAlive(player1Pokemon) {
				fmt.Println(player1Pokemon.Name, "is dead")
				conn.WriteToUDP([]byte(fmt.Sprintf("%s is dead\n", player1Pokemon.Name)), addr1)
				player1Pokemon = switchPokemon(*player1Pokemons, conn, addr1)
				if player1Pokemon == nil {
					fmt.Println("Player 1 has no pokemon left")
					fmt.Println("Player 1 lost")
					conn.WriteToUDP([]byte("You have no pokemon left. You lost.\n"), addr1)
					conn.WriteToUDP([]byte("Player 1 has no pokemon left. Player 2 wins.\n"), addr2)
					break
				} else {
					fmt.Println("Player 1 switched to", player1Pokemon.Name)
					conn.WriteToUDP([]byte(fmt.Sprintf("Player 1 switched to %s\n", player1Pokemon.Name)), addr1)
				}
			}

			fmt.Printf("Player 1 turn. Your current pokemon is %s. Choose your action:\n", player1Pokemon.Name)
			conn.WriteToUDP([]byte(fmt.Sprintf("Your turn. Your current pokemon is %s. Choose your action:\n", player1Pokemon.Name)), addr1)
			command := readCommands(conn, addr1)
			switch command {
			case "attack":
				attack(player1Pokemon, player2Pokemon, conn, addr1)
			case "switch":
				displaySelectedPokemons(*player1Pokemons, conn, addr1)
				player1Pokemon = switchToChosenPokemon(*player1Pokemons, conn, addr1)
				fmt.Println("Player 1 switched to", player1Pokemon.Name)
				conn.WriteToUDP([]byte(fmt.Sprintf("Switched to %s\n", player1Pokemon.Name)), addr1)
			case "surrender":
				surrender(player1, player2, conn, player1.Addr, player2.Addr)
				winner = player2
				loser = player1
				break
			case "?":
				displayCommandsList(conn, addr1)
			}

			player1.IsTurn = false
			player2.IsTurn = true
		}

		if player2.IsTurn {
			if !isAlive(player2Pokemon) {
				fmt.Println(player2Pokemon.Name, "is dead")
				conn.WriteToUDP([]byte(fmt.Sprintf("%s is dead\n", player2Pokemon.Name)), addr2)
				player2Pokemon = switchPokemon(*player2Pokemons, conn, addr2)
				if player2Pokemon == nil {
					fmt.Println("Player 2 has no pokemon left")
					fmt.Println("Player 2 lost")
					conn.WriteToUDP([]byte("You have no pokemon left. You lost.\n"), addr2)
					conn.WriteToUDP([]byte("Player 2 has no pokemon left. Player 1 wins.\n"), addr1)
					break
				} else {
					fmt.Println("Player 2 switched to", player2Pokemon.Name)
					conn.WriteToUDP([]byte(fmt.Sprintf("Player 2 switched to %s\n", player2Pokemon.Name)), addr2)
				}
			}

			fmt.Printf("Player 2 turn. Your current pokemon is %s. Choose your action:\n", player2Pokemon.Name)
			conn.WriteToUDP([]byte(fmt.Sprintf("Your turn. Your current pokemon is %s. Choose your action:\n", player2Pokemon.Name)), addr2)
			command := readCommands(conn, addr2)
			switch command {
			case "attack":
				attack(player2Pokemon, player1Pokemon, conn, addr2)
			case "switch":
				displaySelectedPokemons(*player2Pokemons, conn, addr2)
				player2Pokemon = switchToChosenPokemon(*player2Pokemons, conn, addr2)
				fmt.Println("Player 2 switched to", player2Pokemon.Name)
				conn.WriteToUDP([]byte(fmt.Sprintf("Switched to %s\n", player2Pokemon.Name)), addr2)
			case "surrender":
				surrender(player2, player1, conn, player2.Addr, player1.Addr)
				winner = player1
				loser = player2
				break
			case "?":
				displayCommandsList(conn, addr2)
			}

			player2.IsTurn = false
			player1.IsTurn = true
		}

		time.Sleep(500 * time.Millisecond)
	}
	addExp(winner, loser, conn)
}

func surrender(loser *Player, winner *Player, conn *net.UDPConn, loserAddr, winnerAddr *net.UDPAddr) {
	fmt.Printf("%s surrendered!\n", loser.Name)
	fmt.Printf("%s won!\n", winner.Name)
	conn.WriteToUDP([]byte("You surrendered. You lost.\n"), loserAddr)
	conn.WriteToUDP([]byte("The enemy surrendered. You win.\n"), winnerAddr)
}
func attack(attacker *Pokemon, defender *Pokemon, conn *net.UDPConn, addr *net.UDPAddr) {
	// Calculate the damage
	var dmg float32
	var attackerMove = chooseAttack(*attacker)
	var attackerDmg = attacker.Stats.Attack
	fmt.Println(attacker.Name, "chose", attackerMove.Name, "to attack", defender.Name)
	conn.WriteToUDP([]byte(fmt.Sprintf("%s chose %s to attack %s\n", attacker.Name, attackerMove.Name, defender.Name)), addr)

	switch attackerMove.Name {
	case "Normal":
		dmg = float32(attackerDmg - defender.Stats.Defense)
	case "Special":
		attackingElement := attackerMove.Element
		dmgWhenAttacked := defender.DamegeWhenAttacked
		defendingElement := []string{}
		for _, element := range dmgWhenAttacked {
			defendingElement = append(defendingElement, element.Element)
		}
		highestCoefficient := float32(0)

		// Check for the highest coefficient
		for i, element := range defendingElement {
			if isContain(attackingElement, element) {
				if highestCoefficient < dmgWhenAttacked[i].Coefficient {
					highestCoefficient = dmgWhenAttacked[i].Coefficient
				}
			}
		}

		// If the attacker has an element that the defender doesn't have, set the coefficient to 1
		for _, element := range defendingElement {
			if !isContain(attackingElement, element) && highestCoefficient < 1 {
				highestCoefficient = 1
			}
		}

		dmg = attackerDmg*highestCoefficient - defender.Stats.Sp_Defense
	}

	if dmg < 0 {
		dmg = 0
	}
	fmt.Println(attacker.Name, "attacked", defender.Name, "with", attackerMove.Name, "and dealt", dmg, "damage")
	conn.WriteToUDP([]byte(fmt.Sprintf("%s attacked %s with %s and dealt %.2f damage\n", attacker.Name, defender.Name, attackerMove.Name, dmg)), addr)
	defender.Stats.HP -= dmg
}

func chooseAttack(pokemon Pokemon) Moves {
	var n int
	//70% Normal Attack
	if rand.IntN(1000) <= 700 {
		n = 0 //Normal attack
	} else {
		n = 1 // Sp attack
	}
	return pokemon.Moves[n]
}

func isContain[T any](arr []T, element T) bool {
	for _, a := range arr {
		if reflect.DeepEqual(a, element) {
			return true
		}
	}
	return false
}

func getFirstAttacker(allBattlingPokemons []Pokemon) *Pokemon {
	var highestSpeed = 0
	var choosenPokemonIndex = 0
	for i, pokemon := range allBattlingPokemons {
		if pokemon.Stats.Speed > highestSpeed {
			highestSpeed = pokemon.Stats.Speed
			choosenPokemonIndex = i
		}
	}

	return &allBattlingPokemons[choosenPokemonIndex]
}

func getFirstDefender(defenderPokemons []Pokemon) *Pokemon {
	var highestSpeed = 0
	var choosenPokemonIndex = 0
	for i, pokemon := range defenderPokemons {
		if pokemon.Stats.Speed > highestSpeed {
			highestSpeed = pokemon.Stats.Speed
			choosenPokemonIndex = i
		}
	}

	return &defenderPokemons[choosenPokemonIndex]
}

func isAlive(pokemon *Pokemon) bool {
	return pokemon.Stats.HP > 0
}

func switchPokemon(pokemonsList []Pokemon, conn *net.UDPConn, addr *net.UDPAddr) *Pokemon {
	for i := 0; i < len(pokemonsList); i++ {
		if isAlive(&pokemonsList[i]) {
			return &pokemonsList[i]
		}
	}
	return nil
}

func displayCommandsList(conn *net.UDPConn, addr *net.UDPAddr) {
	conn.WriteToUDP([]byte("List of commands:\n"), addr)
	conn.WriteToUDP([]byte("\tattack: to attack the opponent\n"), addr)
	conn.WriteToUDP([]byte("\tswitch: to switch to another pokemon\n"), addr)
}

func displaySelectedPokemons(pokemonsList []Pokemon, conn *net.UDPConn, addr *net.UDPAddr) {
	conn.WriteToUDP([]byte("You have:\n"), addr)
	for i, pokemon := range pokemonsList {
		conn.WriteToUDP([]byte(fmt.Sprintf("%d. %s\n", i, pokemon.Name)), addr)
	}
	conn.WriteToUDP([]byte("Please enter the index of the pokemon you want to switch to:\n"), addr)
}

func switchToChosenPokemon(pokemonsList []Pokemon, conn *net.UDPConn, addr *net.UDPAddr) *Pokemon {
	for {
		index := readIndex(conn, addr)
		if index < 0 || index >= len(pokemonsList) {
			conn.WriteToUDP([]byte("Please enter a valid index.\n"), addr)
			continue
		}
		if isAlive(&pokemonsList[index]) {
			return &pokemonsList[index]
		} else {
			conn.WriteToUDP([]byte("This pokemon is dead. Please select another one.\n"), addr)
		}
	}
}

func readCommands(conn *net.UDPConn, addr *net.UDPAddr) string {
	buffer := make([]byte, 1024)
	n, _, _ := conn.ReadFromUDP(buffer)
	command := strings.TrimSpace(string(buffer[:n]))
	if command == "attack" || command == "switch" || command == "?" || command == "surrender" {
		return strings.ToLower(command)
	}
	conn.WriteToUDP([]byte("Please enter a valid command\n"), addr)
	return readCommands(conn, addr)
}

func readIndex(conn *net.UDPConn, addr *net.UDPAddr) int {
	buffer := make([]byte, 1024)
	n, _, _ := conn.ReadFromUDP(buffer)
	input := strings.TrimSpace(string(buffer[:n]))
	index, _ := strconv.Atoi(input)
	return index
}
func printPokemonInfo(index int, pokemon Pokemon) {
	fmt.Println(index, ":", pokemon.Name)

	fmt.Println("\tElements: ")
	for _, element := range pokemon.Elements {
		fmt.Println("\t\tElement:", element)
	}

	fmt.Println("\tStats:")
	fmt.Println("\t\tHP:", pokemon.Stats.HP)
	fmt.Println("\t\tAttack:", pokemon.Stats.Attack)
	fmt.Println("\t\tDefense:", pokemon.Stats.Defense)
	fmt.Println("\t\tSpeed:", pokemon.Stats.Speed)
	fmt.Println("\t\tSp_Attack:", pokemon.Stats.Sp_Attack)
	fmt.Println("\t\tSp_Defense:", pokemon.Stats.Sp_Defense)

	fmt.Println("\tDamage When Attacked:")
	for _, element := range pokemon.DamegeWhenAttacked {
		fmt.Printf("\t\tElement: %s. Coefficient: %f\n", element.Element, element.Coefficient)
	}
}

func selectPokemon(player *Player, conn *net.UDPConn, addr *net.UDPAddr) *[]Pokemon {
	var selectedPokemons = []Pokemon{}
	counter := 1
	for {
		if len(selectedPokemons) == 3 {
			break
		}
		conn.WriteToUDP([]byte(fmt.Sprintf("Enter the index of the %d pokemon you want to select: ", counter)), addr)
		index := readIndex(conn, addr)
		if index < 0 || index >= len(player.Inventory) {
			conn.WriteToUDP([]byte("Invalid index\n"), addr)
			continue
		}

		if isContain(selectedPokemons, player.Inventory[index]) {
			conn.WriteToUDP([]byte("You have selected this pokemon. Please select another one.\n"), addr)
			continue
		}

		conn.WriteToUDP([]byte(fmt.Sprintf("Selected %s\n", player.Inventory[index].Name)), addr)
		counter++
		selectedPokemons = append(selectedPokemons, player.Inventory[index])
	}

	conn.WriteToUDP([]byte("You have selected: "), addr)
	for _, pokemon := range selectedPokemons {
		conn.WriteToUDP([]byte(fmt.Sprintf("%s ", pokemon.Name)), addr)
	}
	conn.WriteToUDP([]byte("\n"), addr)

	return &selectedPokemons
}

// NOT DONE
func addExp(winner *Player, loser *Player, conn *net.UDPConn) {}
