#!/bin/bash
cd ~/Desktop/projects/spotify/spotify
for var in "bungusbuster" "anneteresa-gb"
do 
	./spotify --u $var --r
done
