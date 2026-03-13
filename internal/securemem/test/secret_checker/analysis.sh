#!/bin/sh

apk update >/dev/null 2>&1
apk add gdb >/dev/null 2>&1

UNEXPOSED_SECRET=MYSECRETKEY123458901234567890123
EXPOSED_SECRET=EXPOSED_SECRET123456789012345678

set +x
cd /app
echo "👷 Building secret-runner"
GOEXPERIMENT=runtimesecret go build -o ./tmp/secret-runner internal/securemem/test/secret_checker/main.go
echo "🚀 Starting secret-runner in the background..."
cd ./tmp
./secret-runner &

wait_for_file() {
  FILE=$1
  TIMEOUT=${2:-10} # default 10 seconds
  INTERVAL=${3:-2} # check every 2 second
  ELAPSED=0

  echo "⏳ Waiting for file: $FILE (timeout: ${TIMEOUT}s)"

  while [ ! -f "$FILE" ]; do
    if [ "$ELAPSED" -ge "$TIMEOUT" ]; then
      echo "❌ Timeout! File not found: $FILE"
      return 1
    fi
    sleep "$INTERVAL"
    ELAPSED=$((ELAPSED + INTERVAL))
    echo "   ... waiting ${ELAPSED}s"
  done

  echo "✅ File found: $FILE (after ${ELAPSED}s)"
  return 0
}

wait_for_file "start"
rm start

pid=$(pgrep secret-runner)
echo "PID: $pid"
echo "DUMPING CORE"
gcore -o core ${pid} >/dev/null 2>&1
echo " 🔎 SEARCHING FOR SECRET"
echo "-----------------------"
strings core.${pid} | grep $UNEXPOSED_SECRET | xargs -I {} echo "☣️ ALERT UNEXPOSED SECRET FOUND: {}"
strings core.${pid} | grep $EXPOSED_SECRET | xargs -I {} echo "✅ EXPOSED SECRET FOUND: {}"
echo "FINISHED"
rm core.${pid}
rm secret-runner
exit 0
