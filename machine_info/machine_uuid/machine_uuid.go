package machine_uuid

import (
	"SuperNet-Node/utils/log_utils"
	"crypto/sha256"
	"encoding/hex"
)

type MachineUUID string

func GetInfoMachineUUID(cpu, gpu, ip, country, region, city string) (MachineUUID, error) {
	logs.Normal("Getting machine ID...")
	hash := sha256.Sum256([]byte(cpu + gpu + ip + country + region + city))
	return MachineUUID(hex.EncodeToString(hash[:])[:32]), nil
}
