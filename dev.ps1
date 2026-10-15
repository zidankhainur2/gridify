<#
.SYNOPSIS
    Gridify Unified Development Server Runner (PowerShell)

.DESCRIPTION
    Script otomatis untuk menjalankan seluruh layanan Gridify (PostgreSQL, Backend Go API, dan Frontend Next.js).
    Secara cerdas mendeteksi dan menyalakan paksa sistem yang belum aktif:
      1. Docker Desktop / Docker Engine (jika mati, otomatis dinyalakan dan ditunggu hingga siap).
      2. PostgreSQL 17 Container (dijalankan via docker compose dan divalidasi dengan pg_isready).
      3. Migrasi Database (otomatis menjalankan 'migrate up' agar schema selalu mutakhir).
      4. Dependensi Frontend (otomatis menjalankan 'npm install' jika node_modules belum ada).
      5. Pembersihan Port (menghentikan paksa proses lama yang menempel di port 8080 atau 3000).
      6. Menjalankan Backend Go API dan Frontend Next.js secara bersamaan dengan log berwarna.

.PARAMETER Windowed
    Jalankan backend dan frontend di jendela terminal terpisah (PowerShell Window).

.PARAMETER NoMigrate
    Lewati proses migrasi database saat startup.

.PARAMETER Seed
    Jalankan seeder database dummy turnamen (cmd/seed) setelah migrasi berhasil.

.PARAMETER Stop
    Hentikan seluruh proses backend (port 8080) dan frontend (port 3000) yang sedang berjalan.

.PARAMETER StopDb
    Gunakan bersamaan dengan -Stop untuk mematikan kontainer Docker PostgreSQL juga.

.EXAMPLE
    .\dev.ps1
    Menjalankan seluruh stack layanan secara terpadu di konsol saat ini.

.EXAMPLE
    .\dev.ps1 -Windowed
    Menjalankan backend dan frontend di jendela terpisah.

.EXAMPLE
    .\dev.ps1 -Seed
    Menjalankan layanan sekaligus mengisi database dengan data dummy turnamen.

.EXAMPLE
    .\dev.ps1 -Stop
    Menghentikan semua proses backend dan frontend yang masih berjalan di latar belakang.
#>

[CmdletBinding()]
param (
    [switch]$Windowed,
    [switch]$NoMigrate,
    [switch]$Seed,
    [switch]$Stop,
    [switch]$StopDb
)

$ErrorActionPreference = "Stop"
$ScriptDir = $PSScriptRoot
if (-not $ScriptDir) {
    $ScriptDir = (Get-Location).Path
}

# ====================================================================
# Visual Helpers & Formatting
# ====================================================================

