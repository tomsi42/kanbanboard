// Command kanban-mcp is an MCP server that exposes kanbanboard tasks to AI agents
// over the Model Context Protocol (stdio transport, JSON-RPC 2.0).
//
// Configuration via environment variables:
//
//	KANBAN_API_URL   — base URL of the kanbanboard API (default: http://localhost:8080)
//	KANBAN_API_TOKEN — Bearer token for authentication (required)
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// ─── JSON-RPC 2.0 types ────────────────────────────────────────────────────

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ─── MCP types ─────────────────────────────────────────────────────────────

type toolDef struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	InputSchema jsonSchema `json:"inputSchema"`
}

type jsonSchema struct {
	Type       string              `json:"type"`
	Properties map[string]schemaProp `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

type schemaProp struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type toolResult struct {
	Content []contentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ─── API client ────────────────────────────────────────────────────────────

type apiClient struct {
	baseURL string
	token   string
	http    *http.Client
}

func (c *apiClient) do(method, path string, body any) ([]byte, int, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+"/api/v1"+path, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	return data, resp.StatusCode, err
}

func (c *apiClient) get(path string) ([]byte, int, error) {
	return c.do("GET", path, nil)
}

func (c *apiClient) post(path string, body any) ([]byte, int, error) {
	return c.do("POST", path, body)
}

// ─── Tool definitions ──────────────────────────────────────────────────────

var tools = []toolDef{
	{
		Name:        "list_my_tasks",
		Description: "List all tasks currently assigned to you across all projects. Returns task IDs, references, titles, priorities, current columns, and project context.",
		InputSchema: jsonSchema{Type: "object"},
	},
	{
		Name:        "get_task",
		Description: "Get full details of a task by its reference (e.g. KB-7) or UUID. Returns title, description, priority, column, assignee, blockers, and project context.",
		InputSchema: jsonSchema{
			Type: "object",
			Properties: map[string]schemaProp{
				"task_ref": {Type: "string", Description: "Task reference in TAG-N format (e.g. KB-7). Use this OR task_id."},
				"task_id":  {Type: "string", Description: "Task UUID. Use this OR task_ref."},
				"project_id": {Type: "string", Description: "Project UUID — required when using task_id."},
			},
		},
	},
	{
		Name:        "create_task",
		Description: "Create a new task in a project. Optionally specify a column; defaults to the first column (Backlog). Optionally set a parent task to create a subtask.",
		InputSchema: jsonSchema{
			Type: "object",
			Properties: map[string]schemaProp{
				"project_id":     {Type: "string", Description: "Project UUID (required)."},
				"title":          {Type: "string", Description: "Task title (required)."},
				"column_id":      {Type: "string", Description: "Column UUID. Defaults to first column if omitted."},
				"parent_task_id": {Type: "string", Description: "Parent task UUID for creating a subtask or sub-subtask."},
				"description":    {Type: "string", Description: "Optional task description."},
				"priority":       {Type: "string", Description: "Priority: none, low, medium, high, or critical."},
			},
			Required: []string{"project_id", "title"},
		},
	},
	{
		Name:        "handoff_task",
		Description: "Atomically move a task to a new column and optionally reassign it to another agent. Used to pass work through the agent pipeline (e.g. coder → tester → reviewer).",
		InputSchema: jsonSchema{
			Type: "object",
			Properties: map[string]schemaProp{
				"task_id":     {Type: "string", Description: "Task UUID (required)."},
				"project_id":  {Type: "string", Description: "Project UUID (required)."},
				"column_id":   {Type: "string", Description: "Target column UUID (required)."},
				"assignee_id": {Type: "string", Description: "New assignee UUID. Pass empty string to unassign. Omit to leave unchanged."},
			},
			Required: []string{"task_id", "project_id", "column_id"},
		},
	},
	{
		Name:        "add_comment",
		Description: "Add a comment to a task. Use this to document work done, note blockers, or communicate status before handing off.",
		InputSchema: jsonSchema{
			Type: "object",
			Properties: map[string]schemaProp{
				"task_id":    {Type: "string", Description: "Task UUID (required)."},
				"project_id": {Type: "string", Description: "Project UUID (required)."},
				"text":       {Type: "string", Description: "Comment text (required)."},
			},
			Required: []string{"task_id", "project_id", "text"},
		},
	},
}

// ─── Tool execution ────────────────────────────────────────────────────────

func (c *apiClient) callTool(name string, rawArgs json.RawMessage) toolResult {
	var args map[string]string
	if len(rawArgs) > 0 {
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return errResult("invalid arguments: " + err.Error())
		}
	}
	if args == nil {
		args = map[string]string{}
	}

	switch name {
	case "list_my_tasks":
		return c.listMyTasks()
	case "get_task":
		return c.getTask(args)
	case "create_task":
		return c.createTask(args)
	case "handoff_task":
		return c.handoffTask(args)
	case "add_comment":
		return c.addComment(args)
	default:
		return errResult("unknown tool: " + name)
	}
}

func (c *apiClient) listMyTasks() toolResult {
	data, status, err := c.get("/tasks/mine")
	if err != nil {
		return errResult("API request failed: " + err.Error())
	}
	if status != http.StatusOK {
		return errResult(fmt.Sprintf("API error %d: %s", status, string(data)))
	}
	return textResult(string(data))
}

func (c *apiClient) getTask(args map[string]string) toolResult {
	ref := args["task_ref"]
	taskID := args["task_id"]
	projectID := args["project_id"]

	if ref != "" {
		data, status, err := c.get("/tasks/by-ref/" + ref)
		if err != nil {
			return errResult("API request failed: " + err.Error())
		}
		if status == http.StatusNotFound {
			return errResult("task not found: " + ref)
		}
		if status != http.StatusOK {
			return errResult(fmt.Sprintf("API error %d: %s", status, string(data)))
		}
		return textResult(string(data))
	}

	if taskID == "" {
		return errResult("either task_ref or task_id is required")
	}
	if projectID == "" {
		return errResult("project_id is required when using task_id")
	}

	data, status, err := c.get("/projects/" + projectID + "/tasks/" + taskID)
	if err != nil {
		return errResult("API request failed: " + err.Error())
	}
	if status == http.StatusNotFound {
		return errResult("task not found: " + taskID)
	}
	if status != http.StatusOK {
		return errResult(fmt.Sprintf("API error %d: %s", status, string(data)))
	}
	return textResult(string(data))
}

func (c *apiClient) createTask(args map[string]string) toolResult {
	projectID := args["project_id"]
	if projectID == "" {
		return errResult("project_id is required")
	}
	title := args["title"]
	if title == "" {
		return errResult("title is required")
	}

	body := map[string]any{"title": title}
	if v := args["column_id"]; v != "" {
		body["columnId"] = v
	}
	if v := args["parent_task_id"]; v != "" {
		body["parentTaskId"] = v
	}

	// Create the task first
	data, status, err := c.post("/projects/"+projectID+"/tasks", body)
	if err != nil {
		return errResult("API request failed: " + err.Error())
	}
	if status != http.StatusCreated {
		return errResult(fmt.Sprintf("API error %d: %s", status, string(data)))
	}

	// Apply optional fields via update
	if desc := args["description"]; desc != "" || args["priority"] != "" {
		var task map[string]any
		if err := json.Unmarshal(data, &task); err == nil {
			if taskID, ok := task["id"].(string); ok {
				update := map[string]any{}
				if desc != "" {
					update["description"] = desc
				}
				if p := args["priority"]; p != "" {
					update["priority"] = p
				}
				c.do("PUT", "/projects/"+projectID+"/tasks/"+taskID, update)
				// Re-fetch updated task
				updated, s2, _ := c.get("/projects/" + projectID + "/tasks/" + taskID)
				if s2 == http.StatusOK {
					data = updated
				}
			}
		}
	}

	return textResult(string(data))
}

func (c *apiClient) handoffTask(args map[string]string) toolResult {
	taskID := args["task_id"]
	projectID := args["project_id"]
	columnID := args["column_id"]

	if taskID == "" || projectID == "" || columnID == "" {
		return errResult("task_id, project_id, and column_id are required")
	}

	body := map[string]any{"columnId": columnID}
	if assigneeID, ok := args["assignee_id"]; ok {
		// present in args — may be "" (unassign) or a UUID
		body["assigneeId"] = &assigneeID
	}

	data, status, err := c.post("/projects/"+projectID+"/tasks/"+taskID+"/handoff", body)
	if err != nil {
		return errResult("API request failed: " + err.Error())
	}
	if status != http.StatusOK {
		return errResult(fmt.Sprintf("API error %d: %s", status, string(data)))
	}
	return textResult(string(data))
}

func (c *apiClient) addComment(args map[string]string) toolResult {
	taskID := args["task_id"]
	projectID := args["project_id"]
	text := args["text"]

	if taskID == "" || projectID == "" || text == "" {
		return errResult("task_id, project_id, and text are required")
	}

	data, status, err := c.post(
		"/projects/"+projectID+"/tasks/"+taskID+"/comments",
		map[string]string{"text": text},
	)
	if err != nil {
		return errResult("API request failed: " + err.Error())
	}
	if status != http.StatusCreated {
		return errResult(fmt.Sprintf("API error %d: %s", status, string(data)))
	}
	return textResult(string(data))
}

// ─── Helpers ───────────────────────────────────────────────────────────────

func textResult(text string) toolResult {
	return toolResult{Content: []contentBlock{{Type: "text", Text: text}}}
}

func errResult(msg string) toolResult {
	return toolResult{
		Content: []contentBlock{{Type: "text", Text: "Error: " + msg}},
		IsError: true,
	}
}

func respond(enc *json.Encoder, id json.RawMessage, result any) {
	enc.Encode(rpcResponse{JSONRPC: "2.0", ID: id, Result: result})
}

func respondErr(enc *json.Encoder, id json.RawMessage, code int, msg string) {
	enc.Encode(rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: msg}})
}

// ─── Main loop ─────────────────────────────────────────────────────────────

func main() {
	apiURL := os.Getenv("KANBAN_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}
	token := os.Getenv("KANBAN_API_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "KANBAN_API_TOKEN is required")
		os.Exit(1)
	}

	client := &apiClient{
		baseURL: strings.TrimRight(apiURL, "/"),
		token:   token,
		http:    &http.Client{},
	}

	log.SetOutput(os.Stderr) // MCP uses stdout for protocol; logs go to stderr
	log.Printf("kanban-mcp starting, API: %s", client.baseURL)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 4*1024*1024), 4*1024*1024)
	enc := json.NewEncoder(os.Stdout)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req rpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			respondErr(enc, nil, -32700, "parse error")
			continue
		}

		// Notifications have no ID — handle but don't respond
		if req.ID == nil {
			continue
		}

		switch req.Method {
		case "initialize":
			respond(enc, req.ID, map[string]any{
				"protocolVersion": "2024-11-05",
				"capabilities":    map[string]any{"tools": map[string]any{}},
				"serverInfo":      map[string]any{"name": "kanbanboard", "version": "1.3"},
			})

		case "tools/list":
			respond(enc, req.ID, map[string]any{"tools": tools})

		case "tools/call":
			var params struct {
				Name      string          `json:"name"`
				Arguments json.RawMessage `json:"arguments"`
			}
			if err := json.Unmarshal(req.Params, &params); err != nil {
				respondErr(enc, req.ID, -32602, "invalid params")
				continue
			}
			result := client.callTool(params.Name, params.Arguments)
			respond(enc, req.ID, result)

		case "ping":
			respond(enc, req.ID, map[string]any{})

		default:
			respondErr(enc, req.ID, -32601, "method not found: "+req.Method)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("stdin error: %v", err)
	}
}
