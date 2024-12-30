package machine_uuid

import (
	"SuperNet-Node/utils/log_utils"
	"crypto/sha256"
	"encoding/hex"
)

type MachineUUID string

func GetInfoMachineUUID(cpu, gpu, ip, country, region, city string) (MachineUUID, error) {
	logs.Normal("Getting machine ID...")

	// Easy debugging
	// mID, err := os.ReadFile("/etc/machine-id")
	// if err != nil {
	// 	return "", fmt.Errorf("> ReadFile: %v", err)
	// }
	// id := strings.TrimSpace(string(mID))
	// // id = "ec25eab7e9b2f62b32f2ac249a0fd3b0"
	// return MachineUUID(id), nil

	hash := sha256.Sum256([]byte(cpu + gpu + ip + country + region + city))
    return MachineUUID(hex.EncodeToString(hash[:])[:32]), nil
}
