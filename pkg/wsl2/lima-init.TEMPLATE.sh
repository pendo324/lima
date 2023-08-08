set -eu; \
export LOG_FILE=/var/log/lima-init.log; \
exec > >(tee \$LOG_FILE) 2>&1; \
export LIMA_CIDATA_MNT=$(/usr/bin/wslpath '{{.CIDataPath}}'); \
exec \$LIMA_CIDATA_MNT/boot.sh;
