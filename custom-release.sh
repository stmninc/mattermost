#!/bin/bash

# 1. Accept two commit hashes or tags as arguments
# 2. Compare the difference between the two arguments and get a list of commit hashes
# 3. Cherry-pick the commits in order to the current branch
# 4. If a conflict occurs during cherry-pick, abort the process
# 5. Allow resuming cherry-pick from where it left off when the script is re-executed

set -euo pipefail

# Path to the state file
STATE_FILE=".custom_release_state"

# Display usage
usage() {
    echo "Usage: $0 <base_commit> <target_commit>"
    echo "  base_commit: Base commit hash or tag"
    echo "  target_commit: Target commit hash or tag"
    echo ""
    echo "Options:"
    echo "  --continue: Resume cherry-pick"
    echo "  --abort: Abort cherry-pick and clear state"
    exit 1
}

# Clear state
clean_state() {
    if [ -f "$STATE_FILE" ]; then
        rm -f "$STATE_FILE"
        echo "State file cleared"
    fi
}

# Abort cherry-pick
abort_cherry_pick() {
    echo "Aborting cherry-pick..."
    git cherry-pick --abort 2>/dev/null || true
    clean_state
    echo "Aborted"
    exit 0
}

# Check arguments
if [ "$1" == "--abort" ]; then
    abort_cherry_pick
fi

# Resume from where it left off
if [ "$1" == "--continue" ]; then
    if [ ! -f "$STATE_FILE" ]; then
        echo "Error: No resumable state found"
        exit 1
    fi
    
    # Check if there's an ongoing cherry-pick
    if [ -d ".git/sequencer" ] || [ -f ".git/CHERRY_PICK_HEAD" ]; then
        # Check if conflicts are resolved
        if [ -n "$(git ls-files -u)" ]; then
            echo "Error: Conflicts not resolved. You have unmerged files."
            echo "Please resolve conflicts and stage them with 'git add' before running --continue"
            exit 1
        fi
        
        # Continue cherry-pick
        echo "Continuing cherry-pick..."
        if git cherry-pick --continue; then
            echo "Cherry-pick completed"
        else
            echo "Error: Failed to continue cherry-pick"
            exit 1
        fi
    else
        echo "No ongoing cherry-pick found, resuming from next commit..."
    fi
    
    # Load state from file
    source "$STATE_FILE"
    RESUME=true
else
    # New execution
    if [ $# -ne 2 ]; then
        usage
    fi
    
    BASE_COMMIT="$1"
    TARGET_COMMIT="$2"
    
    # Verify commits exist
    if ! git rev-parse --verify "$BASE_COMMIT" >/dev/null 2>&1; then
        echo "Error: '$BASE_COMMIT' is not a valid commit or tag"
        exit 1
    fi
    
    if ! git rev-parse --verify "$TARGET_COMMIT" >/dev/null 2>&1; then
        echo "Error: '$TARGET_COMMIT' is not a valid commit or tag"
        exit 1
    fi
    
    # Warn if existing state file is found
    if [ -f "$STATE_FILE" ]; then
        echo "Warning: Incomplete cherry-pick process exists"
        echo "Use --continue to resume or --abort to cancel"
        exit 1
    fi
    
    # Get list of commit hashes including both base and target commits (oldest first)
    echo "Retrieving commit list..."
    COMMITS=$(git rev-list --reverse "$BASE_COMMIT^..$TARGET_COMMIT")
    
    if [ -z "$COMMITS" ]; then
        echo "No differences found"
        exit 0
    fi
    
    # Display commit count
    COMMIT_COUNT=$(echo "$COMMITS" | wc -l | tr -d ' ')
    echo "Number of commits to cherry-pick: $COMMIT_COUNT"
    
    # Save to state file
    declare -p BASE_COMMIT TARGET_COMMIT COMMITS > "$STATE_FILE"
    echo "CURRENT_INDEX=0" >> "$STATE_FILE"
    
    RESUME=false
fi

# Load state from file
source "$STATE_FILE"

# Cherry-pick commits one by one
INDEX=0
TOTAL=$(echo "$COMMITS" | wc -l | tr -d ' ')

for COMMIT in $COMMITS; do
    INDEX=$((INDEX + 1))
    
    # Skip already processed commits when resuming
    if [ "$RESUME" = true ] && [ $INDEX -le $CURRENT_INDEX ]; then
        continue
    fi
    
    echo ""
    echo "[$INDEX/$TOTAL] Cherry-picking: $COMMIT"
    
    # Display commit information
    git log --oneline -1 "$COMMIT"
    
    # Execute cherry-pick
    if git cherry-pick "$COMMIT"; then
        # Update state on success
        CURRENT_INDEX=$INDEX
        declare -p BASE_COMMIT TARGET_COMMIT COMMITS CURRENT_INDEX > "$STATE_FILE"
        echo "✓ Completed"
    else
        # When conflict occurs
        echo ""
        echo "================================"
        echo "Conflict occurred"
        echo "================================"
        echo ""
        echo "After resolving conflicts, run the following commands:"
        echo "  1. Resolve conflicts"
        echo "  2. git add <filename>"
        echo "  3. $0 --continue"
        echo ""
        echo "Or to abort:"
        echo "  $0 --abort"
        echo ""
        
        # Save state
        CURRENT_INDEX=$INDEX
        declare -p BASE_COMMIT TARGET_COMMIT COMMITS CURRENT_INDEX > "$STATE_FILE"
        
        exit 1
    fi
done

# Remove state file on complete success
clean_state
echo ""
echo "================================"
echo "All cherry-picks completed!"
echo "================================"
