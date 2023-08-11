//go:build windows
// +build windows

package windows

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
	"golang.org/x/sys/windows/registry"
)

const (
	guestCommunicationsPrefix = `Computer\HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices\`
	magicVSOCKSuffix          = "-facb-11e6-bd58-64006a7986d3"
)

func AddVSockRegistryKey(port int) error {
	rootKey, err := getGuestCommunicationServicesKey()
	if err != nil {
		return err
	}

	used, err := getUsedPorts(rootKey)
	if err != nil {
		return err
	}

	if slices.Contains(used, port) {
		return fmt.Errorf("port %q in use", port)
	}

	vsockKeyPath := fmt.Sprintf(`%x%s`, port, magicVSOCKSuffix)
	vSockKey, _, err := registry.CreateKey(
		rootKey,
		vsockKeyPath,
		registry.ALL_ACCESS,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to create new key (%s%s): %w",
			guestCommunicationsPrefix,
			vsockKeyPath,
			err,
		)
	}
	defer vSockKey.Close()

	return nil
}

func RemoveVSockRegistryKey(port int) error {
	rootKey, err := getGuestCommunicationServicesKey()
	if err != nil {
		return err
	}

	vsockKeyPath := fmt.Sprintf(`%x%s`, port, magicVSOCKSuffix)
	if err := registry.DeleteKey(rootKey, vsockKeyPath); err != nil {
		return fmt.Errorf(
			"failed to create new key (%s%s): %w",
			guestCommunicationsPrefix,
			vsockKeyPath,
			err,
		)
	}

	return nil
}

func IsPortFree(port int) (bool, error) {
	rootKey, err := getGuestCommunicationServicesKey()
	if err != nil {
		return false, err
	}

	used, err := getUsedPorts(rootKey)
	if err != nil {
		return false, err
	}

	if slices.Contains(used, port) {
		return false, nil
	}

	return true, nil
}

func getGuestCommunicationServicesKey() (registry.Key, error) {
	rootKey, err := registry.OpenKey(registry.LOCAL_MACHINE, guestCommunicationsPrefix, registry.QUERY_VALUE)
	if err != nil {
		return 0, fmt.Errorf(
			"failed to open GuestCommunicationServices key (%s): %w",
			guestCommunicationsPrefix,
			err,
		)
	}
	defer rootKey.Close()

	return rootKey, nil
}

func getUsedPorts(key registry.Key) ([]int, error) {
	keys, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return nil, fmt.Errorf("failed to read subkey names for %s: %w", guestCommunicationsPrefix, err)
	}

	out := []int{}
	for _, k := range keys {
		split := strings.Split(k, magicVSOCKSuffix)
		if len(split) == 2 {
			i, err := strconv.Atoi(split[0])
			if err != nil {
				return nil, fmt.Errorf("failed convert %q to int: %w", split[0], err)
			}
			out = append(out, i)
		}
	}

	return out, nil
}

func GetRandomFreePort(min, max int) (int, error) {
	rootKey, err := getGuestCommunicationServicesKey()
	if err != nil {
		return 0, err
	}

	used, err := getUsedPorts(rootKey)
	if err != nil {
		return 0, err
	}

	type pair struct{ v, offset int }
	tree := make([]pair, 1, len(used)+1)
	tree[0] = pair{0, min}

	sort.Ints(used)
	for i, v := range used {
		if tree[len(tree)-1].v+tree[len(tree)-1].offset == v {
			tree[len(tree)-1].offset++
		} else {
			tree = append(tree, pair{v - min - i, min + i + 1})
		}
	}

	v := rand.Intn(max - min + 1 - len(used))

	for len(tree) > 1 {
		m := len(tree) / 2
		if v < tree[m].v {
			tree = tree[:m]
		} else {
			tree = tree[m:]
		}
	}

	return tree[0].offset + v, nil
}
