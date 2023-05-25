package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	spec "github.com/ZacheryJohnson/spacetraders_cli_go"

	"github.com/urfave/cli/v2" // imports as package "cli"
)

func main() {
	app := &cli.App{
		Name:  "traders",
		Usage: "",
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "initializes a new game instance",
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
				Name:  "get",
				Usage: "returns information on different resources",
				Subcommands: []*cli.Command{
					{
						Name:  "agent",
						Usage: "gets information on your agent",
						Action: func(ctx *cli.Context) error {
							_, err := get_agent()
							return err
						},
					},
					{
						Name:  "contracts",
						Usage: "gets information on your contracts",
						Action: func(ctx *cli.Context) error {
							_, err := get_contracts()
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
							_, err := get_headquarters() // ZJ-TODO: allow non-current positions as args
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

func getToken() (string, error) {
	tokenPath := path.Join(getHomeDir(), "user.token")
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func getHomeDir() string {
	// Make a new directory at a known good path for a token + config
	dirname, dirErr := os.UserHomeDir()
	if dirErr != nil {
		panic(dirErr)
	}

	return path.Join(dirname, ".spacetraders")
}

func isInitialized() bool {
	token, err := getToken()
	if err != nil {
		return false
	}

	return len(token) > 0
}

func initialize(symbol string) error {
	if isInitialized() {
		fmt.Println("You already have an account initilized. Delete the old account before creating a new one.")
		return nil
	}

	dirpath := getHomeDir()

	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		mkdirErr := os.Mkdir(dirpath, 0755)
		if mkdirErr != nil {
			panic(mkdirErr)
		}
	}

	req := *spec.NewRegisterRequest("COSMIC", symbol)
	reqJson, err := req.MarshalJSON()
	fmt.Println(string(reqJson))
	if err != nil {
		panic(err)
	}

	resp, err := http.Post("https://api.spacetraders.io/v2/register", "application/json", bytes.NewReader(reqJson))
	if resp.StatusCode != 201 {
		fmt.Println(resp)
		panic(err)
	}

	respBody, _ := ioutil.ReadAll(resp.Body)

	var data map[string]map[string]interface{}
	if err := json.Unmarshal(respBody, &data); err != nil {
		panic(err)
	}

	acctToken := data["data"]["token"].(string)
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

	req, err := http.NewRequest(http.MethodGet, "https://api.spacetraders.io/v2/my/agent", nil)
	if err != nil {
		return *spec.NewAgentWithDefaults(), err
	}

	token, err := getToken()
	if err != nil {
		return *spec.NewAgentWithDefaults(), err
	}

	req.Header = http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode >= 300 {
		fmt.Println(resp)
		return *spec.NewAgentWithDefaults(), err
	}

	respBody, _ := ioutil.ReadAll(resp.Body)
	var agentResp spec.NullableGetMyAgent200Response
	agentResp.UnmarshalJSON(respBody)
	agent := agentResp.Get().Data

	fmt.Println(agent.AccountId)
	fmt.Println(agent.Credits)
	fmt.Println(agent.Headquarters)
	fmt.Println(agent.Symbol)

	return agent, nil
}

func get_contracts() ([]spec.Contract, error) {
	if !isInitialized() {
		fmt.Println("You do not have an account. Create one with `initialize` first.")
		return []spec.Contract{}, nil
	}

	req, err := http.NewRequest(http.MethodGet, "https://api.spacetraders.io/v2/my/contracts", nil)
	if err != nil {
		return []spec.Contract{}, err
	}

	token, err := getToken()
	if err != nil {
		return []spec.Contract{}, err
	}

	req.Header = http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode >= 300 {
		fmt.Println(resp)
		return []spec.Contract{}, err
	}

	respBody, _ := ioutil.ReadAll(resp.Body)
	var contractsResp spec.NullableGetContracts200Response
	contractsResp.UnmarshalJSON(respBody)
	contracts := contractsResp.Get().Data

	fmt.Println("Contracts:")
	for idx, contract := range contracts {
		fmt.Printf("(%d) ---------\n", idx+1)
		fmt.Println("Id:", contract.Id)
		fmt.Println("Accepted:", contract.Accepted)
		fmt.Println("Expiration:", contract.Expiration)
		fmt.Println("Fulfilled:", contract.Fulfilled)
		fmt.Println("Faction Symbol:", contract.FactionSymbol)
		fmt.Println("Terms:", contract.Terms)
		fmt.Println("Type:", contract.Type)
	}

	return contracts, nil
}

func get_headquarters() (spec.Waypoint, error) {
	if !isInitialized() {
		fmt.Println("You do not have an account. Create one with `initialize` first.")
		return *spec.NewWaypointWithDefaults(), nil
	}

	agent, err := get_agent()
	lastIdx := strings.LastIndex(agent.Headquarters, "-")
	systemStr := agent.Headquarters[:lastIdx]
	waypointStr := agent.Headquarters[lastIdx+1:]
	uri := fmt.Sprintf("https://api.spacetraders.io/v2/systems/%s/waypoints/%s-%s", systemStr, systemStr, waypointStr)

	fmt.Println(uri)

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return *spec.NewWaypointWithDefaults(), err
	}

	token, err := getToken()
	if err != nil {
		return *spec.NewWaypointWithDefaults(), err
	}

	req.Header = http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode >= 300 {
		fmt.Println(resp)
		return *spec.NewWaypointWithDefaults(), err
	}

	respBody, _ := ioutil.ReadAll(resp.Body)
	var waypointResp spec.NullableGetWaypoint200Response
	waypointResp.UnmarshalJSON(respBody)
	waypoint := waypointResp.Get().Data
	fmt.Println(waypoint.X)
	fmt.Println(waypoint.Y)
	fmt.Println(waypoint.Type)
	fmt.Println(waypoint.Symbol)
	fmt.Println(waypoint.SystemSymbol)

	// ZJ-TODO: just return one? what do
	return *spec.NewWaypointWithDefaults(), nil
}
