#!/bin/bash
cd ~/spotify/ingest
source .env
IFS=',' read -ra ADDR <<< "${users}"
for var in "${ADDR[@]}"; do
	./spotify --u $var --a --r --t
done