function Show-Banner {
    Write-Host @"

  ====================================================================
     ____       _     _ _  __       
    / ___|_ __ (_) __| (_)/ _|_   _ 
   | |  _| '__|| |/ _` | | |_| | | |   DEVELOPMENT ENVIRONMENT
   | |_| | |   | | (_| | |  _| |_| |   Esports Operations Platform
    \____|_|   |_|\__,_|_|_|  \__, |
                              |___/ 
  ====================================================================
"@ -ForegroundColor DarkCyan
}

function Write-Step {
    param([string]$Title)
    Write-Host "`n>>> $Title" -ForegroundColor Cyan
}

function Write-Info {
    param([string]$Tag, [string]$Message)
    Write-Host "  [$Tag] " -NoNewline -ForegroundColor DarkCyan
    Write-Host $Message -ForegroundColor White
}

function Write-Success {
    param([string]$Tag, [string]$Message)
    Write-Host "  [$Tag] " -NoNewline -ForegroundColor Green
    Write-Host $Message -ForegroundColor Green
}

function Write-Warn {
    param([string]$Tag, [string]$Message)
    Write-Host "  [$Tag] " -NoNewline -ForegroundColor Yellow
    Write-Host $Message -ForegroundColor Yellow
}

function Write-Err {
    param([string]$Tag, [string]$Message)
    Write-Host "  [$Tag] " -NoNewline -ForegroundColor Red
    Write-Host $Message -ForegroundColor Red
}

# ====================================================================
# Port & Process Management
# ====================================================================

function Stop-ProcessOnPort {
    param(
        [int]$Port,
        [string]$ServiceName
    )

    $conns = Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue
    if ($conns) {
        $pids = $conns | Select-Object -ExpandProperty OwningProcess -Unique
        foreach ($p in $pids) {
            if ($p -gt 0 -and $p -ne $PID) {
                Write-Warn "PORT" "Port $Port ($ServiceName) sedang dipakai oleh PID $p. Menghentikan paksa..."
                try {
                    & taskkill.exe /F /T /PID $p 2>$null | Out-Null
                } catch {
                    try {
                        Stop-Process -Id $p -Force -ErrorAction SilentlyContinue
                    } catch {}
                }
            }
        }
        Start-Sleep -Milliseconds 400
    }
}

# ====================================================================
# Docker & Database Engine Management
# ====================================================================

function Find-DockerDesktopPath {
    $possiblePaths = @(
        "$env:ProgramFiles\Docker\Docker\Docker Desktop.exe",
        "${env:ProgramFiles(x86)}\Docker\Docker\Docker Desktop.exe",
        "$env:LOCALAPPDATA\Programs\Docker\Docker\Docker Desktop.exe"
    )

    foreach ($path in $possiblePaths) {
        if ($path -and (Test-Path $path)) {
            return $path
        }
    }

    try {
        $regPath = (Get-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\Docker Desktop.exe" -ErrorAction SilentlyContinue).'(Default)'
        if ($regPath -and (Test-Path $regPath)) {
            return $regPath
        }
    } catch {}

    $cmd = Get-Command "Docker Desktop.exe" -ErrorAction SilentlyContinue
    if ($cmd) {
        return $cmd.Source
    }

    return $null
}

function Test-DockerDaemon {
    try {
        $null = & docker info 2>&1
        return ($LASTEXITCODE -eq 0)
    } catch {
        return $false
    }
}

function Ensure-DockerRunning {
    Write-Step "1. Memeriksa Status Docker Engine"

    if (Test-DockerDaemon) {
        Write-Success "DOCKER" "Docker daemon sudah berjalan dan responsif."
        return
    }

    Write-Warn "DOCKER" "Docker daemon tidak terdeteksi aktif. Mencari Docker Desktop..."
    $dockerPath = Find-DockerDesktopPath

    if (-not $dockerPath) {
        Write-Err "DOCKER" "Aplikasi Docker Desktop tidak ditemukan di lokasi standar!"
        Write-Err "DOCKER" "Pastikan Docker Desktop telah terpasang di komputer Anda."
        throw "Docker Desktop tidak ditemukan."
    }

    Write-Info "DOCKER" "Menyalakan paksa Docker Desktop dari: $dockerPath"
    try {
        Start-Process -FilePath $dockerPath -WindowStyle Minimized
    } catch {
        Write-Err "DOCKER" "Gagal menjalankan Docker Desktop: $_"
        throw $_
    }

    Write-Info "DOCKER" "Menunggu Docker Engine menyala dan siap menerima perintah..."
    $timeoutSeconds = 120
    $startTime = Get-Date
    $started = $false

    while (-not $started) {
        Start-Sleep -Seconds 3
        if (Test-DockerDaemon) {
            $started = $true
            break
        }

        $elapsed = [int]((Get-Date) - $startTime).TotalSeconds
        if ($elapsed -ge $timeoutSeconds) {
            Write-Err "DOCKER" "Waktu tunggu habis ($timeoutSeconds detik). Docker Engine belum merespons."
            Write-Err "DOCKER" "Silakan buka antarmuka Docker Desktop untuk memastikan status inisialisasi WSL/Hyper-V."
            throw "Timeout menunggu Docker Engine."
        }

        Write-Host -NoNewline "." -ForegroundColor Yellow
    }
    Write-Host ""
    Write-Success "DOCKER" "Docker Engine berhasil menyala dan siap digunakan!"
}

function Ensure-PostgresRunning {
    Write-Step "2. Memeriksa Layanan Database PostgreSQL"

    Write-Info "POSTGRES" "Menjalankan kontainer PostgreSQL via Docker Compose..."
    Push-Location $ScriptDir
    try {
        & docker compose up -d postgres
        if ($LASTEXITCODE -ne 0) {
            throw "Gagal menjalankan 'docker compose up -d postgres'"
        }
    } finally {
        Pop-Location
    }

    Write-Info "POSTGRES" "Menunggu database PostgreSQL siap menerima query..."
    $maxWait = 30
    $waited = 0
    $ready = $false

    while ($waited -lt $maxWait) {
        & docker compose exec -T postgres pg_isready -U postgres -d gridify 2>&1 | Out-Null
        if ($LASTEXITCODE -eq 0) {
            $ready = $true
            break
        }
        Start-Sleep -Seconds 1
        $waited++
        Write-Host -NoNewline "." -ForegroundColor Cyan
    }
    Write-Host ""

    if ($ready) {
        Write-Success "POSTGRES" "PostgreSQL 17 aktif dan siap menerima koneksi (Port 5432)!"
    } else {
        Write-Warn "POSTGRES" "Pemeriksaan pg_isready belum memberi respons positif. Melanjutkan inisialisasi..."
    }
}

# ====================================================================
# Environment & Dependencies
# ====================================================================

function Load-EnvironmentVariables {
    $envPath = "$ScriptDir\.env"
    if (Test-Path $envPath) {
        Get-Content $envPath | ForEach-Object {
            $line = $_.Trim()
            if ($line -and -not $line.StartsWith("#") -and $line.Contains("=")) {
                $parts = $line.Split("=", 2)
                $k = $parts[0].Trim()
                $v = $parts[1].Trim().Trim('"').Trim("'")
                # Jangan set PORT secara global agar Next.js tidak bentrok dengan Go API (8080)
                if ($k -and $k -ne "PORT") {
                    [Environment]::SetEnvironmentVariable($k, $v, "Process")
                }
            }
        }
    }
    # Pastikan PORT global tidak menempel di sesi PowerShell saat ini
    [Environment]::SetEnvironmentVariable("PORT", $null, "Process")
}

function Ensure-Prerequisites {
    Write-Step "3. Memeriksa Tooling & Dependensi"

    # Go check
    $go = Get-Command go -ErrorAction SilentlyContinue
    if (-not $go) {
        Write-Err "SYSTEM" "Go compiler tidak ditemukan di PATH. Pasang Go 1.24+ untuk melanjutkan."
        throw "Go compiler tidak tersedia."
    }
    Write-Success "SYSTEM" "Go siap: $(& go version)"

    # Node check
    $npm = Get-Command npm -ErrorAction SilentlyContinue
    if (-not $npm) {
        Write-Err "SYSTEM" "Node.js / npm tidak ditemukan di PATH."
        throw "Node.js / npm tidak tersedia."
    }
    Write-Success "SYSTEM" "Node siap: $(& node -v) | npm $(& npm -v)"

    # Frontend node_modules check
    $nodeModules = "$ScriptDir\frontend\node_modules"
    if (-not (Test-Path $nodeModules)) {
        Write-Warn "FRONTEND" "Direktori 'node_modules' tidak ditemukan. Menjalankan npm install secara otomatis..."
        Push-Location "$ScriptDir\frontend"
        try {
            & npm install
            if ($LASTEXITCODE -ne 0) {
                throw "Gagal menjalankan 'npm install'"
            }
            Write-Success "FRONTEND" "Instalasi paket frontend selesai."
        } finally {
            Pop-Location
        }
    } else {
        Write-Success "FRONTEND" "Dependensi frontend lengkap (node_modules tersedia)."
    }
}

# ====================================================================
# Migrations & Seeding
# ====================================================================

function Invoke-DatabaseMigrations {
    if ($NoMigrate) {
        Write-Warn "MIGRATE" "Migrasi database dilewati (--NoMigrate)."
        return
    }

    Write-Step "4. Sinkronisasi Skema Database (Migrations)"
    Write-Info "MIGRATE" "Menjalankan migrasi database (cmd/migrate up)..."
    Push-Location "$ScriptDir\backend"
    try {
        & go run ./cmd/migrate up
        if ($LASTEXITCODE -ne 0) {
            throw "Database migration gagal dengan exit code $LASTEXITCODE."
        }
        Write-Success "MIGRATE" "Skema database up-to-date."
    } finally {
        Pop-Location
    }

    if ($Seed) {
        Write-Info "SEED" "Menjalankan seeder akun dan turnamen dummy..."
        Push-Location "$ScriptDir\backend"
        try {
            & go run ./cmd/seed
            if ($LASTEXITCODE -ne 0) {
                Write-Warn "SEED" "Seeder selesai dengan kode keluar $LASTEXITCODE."
            } else {
                Write-Success "SEED" "Data dummy berhasil dimasukkan ke basis data."
            }
        } finally {
            Pop-Location
        }
    }
}

# ====================================================================
# Main Execution Flow
# ====================================================================

Show-Banner

# Handle -Stop parameter
if ($Stop) {
    Write-Step "Menghentikan Semua Layanan Gridify"
    Stop-ProcessOnPort -Port 8080 -ServiceName "Backend API"
    Stop-ProcessOnPort -Port 3000 -ServiceName "Frontend Web"

    if ($StopDb) {
        Write-Info "DOCKER" "Mematikan kontainer PostgreSQL..."
        Push-Location $ScriptDir
        & docker compose stop postgres
        Pop-Location
        Write-Success "DOCKER" "Kontainer PostgreSQL dimatikan."
    }

    Write-Success "GRIDIFY" "Seluruh proses Gridify berhasil dihentikan secara bersih."
    return
}

# Step 1: Check & Force Start Docker
Ensure-DockerRunning

# Step 2: Ensure PostgreSQL is up and healthy
Ensure-PostgresRunning

# Step 3: Load environment variables from .env
Load-EnvironmentVariables

# Step 4: Ensure prerequisites & npm dependencies
Ensure-Prerequisites

# Step 5: Run migrations & optional seed
Invoke-DatabaseMigrations

# Step 6: Free up ports 8080 and 3000 from old hanging processes
Write-Step "5. Memeriksa Port & Membersihkan Proses Lama"
Stop-ProcessOnPort -Port 8080 -ServiceName "Backend API"
Stop-ProcessOnPort -Port 3000 -ServiceName "Frontend Web"
Write-Success "PORT" "Port 8080 dan 3000 bebas dan siap digunakan."

# ====================================================================
# Service Execution Modes
# ====================================================================

if ($Windowed) {
    Write-Step "6. Menjalankan Layanan di Jendela Terpisah"

    $backendScript = @"
`$host.UI.RawUI.WindowTitle = 'Gridify Backend API (Port 8080)'
Set-Location '$ScriptDir\backend'
Write-Host '===================================================' -ForegroundColor Cyan
Write-Host '  GRIDIFY BACKEND GO API SERVER' -ForegroundColor Cyan
Write-Host '  Endpoint: http://localhost:8080' -ForegroundColor Cyan
Write-Host '===================================================' -ForegroundColor Cyan
go run ./cmd/api
"@

    $frontendScript = @"
`$host.UI.RawUI.WindowTitle = 'Gridify Frontend Web (Port 3000)'
Set-Location '$ScriptDir\frontend'
`$env:PORT = '3000'
`$env:NEXT_PUBLIC_API_URL = 'http://localhost:8080/api/v1'
Write-Host '===================================================' -ForegroundColor Magenta
Write-Host '  GRIDIFY FRONTEND NEXT.JS DEV SERVER' -ForegroundColor Magenta
Write-Host '  URL: http://localhost:3000' -ForegroundColor Magenta
Write-Host '===================================================' -ForegroundColor Magenta
npm run dev -- -p 3000
"@

    # Launch windows
    Start-Process powershell.exe -ArgumentList "-NoExit", "-Command", $backendScript
    Start-Process powershell.exe -ArgumentList "-NoExit", "-Command", $frontendScript

    Write-Host @"

  ====================================================================
  Layanan Gridify Berhasil Dinyalakan di Jendela Terpisah!
  --------------------------------------------------------------------
  Frontend Web  : http://localhost:3000
  Backend API   : http://localhost:8080
  PostgreSQL    : localhost:5432 (gridify)
  --------------------------------------------------------------------
  Tips:
  - Tekan [ENTER] di jendela ini untuk mematikan semua layanan sekaligus.
  - Atau jalankan: .\dev.ps1 -Stop
  ====================================================================
"@ -ForegroundColor Green

    Read-Host "Tekan [ENTER] kapan saja untuk menghentikan seluruh layanan..."
    Stop-ProcessOnPort -Port 8080 -ServiceName "Backend API"
    Stop-ProcessOnPort -Port 3000 -ServiceName "Frontend Web"
    Write-Success "GRIDIFY" "Semua layanan telah dihentikan."
    return
}

# Default: Single Console Unified Output Mode
Write-Step "6. Menjalankan Backend & Frontend Secara Bersamaan"

Write-Host @"
  --------------------------------------------------------------------
  Frontend Web  : http://localhost:3000
  Backend API   : http://localhost:8080
  PostgreSQL    : localhost:5432 (Database: gridify)
  --------------------------------------------------------------------
  Tekan [Ctrl + C] untuk mematikan seluruh layanan sekaligus secara aman.
  ====================================================================
"@ -ForegroundColor DarkCyan

# Setup Backend Process (Go API)
$backendPsi = New-Object System.Diagnostics.ProcessStartInfo
$backendPsi.FileName = "cmd.exe"
$backendPsi.Arguments = "/c go run ./cmd/api"
$backendPsi.WorkingDirectory = "$ScriptDir\backend"
$backendPsi.UseShellExecute = $false
$backendPsi.RedirectStandardOutput = $true
$backendPsi.RedirectStandardError = $true
$backendPsi.CreateNoWindow = $true
$backendPsi.EnvironmentVariables["PORT"] = "8080"

$backendProc = New-Object System.Diagnostics.Process
$backendProc.StartInfo = $backendPsi
$backendProc.EnableRaisingEvents = $true

$bOutEvent = Register-ObjectEvent -InputObject $backendProc -EventName "OutputDataReceived" -Action {
    if ($EventArgs.Data) {
        Write-Host "[API]  " -NoNewline -ForegroundColor Cyan
        Write-Host $EventArgs.Data
    }
}
$bErrEvent = Register-ObjectEvent -InputObject $backendProc -EventName "ErrorDataReceived" -Action {
    if ($EventArgs.Data) {
        Write-Host "[API]  " -NoNewline -ForegroundColor DarkCyan
        Write-Host $EventArgs.Data -ForegroundColor Yellow
    }
}

# Setup Frontend Process (Next.js)
$frontendPsi = New-Object System.Diagnostics.ProcessStartInfo
$frontendPsi.FileName = "cmd.exe"
$frontendPsi.Arguments = "/c npm run dev -- -p 3000"
$frontendPsi.WorkingDirectory = "$ScriptDir\frontend"
$frontendPsi.UseShellExecute = $false
$frontendPsi.RedirectStandardOutput = $true
$frontendPsi.RedirectStandardError = $true
$frontendPsi.CreateNoWindow = $true
$frontendPsi.EnvironmentVariables["PORT"] = "3000"
$frontendPsi.EnvironmentVariables["NEXT_PUBLIC_API_URL"] = "http://localhost:8080/api/v1"

$frontendProc = New-Object System.Diagnostics.Process
$frontendProc.StartInfo = $frontendPsi
$frontendProc.EnableRaisingEvents = $true

$fOutEvent = Register-ObjectEvent -InputObject $frontendProc -EventName "OutputDataReceived" -Action {
    if ($EventArgs.Data) {
        Write-Host "[WEB]  " -NoNewline -ForegroundColor Magenta
        Write-Host $EventArgs.Data
    }
}
$fErrEvent = Register-ObjectEvent -InputObject $frontendProc -EventName "ErrorDataReceived" -Action {
    if ($EventArgs.Data) {
        Write-Host "[WEB]  " -NoNewline -ForegroundColor DarkMagenta
        Write-Host $EventArgs.Data -ForegroundColor Yellow
    }
}

try {
    # Start Backend
    $backendProc.Start() | Out-Null
    $backendProc.BeginOutputReadLine()
    $backendProc.BeginErrorReadLine()

    # Start Frontend
    $frontendProc.Start() | Out-Null
    $frontendProc.BeginOutputReadLine()
    $frontendProc.BeginErrorReadLine()

    # Loop waiting for processes or Ctrl+C
    while (-not $backendProc.HasExited -and -not $frontendProc.HasExited) {
        Start-Sleep -Milliseconds 250
    }

    if ($backendProc.HasExited) {
        Write-Warn "SYSTEM" "Backend API terhenti (ExitCode: $($backendProc.ExitCode))."
    }
    if ($frontendProc.HasExited) {
        Write-Warn "SYSTEM" "Frontend server terhenti (ExitCode: $($frontendProc.ExitCode))."
    }
}
finally {
    Write-Host ""
    Write-Step "Pembersihan & Penghentian Layanan"

    # Unregister events
    if ($bOutEvent) { Unregister-Event -SourceIdentifier $bOutEvent.Name -ErrorAction SilentlyContinue }
    if ($bErrEvent) { Unregister-Event -SourceIdentifier $bErrEvent.Name -ErrorAction SilentlyContinue }
    if ($fOutEvent) { Unregister-Event -SourceIdentifier $fOutEvent.Name -ErrorAction SilentlyContinue }
    if ($fErrEvent) { Unregister-Event -SourceIdentifier $fErrEvent.Name -ErrorAction SilentlyContinue }

    # Terminate Backend Process Tree
    if ($backendProc -and -not $backendProc.HasExited) {
        Write-Info "CLEANUP" "Menghentikan Backend API (PID $($backendProc.Id))..."
        & taskkill.exe /F /T /PID $backendProc.Id 2>$null | Out-Null
        $backendProc.Dispose()
    }

    # Terminate Frontend Process Tree
    if ($frontendProc -and -not $frontendProc.HasExited) {
        Write-Info "CLEANUP" "Menghentikan Frontend dev server (PID $($frontendProc.Id))..."
        & taskkill.exe /F /T /PID $frontendProc.Id 2>$null | Out-Null
        $frontendProc.Dispose()
    }

    # Ensure ports are free
    Stop-ProcessOnPort -Port 8080 -ServiceName "Backend API"
    Stop-ProcessOnPort -Port 3000 -ServiceName "Frontend Web"

    Write-Success "GRIDIFY" "Seluruh layanan berhasil dihentikan secara bersih."
}
