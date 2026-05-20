# Python

import os

readme_content = """# 🚀 MicroQueue: High-Performance Event-Driven Request Broker

O **MicroQueue** é um ecossistema de microsserviços desacoplados e orientados a eventos, projetado para demonstrar uma arquitetura de alta performance, escalabilidade e concorrência híbrida. O projeto combina a velocidade bruta e a eficiência de execução de concorrência do **Go (Golang 1.23+)** na camada de ingestão de dados com a flexibilidade analítica do **Python 3** na camada de processamento e telemetria.

---

## 🏗️ Desenho Arquitetural do Ecossistema

O sistema opera sob o modelo de **Arquitetura Orientada a Eventos (EDA)** através de um canal de trânsito assíncrono por sistema de arquivos (*Shared File-Buffer Queue*), garantindo que o cliente web nunca sofra bloqueios ou gargalos de processamento.

File generated successfully at README.md

```
[ CLIENTE WEB / API TEST ]
            │
    (POST JSON Payload)
            ▼
┌─────────────────────────────────────────────────────────┐
│ 1. HIGH-SPEED INGESTION GATEWAY (Go 1.23+)              │
│    - Escuta na porta HTTP :8080.                        │
│    - Gera IDs únicos temporais em nanossegundos.        │
│    - Valida o método e persiste em bytes no Buffer.     │
│    - Retorna status '202 Accepted' em < 1 milissegundo. │
└───────────────────────────┬─────────────────────────────┘
                            │
               (Escrita Assíncrona em Disco)
                            ▼
                    [ shared_queue/ ]  ◄──── (Buffer Espelhado FIFO)
                            │
               (Varredura e Consumo Temporal)
                            ▼
┌─────────────────────────────────────────────────────────┐
│ 2. ASYNC ANALYTICAL WORKER (Python 3)                   │
│    - Polling ativo e ordenado dos jobs pendentes.       │
│    - Processa e extrai métricas de negócios do JSON.    │
│    - Salva logs imutáveis de auditoria estruturada.     │
│    - Purga o slot consumido da fila de trânsito.        │
└─────────────────────────────────────────────────────────┘
```

---
# 📂 Árvore de Diretórios do Projeto
```
MicroQueue/
│
├── gateway_go/
│   ├── go.mod               # Gerenciador de módulos e dependências do Go
│   └── main.go              # Servidor HTTP Gateway de Ingestão de Alta Velocidade (Go)
│
├── worker_python/
│   └── worker.py            # Motor assíncrono de consumo analítico da fila (Python)
│
├── shared_queue/            # Pasta compartilhada para trânsito de payloads JSON (Ignorada pelo Git)
├── processed_logs/          # Logs permanentes e imutáveis de auditoria de dados (Ignorada pelo Git)
│
├── .gitignore               # Proteção severa de dados locais e isolamento de binários
└── README.md                # Documentação técnica de arquitetura do portfólio
```
---
# ⚡ Simulação e Teste de Carga (Disparo de Eventos)
Com os dois serviços rodando ativamente nas suas respectivas abas, abra uma terceira aba de terminal e envie payloads dinâmicos simulando requisições de clientes reais utilizando o utilitário curl:

```
curl -X POST http://localhost:8080/enqueue \
  -H "Content-Type: application/json" \
  -d '{"client": "Davi_Passos", "score": 9.8, "status": "approved", "event_origin": "WSL2_Terminal"}'
```
