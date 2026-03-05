#!/bin/sh
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "${GREEN}======================================${NC}"
echo "${GREEN}  Starting GoCore Application${NC}"
echo "${GREEN}======================================${NC}"
echo ""

# Function to run database migrations
run_migrations() {
    echo ""
    echo "${YELLOW}Running database migrations...${NC}"
    
    if /app/migrate up; then
        echo "${GREEN}✓ Migrations completed successfully${NC}"
        return 0
    else
        echo "${RED}✗ Migration failed${NC}"
        return 1
    fi
}

# Function to display migration status
show_migration_status() {
    echo ""
    echo "${YELLOW}Current migration status:${NC}"
    /app/migrate status || true
    echo ""
}

# Main execution
main() {
    # Run migrations
    if ! run_migrations; then
        echo "${RED}Exiting due to migration failure${NC}"
        exit 1
    fi
    
    # Show migration status
    show_migration_status
    
    # Start the application
    echo "${GREEN}======================================${NC}"
    echo "${GREEN}  Starting Application Server${NC}"
    echo "${GREEN}======================================${NC}"
    echo ""
    
    # Execute the main command (passed as arguments)
    exec "$@"
}

# Run main function
main "$@"
