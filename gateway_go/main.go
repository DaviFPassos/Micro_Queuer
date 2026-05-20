package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var queueDir string

const (
	// Limite de tamanho: 5MB para evitar DoS
	maxPayloadSize = 5 * 1024 * 1024
)

func init() {
	// Detecta o diretório correto usando o arquivo de código-fonte
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(filename))
	queueDir = filepath.Join(projectRoot, "shared_queue")

	// Garante que a pasta da fila existe
	if err := os.MkdirAll(queueDir, 0755); err != nil {
		fmt.Printf("[ERROR] Falha ao criar pasta de fila: %v\n", err)
		os.Exit(1)
	}
}

func isValidJSON(data []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(data, &js) == nil
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"GATEWAY_ACTIVE","version":"V3","message":"Use POST /enqueue to submit tasks","endpoints":{"enqueue":"POST /enqueue"}}`))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	timestamp := time.Now().UnixNano()

	// Validação de método HTTP
	if r.Method != http.MethodPost {
		fmt.Printf("[GATEWAY REJECT] %s - Method %s not allowed\n", time.Now().Format("15:04:05"), r.Method)
		http.Error(w, `{"error":"Method not allowed. Use POST."}`, http.StatusMethodNotAllowed)
		return
	}

	// Limita tamanho do payload para evitar DoS
	r.Body = http.MaxBytesReader(w, r.Body, maxPayloadSize)

	// Lê o corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("[GATEWAY ERROR] %s - Body read error: %v\n", time.Now().Format("15:04:05"), err)
		http.Error(w, `{"error":"Payload too large or invalid body"}`, http.StatusRequestEntityTooLarge)
		return
	}

	// Valida tamanho mínimo (evita arquivos vazios inúteis)
	if len(body) == 0 {
		fmt.Printf("[GATEWAY ERROR] %s - Empty payload rejected\n", time.Now().Format("15:04:05"))
		http.Error(w, `{"error":"Payload cannot be empty"}`, http.StatusBadRequest)
		return
	}

	// Valida se é JSON válido
	if !isValidJSON(body) {
		fmt.Printf("[GATEWAY ERROR] %s - Invalid JSON format\n", time.Now().Format("15:04:05"))
		http.Error(w, `{"error":"Payload must be valid JSON"}`, http.StatusBadRequest)
		return
	}

	fmt.Printf("[GATEWAY INGEST] %s - Payload received. Size: %d bytes. Ingesting to queue...\n", time.Now().Format("15:04:05"), len(body))

	// Salva a tarefa na fila
	taskPath := filepath.Join(queueDir, fmt.Sprintf("task_%d.json", timestamp))
	err = os.WriteFile(taskPath, body, 0644)
	if err != nil {
		fmt.Printf("[GATEWAY ERROR] %s - Write error: %v\n", time.Now().Format("15:04:05"), err)
		http.Error(w, `{"error":"Failed to queue task"}`, http.StatusInternalServerError)
		return
	}

	fmt.Printf("[GATEWAY SUCCESS] %s - Task task_%d.json persisted successfully!\n", time.Now().Format("15:04:05"), timestamp)
	fmt.Println("----------------------------------------------------------------")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(fmt.Sprintf(`{"status":"QUEUED","task_id":%d,"message":"Task queued successfully"}`, timestamp)))
}

func main() {
	fmt.Println("====================================================")
	fmt.Println("🚀 GO GATEWAY ACTIVE - V3 (Enhanced Security)")
	fmt.Printf(" Queue directory: %s\n", queueDir)
	fmt.Println(" Listening on port :8080 | Micro-Broker Mode")
	fmt.Println(" Max payload: 5MB | JSON validation enabled")
	fmt.Println("====================================================")

	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/enqueue", handleRequest)
	http.HandleFunc("/enqueue/", handleRequest)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("[ERROR] Server failed: %v\n", err)
		os.Exit(1)
	}
}