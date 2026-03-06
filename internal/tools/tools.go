package tools

import (
	"fmt"
	"math/rand"
)

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type MutationType int

const (
	MutationAdd MutationType = iota
	MutationRemove
	MutationUpdate
)

type MutationEvent struct {
	Type     MutationType
	ToolName string
}

var toolVerbs = []string{
	"get", "create", "update", "delete", "list",
	"search", "analyze", "generate", "validate", "export",
	"import", "transform", "sync", "fetch", "deploy",
	"publish", "archive", "restore", "calculate", "aggregate",
}

var toolNouns = []string{
	"users", "reports", "invoices", "orders", "products",
	"metrics", "logs", "events", "configs", "pipelines",
	"documents", "images", "messages", "tickets", "workflows",
	"schemas", "databases", "clusters", "backups", "alerts",
}

var paramNames = []string{
	"id", "name", "query", "limit", "offset",
	"filter", "format", "start_date", "end_date", "status",
	"type", "priority", "tags", "sort_by", "include_deleted",
}

var paramTypes = []string{"string", "integer", "boolean", "number"}

func GenerateRandom(minCount, maxCount int) []Tool {
	count := minCount + rand.Intn(maxCount-minCount+1)
	result := make([]Tool, 0, count)
	used := make(map[string]bool)

	for len(result) < count {
		name := randomToolName()
		if used[name] {
			continue
		}
		used[name] = true
		result = append(result, generateTool(name))
	}
	return result
}

func Mutate(current []Tool) ([]Tool, MutationEvent) {
	if len(current) == 0 {
		tool := generateTool(randomToolName())
		return []Tool{tool}, MutationEvent{Type: MutationAdd, ToolName: tool.Name}
	}

	action := rand.Intn(3)

	switch action {
	case 0: // add
		name := randomToolName()
		for _, t := range current {
			if t.Name == name {
				name = fmt.Sprintf("%s_v%d", name, rand.Intn(100))
				break
			}
		}
		tool := generateTool(name)
		return append(current, tool), MutationEvent{Type: MutationAdd, ToolName: name}

	case 1: // remove
		if len(current) <= 1 {
			return Mutate(current)
		}
		idx := rand.Intn(len(current))
		removed := current[idx].Name
		current = append(current[:idx], current[idx+1:]...)
		return current, MutationEvent{Type: MutationRemove, ToolName: removed}

	default: // update
		idx := rand.Intn(len(current))
		current[idx] = generateTool(current[idx].Name)
		return current, MutationEvent{Type: MutationUpdate, ToolName: current[idx].Name}
	}
}

func randomToolName() string {
	verb := toolVerbs[rand.Intn(len(toolVerbs))]
	noun := toolNouns[rand.Intn(len(toolNouns))]
	return fmt.Sprintf("%s_%s", verb, noun)
}

func generateTool(name string) Tool {
	numParams := 1 + rand.Intn(4)
	properties := make(map[string]interface{})
	required := make([]string, 0)

	used := make(map[string]bool)
	for i := 0; i < numParams; i++ {
		pName := paramNames[rand.Intn(len(paramNames))]
		if used[pName] {
			continue
		}
		used[pName] = true
		properties[pName] = map[string]interface{}{
			"type":        paramTypes[rand.Intn(len(paramTypes))],
			"description": fmt.Sprintf("The %s parameter", pName),
		}
		if rand.Float32() < 0.5 {
			required = append(required, pName)
		}
	}

	return Tool{
		Name:        name,
		Description: fmt.Sprintf("Tool that performs %s operation", name),
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": properties,
			"required":   required,
		},
	}
}
