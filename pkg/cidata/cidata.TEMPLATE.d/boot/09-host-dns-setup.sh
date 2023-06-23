#!/bin/sh
set -eux

readonly chain=LIMADNS

chain_exists() {
	iptables --table nat -n --list "${chain}" >/dev/null 2>&1
}

netns="$(ip netns identify $$)"

if [ "$LIMA_VMTYPE" = "wsl" ] && [ ! -f "/var/run/netns/lima-wsl" ]; then
	ip netns delete lima-wsl || true

	ip netns add lima-wsl
	ip link add veth-default type veth peer name veth-lima-wsl
	ip link set veth-lima-wsl netns lima-wsl
	ip addr add 10.0.3.1/24 dev veth-default
	ip netns exec lima-wsl ip addr add 10.0.3.2/24 dev veth-lima-wsl
	ip link set veth-default up
	ip netns exec lima-wsl ip link set veth-lima-wsl up
	ip netns exec lima-wsl ip route add default via 10.0.3.1
	echo 1 > /proc/sys/net/ipv4/ip_forward
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
	if [ "$LIMA_VMTYPE" = "wsl" ]; then
		iptables -A FORWARD -o eth0 -i veth-default -j ACCEPT
		iptables -A FORWARD -i eth0 -o veth-default -j ACCEPT
		iptables -t nat -A POSTROUTING -s 10.0.3.2/24 -o eth0 -j MASQUERADE
	fi
fi
