#!/bin/bash

save_files=true

while [[ $# -gt 0 ]]; do
    case "$1" in
        -d|--data)
            save_files=false
            shift  
            ;;
        *)
            echo "unknown param '$1'"
            shift
            ;;
    esac
done

rm -rf ./par_data/* ./seq_data/* ./temp/* ./results/* ./*.json ./*.tmp
rm -rf ./ckks/*.bin ./ckks/*.csv ./ckks/*.txt ./ckks/vbuff/*

if [ "$save_files" = false ]; then
    rm -rf ./*.csv ./*.hdf
    echo "Cleared."
else
    echo "Cleared, but save data."
fi