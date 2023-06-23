#!/bin/sh
set -eux

readonly chain=LIMADNS

chain_exists() {
	iptables --table nat -n --list "${chain}" >/dev/null 2>&1
}
netns="$(ip netns identify $$)"

echo "netns: $netns"
if [ "$LIMA_VMTYPE" = "wsl" ] && [ "$netns" != "lima-$LIMA_CIDATA_NAME" ]; then
	echo "restarting script in 'lima-$LIMA_CIDATA_NAME' network namespace"
	ip netns add lima-wsl
	$(nsenter --net=/var/run/netns/lima-wsl $0) && exit
fi

# Wait until iptables has been installed; 35-configure-packages.sh will call this script again
if command -v iptables >/dev/null 2>&1; then
	if ! chain_exists; then
		iptables --table nat --new-chain ${chain}
		iptables --table nat --insert PREROUTING 1 --jump "${chain}"
		iptables --table nat --insert OUTPUT 1 --jump "${chain}"
	fi

	# Remove old rules
	iptables --table nat --flush ${chain}
	# Add rules for the existing ip:port
	iptables --table nat --append "${chain}" --destination "${LIMA_CIDATA_SLIRP_DNS}" --protocol udp --dport 53 --jump DNAT \
		--to-destination "${LIMA_CIDATA_SLIRP_GATEWAY}:${LIMA_CIDATA_UDP_DNS_LOCAL_PORT}"
	iptables --table nat --append "${chain}" --destination "${LIMA_CIDATA_SLIRP_DNS}" --protocol tcp --dport 53 --jump DNAT \
		--to-destination "${LIMA_CIDATA_SLIRP_GATEWAY}:${LIMA_CIDATA_TCP_DNS_LOCAL_PORT}"
fi
