#!/bin/bash

# Exit on error, unset variables, and pipe failures
set -euo pipefail

# Constants
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SRC_DIR="${SCRIPT_DIR}/src/ktp2/cmd"
readonly EXECUTABLE="${SCRIPT_DIR}/ktoc"
readonly LOG_FILE="${SCRIPT_DIR}/build.log"

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly NC='\033[0m' # No Color

# Functions
log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[${timestamp}] ${level}: ${message}" | tee -a "$LOG_FILE"
}

error() {
    log "ERROR" "$@" >&2
    cd "$SCRIPT_DIR" || true  # Attempt to return to original dir, ignore failure
    exit 1
}

info() {
    log "INFO" "$@"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*" | tee -a "$LOG_FILE"
}

# Check for required tools
check_dependencies() {
    command -v go >/dev/null 2>&1 || error "Go is not installed. Please install it first."
}

# Main build process
main() {
    # Initialize log file
    echo "Build started at $(date)" > "$LOG_FILE"
    info "Starting build process in ${SCRIPT_DIR}"

    # Change to source directory
    cd "$SRC_DIR" || error "Failed to change to source directory: ${SRC_DIR}"

    info "Pulling Go dependencies..."
    go mod download || error "Failed to download dependencies"
    go get ktp2/src/ktp2/cmd ktp2/src/ktp2/tests || error "Failed to get cmd and tests dependencies"
    go get -t ktp2/src/ktp2/ktfunc || error "Failed to get ktfunc dependencies"

    info "Building executable..."
    go build -o "$EXECUTABLE" || error "Build failed"

    info "Running tests..."
    cd .. || error "Failed to change directory for tests"
    go test ./... -v -cover -coverprofile=coverage.out || error "Tests failed"
    go tool cover -func=coverage.out | tee -a "$LOG_FILE" || info "Coverage report generation failed, continuing..."

    # Return to original directory
    cd "$SCRIPT_DIR" || error "Failed to return to original directory"

    success "Build process completed successfully"
    info "Executable built at: ${EXECUTABLE}"
}

# Trap errors to ensure cleanup
trap 'error "Script terminated unexpectedly at line ${LINENO}"' ERR

# Check dependencies and run main
check_dependencies
main "$@"  # Pass any command-line arguments (currently unused)

exit 0