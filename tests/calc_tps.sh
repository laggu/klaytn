#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
BENCHNAME="ValueTransfer"

BENCHTIME=${BENCHTIME:-"5s"}
BENCHCOUNT=${BENCHCOUNT:-5}

TMP=`mktemp`

cd $DIR/../tests

CMD="go test -run X -bench $BENCHNAME -benchtime $BENCHTIME -count $BENCHCOUNT"
echo "executing $CMD"
$CMD | tee $TMP
NS=`grep "ns/op" $TMP | awk 'BEGIN{total=0.0;count=0} {total+=$3;count++} END{printf("%f", total/count)}'`
TPS=$(echo "1.0 / $NS * 1000.0 * 1000.0 * 1000" | bc -l)
echo "TPS for a single machine = $TPS"
rm -rf $TMP
