# 🚀 MicroQueue: High-Performance Distributed Request Broker

O **MicroQueue** é um ecossistema de microsserviços desacoplados e orientados a eventos, projetado para demonstrar conceitos avançados de alta performance, concorrência nativa e governança de segurança. 

Originalmente concebido como uma arquitetura híbrida, o projeto foi refatorado para operar **100% em Go (Golang 1.23+)** na camada de backend (Gateway e Workers) para maximizar a eficiência de I/O e o isolamento de memória, utilizando uma interface reativa minimalista em **Vanilla JS** no Frontend e persistência relacional estruturada via **SQLite**.

---

## 🏗️ Desenho Arquitetural do Ecossistema

O sistema opera sob o modelo de **Arquitetura Orientada a Eventos (EDA)** com processamento assíncrono utilizando o próprio sistema de arquivos local (*Shared File-Buffer Queue*) como barramento de trânsito. Isso garante um desacoplamento total: a camada de ingestão nunca bloqueia o cliente web, independentemente da carga.

```text
[ FRONTEND WEB INTERFACE ] (Navegador Windows / Vanilla JS)
            │
    (Fetch API - HTTP POST)
            ▼
┌─────────────────────────────────────────────────────────┐
│ 1. HIGH-SPEED INGESTION GATEWAY (Go 1.23+)              │
│    - Escuta na porta HTTP :8080 (WSL2 / Linux).         │
│    - Filtro Anti-DoS: Limita requisições a 5MB.         │
│    - Validação de Integridade: Sanitização de JSON.      │
│    - Retorna status '202 Accepted' em < 1 milissegundo. │
└───────────────────────────┬─────────────────────────────┘
                            │
               (Escrita Assíncrona em Disco)
                            ▼
                    [ shared_queue/ ]  ◄──── (Fila Transicional FIFO)
                            │
               (Concorrência via Polling Ativo)
              ┌─────────────┴─────────────┐
              ▼                           ▼
┌───────────────────────────┐ ┌───────────────────────────┐
│ 2. DISTRIBUTED WORKER #1  │ │ 3. DISTRIBUTED WORKER #2  │ (Motores em Go)
│ - Consome jobs ordenados. │ │ - Consome jobs ordenados. │
│ - Evita colisões nativas. │ │ - Evita colisões nativas. │
└─────────────┬─────────────┘ └─────────────┬─────────────┘
              │                             │
              └──────────────┬──────────────┘
                             ▼
                 [ SQLite DATABASE (.db) ] ◄─── (Persistência Imutável)
```
---
# 📂 Árvore de Diretórios do Projeto
```
MicroQueue/
│
├── frontend/
│   └── index.html           # Dashboard reativo para disparo de payloads (Vanilla JS)
│
├── gateway_go/
│   ├── go.mod               # Módulo do Gateway de Ingestão
│   └── main.go              # Servidor HTTP Gateway com CORS e Proteção Anti-DoS (Go)
│
├── worker_go/
│   ├── go.mod               # Módulo do Worker e Driver SQL
│   └── worker.go            # Motor concorrente de consumo FIFO e escrita relacional (Go)
│
├── legacy_python/           # Histórico de evolução técnica da primeira versão (Python)
│
├── shared_queue/            # Buffer temporário de trânsito de payloads JSON (Ignorado pelo Git)
├── database.db              # Arquivo unificado do Banco de Dados SQLite (Ignorado pelo Git)
│
├── .gitignore               # Regras severas de governança e proteção contra vazamento de dados
└── README.md                # Documentação arquitetural do ecossistema
```
---
