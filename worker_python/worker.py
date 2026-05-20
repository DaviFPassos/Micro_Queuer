import os
import time
import json
import sys
from datetime import datetime
from pathlib import Path

# Usa caminhos absolutos para evitar problemas de diretório
BASE_DIR = Path(__file__).parent.absolute()
QUEUE_DIR = BASE_DIR.parent / "shared_queue"
LOGS_DIR = BASE_DIR.parent / "processed_logs"

# CRÍTICO: Garante que as pastas existem
os.makedirs(QUEUE_DIR, exist_ok=True)
os.makedirs(LOGS_DIR, exist_ok=True)

print("====================================================")
print("🐍 PYTHON WORKER ACTIVE - V3 (Enhanced & Validated)")
print(f" Queue directory: {QUEUE_DIR}")
print(f" Logs directory: {LOGS_DIR}")
print("====================================================")

def validate_json(file_path):
    # Valida se o arquivo contém JSON válido
    try:
        with open(file_path, 'r') as f:
            json.load(f)
        return True
    except (json.JSONDecodeError, ValueError):
        print(f"[WORKER ❌] Invalid JSON in {file_path}")
        return False

def process_task(task_file):
    # Processa uma tarefa de forma segura
    task_path = QUEUE_DIR / task_file
    
    # Valida se o arquivo existe e contém JSON válido
    if not task_path.exists():
        print(f"[WORKER (ATENTION)] File disappeared: {task_file}")
        return False
    
    if not validate_json(task_path):
        return False
    
    try:
        with open(task_path, 'r') as f:
            payload = json.load(f)
        
        print(f"\n[WORKER] 📥 Pulling job from queue: {task_file}")
        print(f"[WORKER] Executing metrics analysis on payload: {payload}")
        
        # Simula tempo de processamento analítico
        time.sleep(0.5)
        
        # Cria o registro de sucesso permanente
        log_data = {
            "task_processed": task_file,
            "timestamp_end": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
            "status": "PROCESSED_SUCCESSFULLY",
            "original_payload": payload
        }
        
        log_file_path = LOGS_DIR / f"log_{task_file}"
        with open(log_file_path, 'w') as lf:
            json.dump(log_data, lf, indent=4)
            
        print(f"[WORKER (CORRECT)] Analytics completed. Audit log saved at processed_logs/log_{task_file}")
        return True
        
    except Exception as e:
        print(f"[WORKER X] Error processing task {task_file}: {e}")
        return False
    finally:
        # Remove o arquivo da fila com verificação
        try:
            if task_path.exists():
                task_path.unlink()
                print("[WORKER] Queue slot cleared and released.")
                print("-" * 60)
        except Exception as e:
            print(f"[WORKER (ATENTION)] Failed to remove {task_file}: {e}")

def main():
    # Loop principal que monitora a fila
    while True:
        try:
            # Lista arquivos JSON na fila
            if not QUEUE_DIR.exists():
                print("[WORKER (ATENTION)] Queue directory not found, retrying...")
                time.sleep(1)
                continue
            
            files = sorted([f.name for f in QUEUE_DIR.glob('task_*.json')])
            
            if files:
                # Processa o arquivo mais antigo (FIFO)
                task_file = files[0]
                process_task(task_file)
            else:
                # Nenhuma tarefa na fila, aguarda
                time.sleep(0.5)
                
        except Exception as e:
            print(f"[WORKER X] Critical error in main loop: {e}")
            time.sleep(1)

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\n[WORKER] Shutting down gracefully...")
        sys.exit(0)
    except Exception as e:
        print(f"[WORKER X] Fatal error: {e}")
        sys.exit(1)