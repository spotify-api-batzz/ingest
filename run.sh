#!/bin/bash
cd ~/Desktop/projects/spotify/spotify
if [ ! -f .env ]
then
  export $(cat .env | xargs)
fi



IFS=',' read -ra ADDR <<< "${users}"
for var in "${ADDR[@]}"; do
  # process "$i"
done
do 
	./spotify --u $var --a --r --t
done
