package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
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

const (
	//numberOfPokemons = 649
	baseURL = "https://pokedex.org/#/"
)

var pokemons []Pokemon

func main() {
	crawlPokemonsDriver(40)
}
func crawlPokemonsDriver(numsOfPokemons int) {
	//var wg sync.WaitGroup
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch()
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}

	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}

	page.Goto(baseURL)

	for i := range numsOfPokemons {

		locator := fmt.Sprintf("button.sprite-%d", i+1)
		button := page.Locator(locator).First()
		time.Sleep(500 * time.Millisecond)

		button.Click()
		time.Sleep(500 * time.Millisecond)

		pokemons = append(pokemons, crawlPokemons(page))

		page.Goto(baseURL)
		page.Reload()
	}
	// wg.Add(1)

	// go func(page playwright.Page) {
	// defer wg.Done()
	// pokemon := Pokemon{}
	// pokemon.Name = crawlName(page)
	// pokemon.Id = crawlId(page)
	// pokemon.Stats = crawlStats(page)
	// pokemon.DamegeWhenAttacked = crawlDamageWhenAttacked(page)
	// pokemon.Elements = crawlElements(page)
	// pokemon.Profile = crawlProfile(page)
	// pokemon.EvolutionLevel, pokemon.NextEvolution = crawlEvo(page, pokemon.Name)
	// pokemon.Moves = createMoves(page, pokemon.Elements)
	// pokemons = append(pokemons, pokemon)
	// page.Goto(baseURL)
	// page.Reload()
	// }(page)
	// wg.Wait()
	js, err := json.MarshalIndent(pokemons, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	os.WriteFile("./lib/pokedex.json", js, 0644)

	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
}

func crawlPokemons(page playwright.Page) Pokemon {
	pokemon := Pokemon{}

	pokemon.Name = crawlName(page)
	pokemon.Id = crawlId(page)
	pokemon.Stats = crawlStats(page)
	pokemon.DamegeWhenAttacked = crawlDamageWhenAttacked(page)
	pokemon.Elements = crawlElements(page)
	pokemon.Profile = crawlProfile(page)
	pokemon.EvolutionLevel, pokemon.NextEvolution = crawlEvo(page, pokemon.Name)
	pokemon.Moves = createMoves(page, pokemon.Elements)
	// pokemon.Exp = 0
	// pokemon.Level = 0
	fmt.Println(pokemon)
	return pokemon
}
func crawlName(page playwright.Page) string {
	name, _ := page.Locator("div.detail-panel > h1.detail-panel-header").TextContent()
	return name
}
func crawlElements(page playwright.Page) []string {
	elements := make([]string, 2)
	i := 0
	entries, _ := page.Locator("div.detail-types > span.monster-type").All()
	for _, entry := range entries {
		element, _ := entry.TextContent()
		//pokemon.Elements = append(pokemon.Elements, element)
		elements[i] = element
		i++
	}
	return elements[0:i]
}
func crawlId(page playwright.Page) int {
	entry, _ := page.Locator("div.detail-national-id > span").TextContent()
	id, _ := strconv.Atoi(strings.SplitAfter(entry, "#")[1])
	return (id)
}
func crawlStats(page playwright.Page) Stats {
	stats := Stats{}
	entries, _ := page.Locator("div.detail-panel-content > div.detail-header > div.detail-infobox > div.detail-stats > div.detail-stats-row").All()
	for _, entry := range entries {
		title, _ := entry.Locator("span:not([class])").TextContent()
		switch title {
		case "HP":
			hp, _ := entry.Locator("span.stat-bar > div.stat-bar-fg").TextContent()
			hp_int, _ := strconv.Atoi(hp)
			stats.HP = float32(hp_int)
		case "Attack":
			attack, _ := entry.Locator("span.stat-bar > div.stat-bar-fg").TextContent()
			attack_int, _ := strconv.Atoi(attack)
			stats.Attack = float32(attack_int)
		case "Defense":
			defense, _ := entry.Locator("span.stat-bar > div.stat-bar-fg").TextContent()
			defensep_int, _ := strconv.Atoi(defense)
			stats.Defense = float32(defensep_int)
		case "Speed":
			speed, _ := entry.Locator("span.stat-bar > div.stat-bar-fg").TextContent()
			stats.Speed, _ = strconv.Atoi(speed)

		case "Sp Atk":
			sp_Attack, _ := entry.Locator("span.stat-bar > div.stat-bar-fg").TextContent()
			sp_Attack_int, _ := strconv.Atoi(sp_Attack)
			stats.Sp_Attack = float32(sp_Attack_int)
		case "Sp Def":
			sp_Defense, _ := entry.Locator("span.stat-bar > div.stat-bar-fg").TextContent()
			sp_Defense_int, _ := strconv.Atoi(sp_Defense)
			stats.Sp_Defense = float32(sp_Defense_int)
		default:
			fmt.Println("Unknown title: ", title)
		}
	}
	return stats
}
func crawlDamageWhenAttacked(page playwright.Page) []DamegeWhenAttacked {

	damegeWhenAttacked := []DamegeWhenAttacked{}
	entries, _ := page.Locator("div.when-attacked > div.when-attacked-row").All()
	for _, entry := range entries {
		//first column
		element1, _ := entry.Locator("span.monster-type:nth-child(1)").TextContent()
		coefficient1, _ := entry.Locator("span.monster-multiplier:nth-child(2)").TextContent()
		coefficients1 := strings.Split(coefficient1, "x")
		coef1, _ := strconv.ParseFloat(coefficients1[0], 32)

		//second column
		element2, _ := entry.Locator("span.monster-type:nth-child(3)").TextContent()
		coefficient2, _ := entry.Locator("span.monster-multiplier:nth-child(4)").TextContent()
		coefficients2 := strings.Split(coefficient2, "x")
		coef2, _ := strconv.ParseFloat(coefficients2[0], 32)

		//append row
		if element1 != "" {
			damegeWhenAttacked = append(damegeWhenAttacked, DamegeWhenAttacked{Element: element1, Coefficient: float32(coef1)})
		}
		if element2 != "" {
			damegeWhenAttacked = append(damegeWhenAttacked, DamegeWhenAttacked{Element: element2, Coefficient: float32(coef2)})
		}
	}
	return damegeWhenAttacked
}

