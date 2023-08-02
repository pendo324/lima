#!/bin/sh
# This script replaces the cloud-init functionality of creating a user and setting its SSH keys
# when using a WSL2 VM.
[ "$LIMA_VMTYPE" = "wsl" ] || exit 0

# create user
sudo useradd -u "${LIMA_CIDATA_UID}" "${LIMA_CIDATA_USER}" -d /home/"${LIMA_CIDATA_USER}".linux/
sudo mkdir /home/"${LIMA_CIDATA_USER}".linux/.ssh/
sudo cp "${LIMA_CIDATA_MNT}"/ssh_authorized_keys /home/"${LIMA_CIDATA_USER}".linux/.ssh/authorized_keys
sudo chown "${LIMA_CIDATA_USER}" /home/"${LIMA_CIDATA_USER}".linux/.ssh/authorized_keys

# add lima to sudoers
sudo echo "lima ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/99_lima_sudoers

# copy some CIDATA to the hardcoded path for requirement checks (TODO: make this not hardcoded)
sudo mkdir -p /mnt/lima-cidata
sudo cp "${LIMA_CIDATA_MNT}"/meta-data /mnt/lima-cidata/meta-data