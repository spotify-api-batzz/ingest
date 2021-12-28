#!/bin/bash
cd ~/Desktop/projects/spotify/spotify
if [ ! -f .env ]
then
  export $(cat .env | xargs)
fi



for var in "bungusbuster" "anneteresa-gb" "tomadams1997"
IFS=',' read -ra ADDR <<< "${users}"
for var in "${ADDR[@]}"; do
  # process "$i"
done
do 
	./spotify --u $var --a --r --t
done
