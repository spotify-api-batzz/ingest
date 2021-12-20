#!/bin/bash
cd ~/Desktop/projects/spotify/spotify
for var in "bungusbuster" "anneteresa-gb" "tomadams1997"
do 
	./spotify --u $var --a --r --t
done
