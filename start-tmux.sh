#!/bin/bash

# Ensure the 'recordcleaner' session exists
if ! tmux has-session -t recordcleaner 2>/dev/null; then
  # Create a new session named 'recordcleaner', detached
  cd /workspaces/recordcleaner
  tmux new-session -d -s recordcleaner
  
  # Split the window horizontally (-h)
  # The left pane will remain a terminal
  # The right pane will run 'gh dash'
  tmux split-window -h -t recordcleaner "gh dash"
fi