func crawlProfile(page playwright.Page) Profile {
	profile := Profile{}
	genderRatio := GenderRatio{}
	entries, _ := page.Locator("div.detail-panel-content > div.detail-below-header > div.monster-minutia").All()
	for _, entry := range entries {
		title1, _ := entry.Locator("strong:not([class]):nth-child(1)").TextContent()
		stat1, _ := entry.Locator("span:not([class]):nth-child(2)").TextContent()
		switch title1 {
		case "Height:":
			heights := strings.Split(stat1, " ")
			height, _ := strconv.ParseFloat(heights[0], 32)
			profile.Height = float32(height)
		case "Catch Rate:":
			catchRates := strings.Split(stat1, "%")
			catchRate, _ := strconv.ParseFloat(catchRates[0], 32)
			profile.CatchRate = float32(catchRate)
		case "Egg Groups:":
			profile.EggGroup = stat1
		case "Abilities:":
			profile.Abilities = stat1
		}

		title2, _ := entry.Locator("strong:not([class]):nth-child(3)").TextContent()
		stat2, _ := entry.Locator("span:not([class]):nth-child(4)").TextContent()
		switch title2 {
		case "Weight:":
			weights := strings.Split(stat2, " ")
			weight, _ := strconv.ParseFloat(weights[0], 32)
			profile.Weight = float32(weight)
		case "Gender Ratio:":
			if stat2 == "N/A" {
				genderRatio.MaleRatio = 0
				genderRatio.FemaleRatio = 0
			} else {
				ratios := strings.Split(stat2, " ")

				maleRatios := strings.Split(ratios[0], "%")
				maleRatio, _ := strconv.ParseFloat(maleRatios[0], 32)
				genderRatio.MaleRatio = float32(maleRatio)

				femaleRatios := strings.Split(ratios[2], "%")
				femaleRatio, _ := strconv.ParseFloat(femaleRatios[0], 32)
				genderRatio.FemaleRatio = float32(femaleRatio)
			}
			profile.GenderRatio = genderRatio
		case "Hatch Steps:":
			profile.HatchSteps, _ = strconv.Atoi(stat2)
		}
	}
	return profile
}

func crawlEvo(page playwright.Page, name string) (int, string) {
	entries, _ := page.Locator("div.evolutions > div.evolution-row").All()
	var evolutionLevel int
	var nextEvolution string
	for _, entry := range entries {
		evolutionLabel, _ := entry.Locator("div.evolution-label > span").TextContent()
		evolutionLabels := strings.Split(evolutionLabel, " ")

		if evolutionLabels[0] == name {
			evolutionLevels := strings.Split(evolutionLabels[len(evolutionLabels)-1], ".")
			evolutionLevel, _ = strconv.Atoi(evolutionLevels[0])

			nextEvolution = evolutionLabels[3]
		}
	}
	return evolutionLevel, nextEvolution
}
func createMoves(page playwright.Page, elements []string) []Moves {
	moves := []Moves{}
	entries, _ := page.Locator("div.monster-moves > div.moves-row").All()
	i := 0
	for _, entry := range entries {
		if i == 2 {
			break
		}
		// simulate clicking the expand button in the move rows
		expandButton := page.Locator("div.moves-inner-row > button.dropdown-button").First()
		expandButton.Click()

		var name string
		element := make([]string, 2)
		if i == 0 {
			name = "Normal Attack"
			element[0] = "None"
			element = element[0:1]
		} else {
			name = "Special Attack"
			element = elements
		}

		acc, _ := entry.Locator("div.moves-row-detail > div.moves-row-stats > span:nth-child(2)").TextContent()
		accs := strings.Split(acc, ": ")
		accValue := strings.Split(accs[1], "%")
		accInt, _ := strconv.Atoi(accValue[0])
		if accInt == 0 {
			accInt = 80
		}
		moves = append(moves, Moves{Name: name, Element: element, Acc: accInt})
		i++
	}

	return moves
}
