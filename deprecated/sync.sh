# This is for osx. Brew install gtimeout!
tmux new -d -s rsync bash -c 'while true; do gtimeout 10 rsync -azP --delete fuzzer athena:~; sleep 1; done'
