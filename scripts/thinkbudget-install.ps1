<#
.SYNOPSIS
    Install a thinking-budget build over an existing Ollama on Windows, and
    take the install out of the Microsoft Store's reach.

.DESCRIPTION
    Two jobs, and the second one is the point.

    Copying the binaries is the easy half. The half that was missing is the
    Add/Remove Programs entry: a machine that ever had stock Ollama installed
    keeps the entry its Inno Setup installer wrote, and the Windows Package
    Manager correlates that entry -- by DisplayName and Publisher, there being
    no ProductCode in the manifest -- to the `Ollama.Ollama` package. The
    Microsoft Store then upgrades it on its own schedule, running the stock
    installer over whatever is there.

    Measured on eleven2go: `0.34.0-thinkbudget` was installed 2026-09-17 07:13
    and replaced by stock 0.34.2 at 21:57 the same day, four seconds apart for
    `ollama.exe` and `ollama app.exe`. The running server kept the binary it had
    already loaded and went on answering `0.34.0-thinkbudget` until it was
    restarted 14 hours later, at which point stock 0.34.2 would not start at all
    (`Failed to start: Unable to init instance`). The lane was down and the A/B
    batch on it burned its remaining runs against a dead endpoint.

    None of that is reachable from inside Ollama. The fork's own updater is
    redirected to the fork's releases (app/updater/fork.go) and gated on consent
    (app/updater/consent.go), and neither is consulted: the Store is a separate
    process doing a separate install. The only lever on this side is the
    identity the Store matches against, so this claims one of our own.

    Verified on eleven2go 2026-09-18 -- after the rewrite,
    `winget list --id Ollama.Ollama` answers "No installed package found
    matching input criteria." and `winget upgrade` no longer lists it.

.PARAMETER Source
    Directory holding `ollama.exe` and `ollama app.exe` to install.

.PARAMETER Version
    Version string recorded in our ARP entry, e.g. 0.34.2-thinkbudget.

.PARAMETER IdentityOnly
    Claim the identity and skip the file copy. For a host that already has the
    right binaries and only needs to be taken off the Store's list.

.EXAMPLE
    .\thinkbudget-install.ps1 -Source C:\ollama-tb\0342 -Version 0.34.2-thinkbudget

.EXAMPLE
    .\thinkbudget-install.ps1 -IdentityOnly -Version 0.34.0-thinkbudget
#>
[CmdletBinding()]
param(
    [string] $Source,
    [Parameter(Mandatory = $true)][string] $Version,
    [switch] $IdentityOnly
)

$ErrorActionPreference = 'Stop'

# Stock Ollama's Inno Setup AppId. This is the key the Store correlates.
$StockKeyName = '{44E83376-CE68-45EB-8FC1-393500EB558C}_is1'
# Ours. A different GUID so the correlation cannot be made by product code
# either, should the manifest ever grow one.
$ForkKeyName = '{7B4D1A62-9F30-4E55-A0C7-5E1C0DE17B01}_is1'
$UninstallRoot = 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall'
$InstallDir = Join-Path $env:LOCALAPPDATA 'Programs\Ollama'
$BackupDir = 'C:\ollama-tb\arp-backup'

function Write-Step($msg) { Write-Host "==> $msg" }

# Refuse while anything is running: replacing a binary under a live process
# leaves the process on the old one and the disk on the new, which is exactly
# the confusion that made the eleven2go outage take a day to see.
$running = @(Get-Process -Name 'ollama', 'ollama app' -ErrorAction SilentlyContinue)
if ($running.Count -gt 0) {
    throw "Ollama is running (PID $($running.Id -join ', ')). Stop it first; a swap under a live process is not observable."
}

if (-not $IdentityOnly) {
    if (-not $Source) { throw 'Provide -Source, or pass -IdentityOnly.' }
    foreach ($name in @('ollama.exe', 'ollama app.exe')) {
        $src = Join-Path $Source $name
        if (-not (Test-Path $src)) { throw "missing from -Source: $name" }
    }
    # The tray app is part of the install, not an optional extra: it is what
    # runs the in-app update check, so shipping only ollama.exe leaves the old
    # app in place.
    Write-Step "installing binaries into $InstallDir"
    foreach ($name in @('ollama.exe', 'ollama app.exe')) {
        Copy-Item (Join-Path $Source $name) (Join-Path $InstallDir $name) -Force
        Write-Host "    $name"
    }
}

# Claim the identity.
$stockPath = Join-Path $UninstallRoot $StockKeyName
$forkPath = Join-Path $UninstallRoot $ForkKeyName

if (Test-Path $stockPath) {
    New-Item -ItemType Directory -Force -Path $BackupDir | Out-Null
    $stamp = Get-Date -Format 'yyyyMMdd-HHmmss'
    $backup = Join-Path $BackupDir "ollama-arp-$stamp.reg"
    & reg.exe export "HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\$StockKeyName" $backup /y | Out-Null
    if (-not (Test-Path $backup)) { throw "could not back up the ARP key; refusing to rewrite it" }
    Write-Step "backed up the stock ARP entry to $backup"

    if (Test-Path $forkPath) { Remove-Item $forkPath -Recurse -Force }
    Copy-Item $stockPath $forkPath -Force
    Remove-Item $stockPath -Recurse -Force
    Write-Step 'renamed the stock ARP entry to ours'
} elseif (-not (Test-Path $forkPath)) {
    # Nothing to rename and nothing of ours: the machine never had stock
    # installed, so there is no entry for the Store to match. Leave it that way
    # rather than inventing one.
    Write-Step 'no Ollama ARP entry on this machine; nothing for the Store to correlate'
}

if (Test-Path $forkPath) {
    Set-ItemProperty $forkPath -Name DisplayName    -Value 'Ollama think-budget'
    Set-ItemProperty $forkPath -Name Publisher      -Value 'mann1x'
    Set-ItemProperty $forkPath -Name DisplayVersion -Value $Version
    Write-Step "identity is now 'Ollama think-budget' / mann1x / $Version"
}

# Say whether it worked, from winget's own mouth rather than from ours.
Write-Step 'checking the Store can no longer correlate this install'
$listed = & winget list --id Ollama.Ollama --disable-interactivity 2>&1 | Out-String
if ($listed -match 'No installed package found') {
    Write-Host '    OK: winget reports no installed Ollama.Ollama'
} else {
    Write-Warning "winget still correlates this install; the Store can still replace it:`n$listed"
}
