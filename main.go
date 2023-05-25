package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	spec "github.com/ZacheryJohnson/spacetraders-cli-go"

	"github.com/urfave/cli/v2" // imports as package "cli"
)

func main() {
	app := &cli.App{
		Name:  "traders",
		Usage: "",
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "initializes a new game account",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "symbol",
						Usage:    "Name for the new account. Must be unique",
						Required: true,
					},
				},
				Action: func(ctx *cli.Context) error {
					return initialize(ctx.String("symbol"))
				},
			},
			{
				Name:  "activate",
				Usage: "activates an existing game account",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "symbol",
						Usage:    "Name of the existing account",
						Required: true,
					},
				},
				Action: func(ctx *cli.Context) error {
					return activate(ctx.String("symbol"))
				},
			},
			{
				Name:  "get",
				Usage: "returns information on different resources",
				Subcommands: []*cli.Command{
					{
						Name:  "agent",
						Usage: "gets information on your agent",
						Action: func(ctx *cli.Context) error {
							agent, err := get_agent()
							prettyPrint(agent)
							return err
						},
					},
					{
						Name:  "contracts",
						Usage: "gets information on your contracts",
						Action: func(ctx *cli.Context) error {
							contracts, err := get_contracts()
							prettyPrint(contracts)
							return err
						},
					},
					{
						Name: "headquarters",
						Aliases: []string{
							"hq",
						},
						Usage: "gets information on your headquarters",
						Action: func(ctx *cli.Context) error {
							headquarters, err := get_headquarters() // ZJ-TODO: allow non-current positions as args
							prettyPrint(headquarters)
							return err
						},
					},
					{
						Name:  "system",
						Usage: "gets information on a system",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "symbol",
								Usage:    "Name of the system",
								Required: true,
							},
						},
						Action: func(ctx *cli.Context) error {
							system, err := get_system(ctx.String("symbol"))
							prettyPrint(system)
							return err
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func prettyPrint(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "\t")
	fmt.Printf("%s\n", s)
}

func getApiClient() spec.APIClient {
	apiClientConfig := spec.NewConfiguration()
	token, err := getToken()
	if err != nil {
		panic(err)
	}
	bearerStr := fmt.Sprintf("Bearer %s", token)
	apiClientConfig.AddDefaultHeader("Authorization", bearerStr)
	return *spec.NewAPIClient(apiClientConfig)
}

func activate(symbol string) error {
	config, readErr := loadConfig()
	if readErr != nil {
		return readErr
	}

	config.ActiveSymbol = symbol

	writeErr := writeConfig(config)
	return writeErr
}

func initialize(symbol string) error {
	if isInitialized() {
		fmt.Println("You already have an account initilized. Delete the old account before creating a new one.")
		return nil
	}

	dirpath := getHomeDir()

	if _, statErr := os.Stat(dirpath); os.IsNotExist(statErr) {
		mkdirErr := os.Mkdir(dirpath, 0755)
		if mkdirErr != nil {
			panic(mkdirErr)
		}
	}

	configPath := path.Join(dirpath, getConfigFileName())
	if _, configStatErr := os.Stat(configPath); os.IsNotExist(configStatErr) {
		newConfig := Config{ActiveSymbol: symbol}
		writeConfig(newConfig)
	}

	client := getApiClient().DefaultApi

	req := client.Register(context.TODO()).RegisterRequest(spec.RegisterRequest{
		Faction: "COSMIC", // ZJ-TODO
		Symbol:  symbol,
	})

	resp, _, rpcErr := client.RegisterExecute(req)
	if rpcErr != nil {
		panic(rpcErr)
	}

	acctToken := resp.Data.Token
	tokenFile := fmt.Sprintf("%s.token", symbol)
	tokenPath := path.Join(dirpath, tokenFile) // ZJ-TODO: support multiple different users?
	if err := os.WriteFile(tokenPath, []byte(acctToken), 0644); err != nil {
		panic(err)
	}

	fmt.Println("Successfully initialized account")
	return nil
}

func get_agent() (spec.Agent, error) {
	if !isInitialized() {
		fmt.Println("You do not have an account. Create one with `initialize` first.")
		return *spec.NewAgentWithDefaults(), nil
	}

	client := getApiClient().AgentsApi
	req := client.GetMyAgent(context.TODO())

	resp, _, err := client.GetMyAgentExecute(req)
	if err != nil {
		return spec.Agent{}, err
	}

	agent := resp.Data
	return agent, nil
}

func get_contracts() ([]spec.Contract, error) {
	if !isInitialized() {
		fmt.Println("You do not have an account. Create one with `initialize` first.")
		return []spec.Contract{}, nil
	}

	client := getApiClient().ContractsApi
	req := client.GetContracts(context.TODO())

	resp, _, err := client.GetContractsExecute(req)
	if err != nil {
		return []spec.Contract{}, err
	}

	contracts := resp.Data
	return contracts, nil
}

func get_headquarters() (spec.Waypoint, error) {
	if !isInitialized() {
		fmt.Println("You do not have an account. Create one with `initialize` first.")
		return spec.Waypoint{}, nil
	}

	agent, err := get_agent()
	if err != nil {
		return spec.Waypoint{}, err
	}
	lastIdx := strings.LastIndex(agent.Headquarters, "-")
	systemStr := agent.Headquarters[:lastIdx]

	fmt.Printf("%s %s\n", systemStr, agent.Headquarters)

	client := getApiClient().SystemsApi
	req := client.GetWaypoint(context.TODO(), systemStr, agent.Headquarters)

	resp, _, rpcErr := client.GetWaypointExecute(req)
	if rpcErr != nil {
		return spec.Waypoint{}, rpcErr
	}

	waypoint := resp.Data
	return waypoint, nil
}

func get_system(symbol string) (spec.System, error) {
	if !isInitialized() {
		fmt.Println("You do not have an account. Create one with `initialize` first.")
		return spec.System{}, nil
	}

	client := getApiClient().SystemsApi
	req := client.GetSystem(context.TODO(), symbol)

	resp, _, rpcErr := client.GetSystemExecute(req)
	if rpcErr != nil {
		return spec.System{}, rpcErr
	}

	system := resp.Data
	return system, nil
}
