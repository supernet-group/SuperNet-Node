package utils

import (
	"SuperNet-Node/machine_info/machine_uuid"
	"SuperNet-Node/pattern"
	"encoding/hex"

	"github.com/gagliardetto/solana-go"
)

func GenMachine(machineOwner solana.PublicKey, machineUUID machine_uuid.MachineUUID) [][]byte {
	
	// Convert machineUUID from machine_uuid.MachineUUID to pattern.MachineUUID
	var byteUUID pattern.MachineUUID
	b, err := hex.DecodeString(string(machineUUID))
	if err != nil {
		panic(err)
	}
	copy(byteUUID[:], b[:16])

	// Create seedMachine, a 2D slice of bytes
	seedMachine := [][]byte{
		[]byte("machine"),
		machineOwner.Bytes(),
		[]byte(byteUUID[:]),
	}
	return seedMachine
}

func GenVault() [][]byte {

	ecpc := solana.MustPublicKeyFromBase58(pattern.DIST_TOKEN_ID)
	seedVault := [][]byte{
		[]byte("vault"),
		ecpc.Bytes(),
	}
	return seedVault
}

func GenTask(machineOwner solana.PublicKey, TaskUUID pattern.TaskUUID) [][]byte {

	seedTask := [][]byte{
		[]byte("task"),
		machineOwner.Bytes(),
		[]byte(TaskUUID[:]),
	}
	return seedTask
}

func GenReward() [][]byte {
	seedReward := [][]byte{
		[]byte("reward"),
		PeriodBytes(),
	}
	return seedReward
}

func GenRewardMachine(machineOwner solana.PublicKey, machineUUID pattern.MachineUUID) [][]byte {
	seedRewardMachine := [][]byte{
		[]byte("reward-machine"),
		PeriodBytes(),
		machineOwner.Bytes(),
		[]byte(machineUUID[:]),
	}
	return seedRewardMachine
}
