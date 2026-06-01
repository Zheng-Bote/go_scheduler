Erzeuge ein vollständiges, lauffähiges Go-Projekt, das als Linux-Commandline-Scheduler fungiert.
Der Scheduler startet und überwacht andere Go-Programme gemäß Konfiguration aus einer PostgreSQL-Datenbank.
Alle Anforderungen unten sind strikt einzuhalten.
Erzeuge vollständigen, kompilierbaren Code ohne Platzhalter.

=====================================================================
ZIELE
=====================================================================
Erstelle ein vollständiges Go-Projekt mit:

1. Scheduler (Go CLI)
2. IPC über Unix-Domain-Socket (JSON-Lines)
3. HTTP-Server für /info und /health
4. Cron-Scheduling (MIT-lizenziert)
5. PostgreSQL-Anbindung (MIT-lizenziert)
6. Starten externer Go-Programme (Jobs)
7. Status-Events der Jobs → Scheduler → PostgreSQL
8. Dockerfile für Ubuntu latest
9. SQL-Migrations für alle Tabellen
10. Beispiel-Jobs, die IPC nutzen
11. **Verschlüsselte INI-Konfiguration (AES-256-GCM + Argon2id)**

=====================================================================
LIZENZBEDINGUNGEN
=====================================================================
Es dürfen ausschließlich Open-Source-Komponenten verwendet werden, die
kommerzielle/gewerbliche Nutzung erlauben:

- Go Stdlib
- github.com/jackc/pgx/v5 (MIT)
- github.com/robfig/cron/v3 (MIT)

Keine weiteren externen Libraries.

=====================================================================
VERSCHLÜSSELTE INI-KONFIGURATION
=====================================================================

Der Scheduler erhält beim Start einen Parameter:

    scheduler <Pfad/zur/INI-Datei>

Diese INI-Datei ist **vollständig verschlüsselt**.

Anforderungen:

1. **AES-256-GCM** für symmetrische Verschlüsselung
2. **Argon2id** für Key-Derivation
   - time = 3
   - memory = 64 MB
   - parallelism = 1
   - key length = 32 Bytes
3. Die INI-Datei enthält:
   - PostgreSQL-Host
   - PostgreSQL-Port
   - PostgreSQL-User
   - PostgreSQL-Passwort
   - PostgreSQL-Datenbankname
4. Der Scheduler:
   - liest die verschlüsselte Datei
   - leitet aus einem Passwort (CLI-Prompt oder ENV) den Key via Argon2id ab
   - entschlüsselt die INI-Datei
   - parsed die INI-Daten
   - baut daraus den PostgreSQL-DSN

5. Erzeuge zusätzlich ein CLI-Tool:

   encrypt-config

   Dieses Tool:
   - liest eine normale INI-Datei
   - fragt ein Passwort ab
   - erzeugt eine verschlüsselte INI-Datei (AES-256-GCM)
   - speichert Salt + Nonce + Ciphertext in einer Datei

=====================================================================
DATENBANK-SCHEMA
=====================================================================

Erstelle SQL-Migrations für folgende Tabellen:

scheduled_programs:

- id SERIAL PK
- name TEXT
- command TEXT
- args TEXT[]
- cron_expr TEXT
- enabled BOOLEAN
- restart_on_exit BOOLEAN
- created_at TIMESTAMPTZ
- updated_at TIMESTAMPTZ

program_runs:

- id SERIAL PK
- program_id INT FK
- pid INT
- started_at TIMESTAMPTZ
- finished_at TIMESTAMPTZ
- exit_code INT
- success BOOLEAN

job_status_events:

- id SERIAL PK
- run_id INT FK
- status TEXT
- message TEXT
- progress INT
- ts TIMESTAMPTZ

scheduler_config:

- id SERIAL PK
- http_port INT
- socket_path TEXT

=====================================================================
SCHEDULER – FUNKTIONALE ANFORDERUNGEN
=====================================================================

1. Scheduler lädt Konfiguration aus scheduler_config:
   - http_port
   - socket_path

2. Scheduler lädt verschlüsselte INI-Datei:
   - Passwort → Argon2id → AES-Key
   - AES-256-GCM entschlüsseln
   - INI parsen
   - PostgreSQL-DSN erzeugen

3. Scheduler startet:
   - Unix-Domain-Socket-Server
   - HTTP-Server
   - Cron-Scheduler
   - Job-Start-Logik

4. HTTP-Server:
   GET /info:
   - JSON: name, description, version
     GET /health:
   - DB-Test (SELECT true)
   - HTTP 200 bei Erfolg

5. IPC über Unix-Domain-Socket:
   - Jobs senden JSON-Lines:
     {
     "run_id": 123,
     "status": "processing",
     "message": "step 1",
     "progress": 42
     }

6. Scheduler speichert IPC-Events in job_status_events.

7. Scheduler startet Jobs:
   - legt program_runs an
   - setzt RUN_ID und SCHEDULER_SOCKET_PATH als Environment-Variablen
   - startet Prozess via exec.Command
   - schreibt PID in DB
   - wartet auf Prozessende
   - aktualisiert program_runs

8. Cron:
   - lädt alle enabled scheduled_programs
   - registriert Cron-Jobs
   - führt runProgram() aus

=====================================================================
JOB-PROGRAMME (Beispiele)
=====================================================================

Erstelle mindestens zwei Beispiel-Jobs:

job1:

- sendet "started"
- simuliert Arbeit (Sleep)
- sendet "processing"
- sendet "finished"

job2:

- sendet mehrere Fortschritts-Events (progress 0–100)

Beide Jobs:

- lesen RUN_ID und SCHEDULER_SOCKET_PATH aus os.Getenv
- senden IPC-Events über Unix-Socket

=====================================================================
PROJEKTSTRUKTUR
=====================================================================

Erzeuge folgende Struktur:

/cmd/scheduler/main.go
/cmd/job1/main.go
/cmd/job2/main.go
/cmd/encrypt-config/main.go
/internal/crypto/...
/internal/config/...
/internal/db/...
/internal/ipc/...
/internal/http/...
/internal/scheduler/...
/internal/jobs/...
/migrations/\*.sql
/go.mod
/go.sum
/Dockerfile
/README.md

=====================================================================
DOCKER
=====================================================================

Dockerfile (Ubuntu latest):

- COPY aller Binaries
- EXPOSE http_port
- CMD scheduler <Pfad/zur/INI-Datei>

=====================================================================
AUSGABEFORMAT
=====================================================================

Gib ALLE Dateien vollständig aus:

- go.mod + go.sum
- vollständiger Go-Code aller Dateien
- vollständige SQL-Migrations
- vollständiges Dockerfile
- vollständiges README.md mit Build- und Run-Anleitung

KEINE Platzhalter, KEINE Auslassungen.

=====================================================================
JETZT STARTEN
=====================================================================

Erzeuge jetzt das vollständige Projekt.
