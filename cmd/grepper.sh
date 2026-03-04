docker exec -it alpine sh
apk add gdb
rm core.1
gcore 1
echo "SEARCHIN"
strings core.1 | grep passphrasewhichneedstobe32bytes
echo "FINISHED"
