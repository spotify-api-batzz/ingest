#!/bin/bash
cd ~/spotify/ingest
# up until 11/2/2023 this was running recent ingest only for batu and gareth.. retard
source .env
IFS=',' read -ra ADDR <<< "${users}"
for var in "${ADDR[@]}"; do
        ./spotify --u $var --r
done
