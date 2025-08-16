#!/usr/bin/env sh
if [ "$1" = "" ]; then
				go run main.go
elif [ "$1" = "watch" ]; then
				echo "running with auto-restart .."
				echo
				echo -e "main.go\ntemplate.html" | entr -r go run main.go
else
				echo "invalid argument: $1"
				echo "usage:"
				echo "	$0 [watch]"
fi;
