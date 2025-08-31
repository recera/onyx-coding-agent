#!/bin/bash

# Onyx AI TUI Global Installation Script

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Onyx AI Coding TUI - Global Installation${NC}"
echo "============================================"
echo ""

# Get the directory of this script (where the TUI code is)
INSTALL_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Define installation paths
INSTALL_PREFIX="${HOME}/.local"
BIN_DIR="${INSTALL_PREFIX}/bin"
LIB_DIR="${INSTALL_PREFIX}/lib/onyx-tui"

# Create directories if they don't exist
echo -e "${YELLOW}📁 Creating installation directories...${NC}"
mkdir -p "$BIN_DIR"
mkdir -p "$LIB_DIR"

# Check dependencies
echo -e "${YELLOW}🔍 Checking dependencies...${NC}"

if ! command -v node &> /dev/null; then
    echo -e "${RED}❌ Node.js is not installed. Please install Node.js first.${NC}"
    echo "   Visit: https://nodejs.org/"
    exit 1
fi

if ! command -v npm &> /dev/null; then
    echo -e "${RED}❌ npm is not installed. Please install npm first.${NC}"
    exit 1
fi

if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed. Please install Go first.${NC}"
    echo "   Visit: https://golang.org/"
    exit 1
fi

NODE_VERSION=$(node --version)
GO_VERSION=$(go version | cut -d' ' -f3)
echo -e "${GREEN}✓${NC} Node.js version: $NODE_VERSION"
echo -e "${GREEN}✓${NC} Go version: $GO_VERSION"
echo ""

# Install Node dependencies
echo -e "${YELLOW}📦 Installing Node.js dependencies...${NC}"
cd "$INSTALL_DIR/agent"
if ! npm install; then
    echo -e "${RED}❌ Failed to install Node dependencies${NC}"
    exit 1
fi
echo -e "${GREEN}✓${NC} Node dependencies installed"

# Build the TypeScript agent
echo -e "${YELLOW}🔨 Building TypeScript agent...${NC}"
if ! npm run build; then
    echo -e "${RED}❌ Failed to build TypeScript agent${NC}"
    exit 1
fi
echo -e "${GREEN}✓${NC} TypeScript agent built successfully"

# Build the Go TUI
echo -e "${YELLOW}🔨 Building Go TUI...${NC}"
cd "$INSTALL_DIR"
if ! go build -o onyx-tui-bin .; then
    echo -e "${RED}❌ Failed to build Go TUI${NC}"
    exit 1
fi
echo -e "${GREEN}✓${NC} TUI built successfully"

# Copy files to installation directory
echo -e "${YELLOW}📋 Installing files...${NC}"
cp -r "$INSTALL_DIR"/* "$LIB_DIR/"
echo -e "${GREEN}✓${NC} Files copied to $LIB_DIR"

# Create the launcher script
echo -e "${YELLOW}✨ Creating launcher script...${NC}"
cat > "$BIN_DIR/onyx" << 'EOF'
#!/bin/bash

# Onyx AI TUI Launcher
# This script can be run from any directory

# Get the current working directory (where user runs the command)
WORK_DIR="$(pwd)"

# Installation directory
LIB_DIR="${HOME}/.local/lib/onyx-tui"

# Check if installation exists
if [ ! -d "$LIB_DIR" ]; then
    echo "❌ Onyx TUI not installed. Please run the install script first."
    exit 1
fi

# Check for API key
if [ -z "$OPENAI_API_KEY" ]; then
    echo "ℹ️  Note: OPENAI_API_KEY not found in environment."
    echo "   You'll be prompted to enter it when the TUI starts."
    echo ""
fi

# Change to lib directory to run the TUI (so it can find the agent)
cd "$LIB_DIR"

# But set the working directory environment variable so the TUI knows where to operate
export ONYX_WORK_DIR="$WORK_DIR"

# Run the TUI
exec ./onyx-tui-bin "$@"
EOF

chmod +x "$BIN_DIR/onyx"
echo -e "${GREEN}✓${NC} Launcher created at $BIN_DIR/onyx"

# Add to PATH if not already there
if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
    echo ""
    echo -e "${YELLOW}📝 Add this to your shell configuration (.bashrc, .zshrc, etc.):${NC}"
    echo ""
    echo "    export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo ""
    echo "Then reload your shell or run:"
    echo "    source ~/.bashrc  # or ~/.zshrc"
else
    echo -e "${GREEN}✓${NC} $BIN_DIR is already in PATH"
fi

echo ""
echo -e "${GREEN}✅ Installation complete!${NC}"
echo ""
echo "To use Onyx AI TUI from any directory:"
echo "  1. Make sure ~/.local/bin is in your PATH"
echo "  2. Run: ${GREEN}onyx${NC}"
echo ""
echo "The TUI will operate on files in your current directory."