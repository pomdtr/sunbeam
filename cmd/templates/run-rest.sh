#!/bin/bash

POSITIONAL_ARGS=()
QUERY_PARAMS=()

while [[ $# -gt 0 ]]; do
  case $1 in
    --*=*) # for args with =, like --foo=bar
      QUERY_PARAMS+=("${1/--/}")
      shift # past argument=value
      ;;
    -*=*) # for args with =, like -f=bar
      QUERY_PARAMS+=("${1/-/}")
      shift # past argument=value
      ;;
    --*)
      if [ -n "$2" ] && [ "${2:0:1}" != "-" ]; then
        QUERY_PARAMS+=("${1/--/}=$2")
        shift # past argument
        shift # past value
      else
        QUERY_PARAMS+=("${1/--/}=true")
        shift # past argument
      fi
      ;;
    -*)
        if [ -n "$2" ] && [ "${2:0:1}" != "-" ]; then
            QUERY_PARAMS+=("${1/-/}=$2")
            shift # past argument
            shift # past value
        else
            QUERY_PARAMS+=("${1/-/}=true")
            shift # past argument
        fi
        ;;
    *)
      ENDPOINT="$ENDPOINT/$1"
      POSITIONAL_ARGS+=("$1") # save positional arg
      shift # past argument
      ;;
  esac
done

ENDPOINT=""
for i in "${POSITIONAL_ARGS[@]}"; do
    if [ -n "$ENDPOINT" ]; then
        ENDPOINT+="/$i"
    else
        ENDPOINT="$i"
    fi

done

QUERY=""
for i in "${QUERY_PARAMS[@]}"; do
    if [ -n "$QUERY" ]; then
        QUERY+="&$i"
    else
        QUERY="$i"
    fi
done

sunbeam fetch -X POST "{{ .Remote }}$ENDPOINT?$QUERY"
