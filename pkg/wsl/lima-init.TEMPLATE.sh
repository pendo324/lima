set -eu

export LIMA_CIDATA_MNT=$(/usr/bin/wslpath .CIDataPath)
LOG_FILE=/var/log/lima-init.log
exec > >(tee $LOG_FILE) 2>&1
exec "${LIMA_CIDATA_MNT}"/boot.sh
