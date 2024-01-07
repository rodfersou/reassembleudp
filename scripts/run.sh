#!/bin/bash
direnv exec $PWD \
    tmux new-session '
             ./scripts/run_reassembleudp.sh
         ' \; \
         split-window -h '
             echo Press return to start &&
             read                       &&
             ./scripts/run_emitter.sh   &&
             read
         '
