package gamectl

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultAddr = "http://127.0.0.1:8080"

func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("gamectl", flag.ContinueOnError)
	flags.SetOutput(stderr)
	addr := flags.String("addr", defaultAddr, "control plane base URL")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	rest := flags.Args()
	if len(rest) == 0 {
		printUsage(stderr)
		return 2
	}

	client := &Client{
		BaseURL: strings.TrimRight(*addr, "/"),
		HTTP:    &http.Client{Timeout: 5 * time.Second},
	}

	var err error
	switch rest[0] {
	case "status":
		err = client.get(stdout, "/control/v1/status")
	case "service":
		err = runService(client, stdout, rest[1:])
	case "config":
		err = runConfig(client, stdout, stderr, rest[1:])
	case "abtest":
		err = runABTest(client, stdout, stderr, rest[1:])
	case "rollout":
		err = runRollout(client, stdout, stderr, rest[1:])
	case "update":
		err = runUpdate(client, stdout, stderr, rest[1:])
	case "nodes":
		err = client.get(stdout, "/control/v1/nodes")
	default:
		printUsage(stderr)
		return 2
	}

	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func (c *Client) get(out io.Writer, path string) error {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return err
	}
	return c.do(out, req)
}

func (c *Client) postJSON(out io.Writer, path string, payload any) error {
	return c.sendJSON(out, http.MethodPost, path, payload)
}

func (c *Client) putJSON(out io.Writer, path string, payload any) error {
	return c.sendJSON(out, http.MethodPut, path, payload)
}

func (c *Client) sendJSON(out io.Writer, method, path string, payload any) error {
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return err
	}
	req, err := http.NewRequest(method, c.BaseURL+path, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(out, req)
}

func (c *Client) do(out io.Writer, req *http.Request) error {
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("%s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	_, err = out.Write(body)
	return err
}

func runService(client *Client, out io.Writer, args []string) error {
	if len(args) == 0 || args[0] == "list" {
		return client.get(out, "/control/v1/services")
	}
	if len(args) != 3 || args[0] != "action" {
		return fmt.Errorf("usage: gamectl service [list|action <service> <start|stop|restart>]")
	}
	return client.postJSON(out, "/control/v1/services/"+args[1]+"/actions", map[string]string{"action": args[2]})
}

func runConfig(client *Client, out, stderr io.Writer, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: gamectl config [list|get|set]")
	}
	switch args[0] {
	case "list":
		return client.get(out, "/control/v1/configs")
	case "get":
		if len(args) != 2 {
			return fmt.Errorf("usage: gamectl config get <key>")
		}
		return client.get(out, "/control/v1/configs/"+args[1])
	case "set":
		key, value, scope, ok := parseConfigSetArgs(args[1:])
		if !ok {
			return fmt.Errorf("usage: gamectl config set <key> <value> [--scope scope]")
		}
		return client.putJSON(out, "/control/v1/configs/"+key, map[string]string{"value": value, "scope": scope})
	default:
		return fmt.Errorf("usage: gamectl config [list|get|set]")
	}
}

func parseConfigSetArgs(args []string) (key, value, scope string, ok bool) {
	scope = "global"
	values := make([]string, 0, 2)
	for i := 0; i < len(args); i++ {
		if args[i] == "--scope" {
			if i+1 >= len(args) {
				return "", "", "", false
			}
			scope = args[i+1]
			i++
			continue
		}
		values = append(values, args[i])
	}
	if len(values) != 2 {
		return "", "", "", false
	}
	return values[0], values[1], scope, true
}

func runABTest(client *Client, out, stderr io.Writer, args []string) error {
	if len(args) == 0 || args[0] == "list" {
		return client.get(out, "/control/v1/ab-tests")
	}
	if args[0] != "create" {
		return fmt.Errorf("usage: gamectl abtest [list|create]")
	}
	flags := flag.NewFlagSet("abtest create", flag.ContinueOnError)
	flags.SetOutput(stderr)
	name := flags.String("name", "", "experiment name")
	feature := flags.String("feature", "", "feature key")
	variants := flags.String("variants", "control,treatment", "comma-separated variants")
	traffic := flags.Int("traffic", 1, "traffic percentage")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	return client.postJSON(out, "/control/v1/ab-tests", map[string]any{
		"name":            *name,
		"feature_key":     *feature,
		"variants":        splitCSV(*variants),
		"traffic_percent": *traffic,
	})
}

func runRollout(client *Client, out, stderr io.Writer, args []string) error {
	if len(args) == 0 || args[0] == "list" {
		return client.get(out, "/control/v1/rollouts")
	}
	if args[0] != "create" {
		return fmt.Errorf("usage: gamectl rollout [list|create]")
	}
	flags := flag.NewFlagSet("rollout create", flag.ContinueOnError)
	flags.SetOutput(stderr)
	feature := flags.String("feature", "", "feature key")
	percent := flags.Int("percent", 1, "target percentage")
	strategy := flags.String("strategy", "user_id_hash", "rollout strategy")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	return client.postJSON(out, "/control/v1/rollouts", map[string]any{
		"feature_key":    *feature,
		"target_percent": *percent,
		"strategy":       *strategy,
	})
}

func runUpdate(client *Client, out, stderr io.Writer, args []string) error {
	if len(args) == 0 || args[0] == "list" {
		return client.get(out, "/control/v1/updates")
	}
	if args[0] != "plan" {
		return fmt.Errorf("usage: gamectl update [list|plan]")
	}
	flags := flag.NewFlagSet("update plan", flag.ContinueOnError)
	flags.SetOutput(stderr)
	service := flags.String("service", "", "service name")
	version := flags.String("version", "", "target version")
	strategy := flags.String("strategy", "rolling", "update strategy")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	return client.postJSON(out, "/control/v1/updates", map[string]string{
		"service":  *service,
		"version":  *version,
		"strategy": *strategy,
	})
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func printUsage(out io.Writer) {
	fmt.Fprintln(out, "usage: gamectl [--addr URL] <status|service|config|abtest|rollout|update|nodes>")
}
