#!/bin/sh

apk update >/dev/null 2>&1
apk add gdb >/dev/null 2>&1

set +x
cd /app
echo "👷 Building secret-runner"
GOEXPERIMENT=runtimesecret go build -o secret-runner internal/securemem/test/secret_checker/main.go
echo "🚀 Starting secret-runner in the background..."
./secret-runner &

wait_for_file() {
  FILE=$1
  TIMEOUT=${2:-60} # default 60 seconds
  INTERVAL=${3:-2} # check every 1 second
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
strings core.${pid} | grep MYSECRETKEY123458901234567890123 | xargs -I {} echo "☣️ ALERT DANGER FOUND: {}"
echo "FINISHED"
rm core.${pid}
rm secret-runner
exit 0
