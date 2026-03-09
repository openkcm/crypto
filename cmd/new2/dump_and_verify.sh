# dump_and_verify.sh <PID>
# Forces a core dump of a running process, then checks that "abc" is absent.
set -euo pipefail

usage() {
  echo "Usage: $0 <PID>"
  exit 1
}
[ $# -eq 1 ] || usage
TARGET_PID=$1

COREFILE="core.$TARGET_PID"

# ── 1. Dump ──────────────────────────────────────────────────────────────────
echo "[*] Dumping PID $TARGET_PID -> $COREFILE ..."
ulimit -c unlimited
gcore -o "$COREFILE" "$TARGET_PID"

# gcore appends the PID, so the actual file is core.<pid>.<pid> on some systems
# or core.<pid> on others — find it.
ACTUAL=""
for candidate in "$COREFILE" "${COREFILE}.${TARGET_PID}"; do
  [ -f "$candidate" ] && {
    ACTUAL="$candidate"
    break
  }
done

[ -n "$ACTUAL" ] || {
  echo "ERROR: core file not found after gcore"
  exit 1
}
echo "[*] Core file: $ACTUAL ($(du -h "$ACTUAL" | cut -f1))"

# ── 2. Verify ────────────────────────────────────────────────────────────────
echo ""
INSECURE_HITS=$(strings "$ACTUAL" | grep -c "^thisIsInsecure$" || true)
SECURE_HITS=$(strings "$ACTUAL" | grep -c "^THISISSECURE$" || true)

if [ "$INSECURE_HITS" -gt 0 ]; then
  echo "[PASS] 'thisIsInsecure' found $INSECURE_HITS time(s) — unprotected memory is visible"
else
  echo "[FAIL] 'thisIsInsecure' NOT found — script may not be working correctly"
fi

if [ "$SECURE_HITS" -eq 0 ]; then
  echo "[PASS] 'THISISSECURE' NOT found — protected memory excluded from dump"
else
  echo "[FAIL] 'THISISSECURE' found $SECURE_HITS time(s) — MADV_DONTDUMP did not work"
fi
