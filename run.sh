#!/bin/bash
cd ~/Desktop/projects/spotify/spotify
source .env
IFS=',' read -ra ADDR <<< "${users}"
for var in "${ADDR[@]}"; do
do 
	./spotify --u $var --a --r --t
done
