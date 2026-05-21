package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // Inicializa o driver do SQLite
)

var (
	queueDir    string
	databasePath string
)

// Estrutura idêntica ao contrato do Gateway para ler o JSON
type Task struct {
	Client string  `json:"client"`
	Score  float64 `json:"score"`
	Event  string  `json:"event"`
}

func init() {
	// Detecta dinamicamente a raiz do projeto para achar as pastas
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(filename))
	
	queueDir = filepath.Join(projectRoot, "shared_queue")
	databasePath = filepath.Join(projectRoot, "database.db")
}

func main() {
	fmt.Println("====================================================")
	fmt.Println("🐹 GO WORKER ACTIVE - V1 (Engine & SQLite Live)")
	fmt.Printf(" Monitoring: %s\n", queueDir)
	fmt.Printf(" Database: %s\n", databasePath)
	fmt.Println("====================================================")

	// 1. CONEXÃO E CRIAÇÃO DO BANCO DE DADOS
	db, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		fmt.Printf("[FATAL] Erro ao abrir SQLite: %v\n", err)
		return
	}
	defer db.Close()

	// Cria a tabela de métricas caso ela não exista no arquivo .db
	schema := `
	CREATE TABLE IF NOT EXISTS metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		client TEXT NOT NULL,
		score REAL NOT NULL,
		event TEXT NOT NULL,
		processed_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err = db.Exec(schema)
	if err != nil {
		fmt.Printf("[FATAL] Erro ao criar tabela SQL: %v\n", err)
		return
	}

	// 2. LOOP INFINITO DE VARREDURA (POLLING)
	for {
		// Lê todos os arquivos da pasta shared_queue
		files, err := os.ReadDir(queueDir)
		if err != nil {
			fmt.Printf("[ERROR] Erro ao ler fila: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Filtra apenas arquivos .json
		var jsonFiles []string
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
				jsonFiles = append(jsonFiles, f.Name())
			}
		}

		// Se houver arquivos na fila, processa o mais antigo (FIFO)
		if len(jsonFiles) > 0 {
			// Ordena os arquivos pelo nome (já que o nome tem o timestamp em nanossegundos)
			sort.Strings(jsonFiles)
			targetFile := jsonFiles[0]
			filePath := filepath.Join(queueDir, targetFile)

			processFile(db, filePath, targetFile)
		}

		// Descansa 500 milissegundos antes da próxima varredura (Super Rápido)
		time.Sleep(500 * time.Millisecond)
	}
}

func processFile(db *sql.DB, path string, filename string) {
	// 1. Lê os bytes do arquivo JSON
	data, err := os.ReadFile(path)
	if err != nil {
		// Se der erro aqui, pode ser que outro Worker rodando em paralelo já tenha pego o arquivo
		return
	}

	// 2. Converte os bytes para a nossa Struct
	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		fmt.Printf("[WORKER ERROR] JSON Corrompido detectado em %s: %v\n", filename, err)
		os.Remove(path) // Remove lixo da fila
		return
	}

	fmt.Printf("[WORKER PROCESSING] 📥 Consumindo job: %s\n", filename)
	
	// Simula processamento rápido (Ex: 300ms de latência de rede/cálculo)
	time.Sleep(300 * time.Millisecond)

	// 3. PERSISTÊNCIA REAL NO BANCO DE DADOS SQLITE
	query := `INSERT INTO metrics (client, score, event) VALUES (?, ?, ?)`
	_, err = db.Exec(query, task.Client, task.Score, task.Event)
	if err != nil {
		fmt.Printf("[WORKER ERROR] Erro ao salvar no SQLite: %v\n", err)
		return
	}

	// 4. PURGA DO ARQUIVO DA FILA (Garante a exclusão segura)
	if err := os.Remove(path); err == nil {
		fmt.Printf("[WORKER SUCCESS] ✔️ Task %s salva no SQLite e limpa do disco.\n", filename)
		fmt.Println("----------------------------------------------------------------")
	}
}