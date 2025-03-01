#Requires -Version 5.1
$ErrorActionPreference = "Stop"

# Define constants
$ScriptRoot = $PSScriptRoot
$SrcDir = Join-Path $ScriptRoot "src\ktp2\cmd"
$Executable = Join-Path $ScriptRoot "ktoc.exe"
$LogFile = Join-Path $ScriptRoot "build.log"

# Function to log messages, supporting pipeline input and optional message
function Write-Log {
    param (
        [Parameter(Mandatory = $false, ValueFromPipeline = $true)]
        [string]$Message = "",
        [switch]$Error
    )
    process {
        $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
        $logMessage = "[$timestamp] $Message"
        if ($Error) {
            Write-Host $logMessage -ForegroundColor Red
        } else {
            Write-Host $logMessage
        }
        $logMessage | Out-File -FilePath $LogFile -Append
    }
}

# Function to handle errors and ensure return to original directory
function Handle-Error {
    param (
        [Parameter(Mandatory = $true)]
        [string]$ErrorMessage
    )
    Write-Log -Error -Message "ERROR: $ErrorMessage"
    try {
        Set-Location $ScriptRoot -ErrorAction Stop
    } catch {
        Write-Log -Error -Message "Failed to return to original directory: $_"
    }
    exit 1
}

# Check for Go installation
if (-not (Get-Command "go" -ErrorAction SilentlyContinue)) {
    Handle-Error "Go is not installed. Please install it first."
}

# Trap unexpected errors to ensure cleanup
try {
    # Start build process
    Write-Log -Message "Starting build process in $ScriptRoot"

    # Change to source directory
    Set-Location $SrcDir
    Write-Log -Message "Changed to directory: $SrcDir"

    # Pull Go dependencies
    Write-Log -Message "Pulling Go dependencies..."
    $output = & go mod download 2>&1 | Out-String
    if ($LASTEXITCODE -ne 0) { Handle-Error "Failed to download dependencies: $output" }
    if ($output.Trim()) { Write-Log -Message $output.Trim() } else { Write-Log -Message "No new dependencies downloaded" }

    $output = & go get ktp2/src/ktp2/cmd ktp2/src/ktp2/tests 2>&1 | Out-String
    if ($LASTEXITCODE -ne 0) { Handle-Error "Failed to get cmd and tests dependencies: $output" }
    if ($output.Trim()) { Write-Log -Message $output.Trim() } else { Write-Log -Message "No updates for cmd/tests dependencies" }

    $output = & go get -t ktp2/src/ktp2/ktfunc 2>&1 | Out-String
    if ($LASTEXITCODE -ne 0) { Handle-Error "Failed to get ktfunc dependencies: $output" }
    if ($output.Trim()) { Write-Log -Message $output.Trim() } else { Write-Log -Message "No updates for ktfunc dependencies" }

    # Build executable
    Write-Log -Message "Building executable..."
    $output = & go build -o $Executable 2>&1 | Out-String
    if ($LASTEXITCODE -ne 0) { Handle-Error "Build failed: $output" }
    if ($output.Trim()) { Write-Log -Message $output.Trim() } else { Write-Log -Message "Build completed with no additional output" }

    # Run tests
    Write-Log -Message "Running tests..."
    Set-Location (Split-Path $SrcDir -Parent)
    $output = & go test ./... -v -cover -coverprofile=coverage.out 2>&1 | Out-String
    if ($LASTEXITCODE -ne 0) { Handle-Error "Tests failed: $output" }
    if ($output.Trim()) { Write-Log -Message $output.Trim() }

    # Write-Log -Message "Coverage..."
    # $output = & go tool cover -func=coverage.out 2>&1 | Out-String
    # Write-Log -Message "Writing logs... "
    if ($output.Trim()) { Write-Log -Message $output.Trim() }

    Write-Log -Message "Return home..."
    Set-Location $ScriptRoot
    Write-Log -Message "Build process completed successfully"
    Write-Log -Message "Executable built at: $Executable"
}
catch {
    Handle-Error "Unexpected error occurred: $_"
}
finally {
    # Ensure we're back in the original directory even if an uncaught exception occurs
    if ((Get-Location).Path -ne $ScriptRoot) {
        try {
            Set-Location $ScriptRoot -ErrorAction Stop
            Write-Log -Message "Returned to original directory after error"
        } catch {
            Write-Log -Error -Message "Failed to return to original directory in finally block: $_"
        }
    }
}

exit 0