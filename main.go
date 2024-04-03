package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type cliCommand struct {
	name        string
	description string
	callback    func(args []string)
}

type Pokemon struct {
	Name           string        `json:"name"`
	Height         int           `json:"height"`
	Weight         int           `json:"weight"`
	Stats          []PokemonStat `json:"stats"`
	Types          []PokemonType `json:"types"`
	BaseExperience int           `json:"base_experience"`
}

type PokemonStat struct {
	BaseStat int `json:"base_stat"`
	Effort   int `json:"effort"`
	Stat     struct {
		Name string `json:"name"`
	} `json:"stat"`
}

type PokemonType struct {
	Slot int `json:"slot"`
	Type struct {
		Name string `json:"name"`
	} `json:"type"`
}

var catchedPokemon = make(map[string]Pokemon)
var definedCommands = make(map[string]cliCommand)

var currentMapIndex = 0

func main() {
	addCommands(cliCommand{name: "help", description: "Displays a help message", callback: help})
	addCommands(cliCommand{name: "exit", description: "Exits the pokedex", callback: exit})
	addCommands(cliCommand{name: "map", description: "Displays 20 locations at a time", callback: mapF})
	addCommands(cliCommand{name: "mapb", description: "Displays 20 locations back", callback: mapB})
	addCommands(cliCommand{name: "explore", description: "shows pokemon available in a location", callback: explore})
	addCommands(cliCommand{name: "catch", description: "catch pokemon", callback: catch})
	addCommands(cliCommand{name: "inspect", description: "inspect pokemon", callback: inspect})
	addCommands(cliCommand{name: "pokedex", description: "show your pokedex", callback: pokedex})

	for {
		fmt.Print("Pokedex > ")
		reader := bufio.NewReader(os.Stdin)
		commandInput, _ := reader.ReadString('\n')
		commandInput = strings.TrimSpace(commandInput) // Remove newline character at the end

		// Split the input into command and arguments
		inputParts := strings.Fields(commandInput) // Fields splits the string by spaces, handling multiple spaces correctly
		if len(inputParts) == 0 {
			continue // Skip if no input
		}

		command := inputParts[0]
		args := inputParts[1:] // All other parts are considered arguments

		v, ok := definedCommands[command]

		if ok {
			v.callback(args)
		} else {
			fmt.Println("Unknown command")
		}
	}
}

func pokedex(args []string) {
	fmt.Println("Your Pokedex:")
	keys := make([]string, 0, len(catchedPokemon))
	for k := range catchedPokemon {
		keys = append(keys, k)
	}

	for x := range keys {
		fmt.Println("  -", keys[x])
	}

}

func inspect(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Please specify a pokemon: inspect pokemon-name")
		return
	}

	v, ok := catchedPokemon[args[0]]

	if !ok {
		fmt.Println("You have not caught that pokemon")
	} else {
		fmt.Println("Name:", v.Name)
		fmt.Println("Height:", v.Height)
		fmt.Println("Weight:", v.Weight)
		fmt.Println("Stats:")
		for x := range v.Stats {
			fmt.Printf("  -%v: %v\n", v.Stats[x].Stat.Name, v.Stats[x].BaseStat)

		}

		fmt.Println("Types:")

		for f := range v.Types {
			fmt.Printf("  - %v\n", v.Types[f].Type.Name)

		}
	}
}

func catch(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Please specify a pokemon: explore pokemon-name")
		return
	}

	fmt.Printf("Throwing a Pokeball at %v...\n", args[0])

	res := GetPokeApi("/api/v2/pokemon/" + args[0])

	pok := Pokemon{}

	err := json.Unmarshal(res, &pok)
	if err != nil {
		log.Fatal(err)
	}

	rand.Seed(time.Now().UnixNano())

	randomNumber := generateRandomNumber()

	if randomNumber > pok.BaseExperience {
		catchedPokemon[pok.Name] = pok
		fmt.Printf("%v was caught!\n", pok.Name)
	} else {
		fmt.Printf("%v escaped!\n", pok.Name)
	}

}

func explore(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Please specify a location: explore location-name")
		return
	}

	fmt.Printf("Exploring %v...", args[0])

	fmt.Println("Found Pokemon:")

	pok := struct {
		PokemonEncounters []struct {
			Pokemon struct {
				Name string `json:"name"`
			} `json:"pokemon"`
		} `json:"pokemon_encounters"`
	}{}

	res := GetPokeApi("/api/v2/location-area/" + args[0])

	err := json.Unmarshal(res, &pok)

	if err != nil {
		log.Fatal(err)
	}
	for _, encounter := range pok.PokemonEncounters {
		fmt.Printf("- %v\n", encounter.Pokemon.Name)
	}

}

func mapF(args []string) {
	res := GetPokeApi("/api/v2/location-area/?offset=" + strconv.Itoa(currentMapIndex) + "&limit=20")
	loc := struct {
		Count    int    `json:"count"`
		Next     string `json:"next"`
		Previous string `json:"previous"`
		Results  []struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"results"`
	}{}
	err := json.Unmarshal(res, &loc)

	if err != nil {

		log.Fatal(err)
	}
	for location := range loc.Results {
		fmt.Println(loc.Results[location].Name)
	}

	currentMapIndex += 20
}

func mapB(args []string) {
	if currentMapIndex == 0 {
		fmt.Println("Already at starting")
		return
	}
	currentMapIndex -= 20

	res := GetPokeApi("/api/v2/location-area/?offset=" + strconv.Itoa(currentMapIndex) + "&limit=20")
	loc := struct {
		Count    int    `json:"count"`
		Next     string `json:"next"`
		Previous string `json:"previous"`
		Results  []struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"results"`
	}{}
	err := json.Unmarshal(res, &loc)

	if err != nil {

		log.Fatal(err)
	}
	for location := range loc.Results {
		fmt.Println(loc.Results[location].Name)
	}

}

func exit(args []string) {
	os.Exit(0)
}

func help(args []string) {
	fmt.Println("Welcome to the pokedex!")
	fmt.Println("Usage:")
	fmt.Println("")

	for command := range definedCommands {
		fmt.Printf("%v: %v\n", definedCommands[command].name, definedCommands[command].description)
	}
	fmt.Println("")

}

func addCommands(command cliCommand) {
	_, ok := definedCommands[command.name]

	if ok {
		return
	}

	definedCommands[command.name] = command
}

func generateRandomNumber() int {
	// Decide the length: 1 digit (max 9), 2 digits (max 99), or 3 digits (max 999)
	lengthOptions := []int{9, 99, 999}

	// Choose a random index to decide the max value for length
	index := rand.Intn(len(lengthOptions))
	maxValue := lengthOptions[index]

	// Generate a random number in the range [1, maxValue]
	return rand.Intn(maxValue) + 1
}
