package cgo

/*
#cgo LDFLAGS: -L${SRCDIR}/..
#cgo pkg-config: ${SRCDIR}/../filcrypto.pc
#include "../filcrypto.h"
#include <stdlib.h>
*/
import "C"

func VerifySeal(registeredProof RegisteredSealProof, commR *ByteArray32, commD *ByteArray32, proverId *ByteArray32, ticket *ByteArray32, seed *ByteArray32, sectorId uint64, proof SliceRefUint8) (bool, error) {
	resp := C.verify_seal(registeredProof, commR, commD, proverId, ticket, seed, C.uint64_t(sectorId), proof)
	defer resp.Destroy()

	if err := CheckErr(resp); err != nil {
		return false, err
	}

	return bool(resp.value), nil
}

func VerifyAggregateSealProof(registeredProof RegisteredSealProof, registeredAggregation RegisteredAggregationProof, proverId *ByteArray32, proof SliceRefUint8, commitInputs SliceRefAggregationInputs) (bool, error) {
	resp := C.verify_aggregate_seal_proof(registeredProof, registeredAggregation, proverId, proof, commitInputs)
	defer resp.Destroy()

	if err := CheckErr(resp); err != nil {
		return false, err
	}
	return bool(resp.value), nil
}

func VerifyWinningPoSt(randomness *ByteArray32, replicas SliceRefPublicReplicaInfo, proofs SliceRefPoStProof, proverId *ByteArray32) (bool, error) {
	resp := C.verify_winning_post(randomness, replicas, proofs, proverId)
	defer resp.Destroy()

	if err := CheckErr(resp); err != nil {
		return false, err
	}

	return bool(resp.value), nil
}

func VerifyWindowPoSt(randomness *ByteArray32, replicas SliceRefPublicReplicaInfo, proofs SliceRefPoStProof, proverId *ByteArray32) (bool, error) {
	resp := C.verify_window_post(randomness, replicas, proofs, proverId)
	defer resp.Destroy()

	if err := CheckErr(resp); err != nil {
		return false, err
	}

	return bool(resp.value), nil
}

func GeneratePieceCommitment(registeredProof RegisteredSealProof, pieceFdRaw int32, unpaddedPieceSize uint64) ([]byte, error) {
	resp := C.generate_piece_commitment(registeredProof, C.int32_t(pieceFdRaw), C.uint64_t(unpaddedPieceSize))
	defer resp.Destroy()

	if err := CheckErr(resp); err != nil {
		return nil, err
	}

	return resp.value.comm_p.Copy(), nil
}

func GenerateDataCommitment(registeredProof RegisteredSealProof, pieces SliceRefPublicPieceInfo) ([]byte, error) {
	resp := C.generate_data_commitment(registeredProof, pieces)
	defer resp.Destroy()

	if err := CheckErr(resp); err != nil {
		return nil, err
	}

	return resp.value.Copy(), nil
}

func WriteWithAlignment(registeredProof RegisteredSealProof, srcFd int32, srcSize uint64, dstFd int32, existingPieceSizes SliceRefUint64) (uint64, uint64, []byte, error) {
	resp := C.write_with_alignment(registeredProof, C.int32_t(srcFd), C.uint64_t(srcSize), C.int32_t(dstFd), existingPieceSizes)
	defer resp.Destroy()
	if err := CheckErr(resp); err != nil {
		return 0, 0, nil, err
	}

	return uint64(resp.value.left_alignment_unpadded), uint64(resp.value.total_write_unpadded), resp.value.comm_p.Copy(), nil
}

func WriteWithoutAlignment(registeredProof RegisteredSealProof, srcFd int32, srcSize uint64, dstFd int32) (uint64, []byte, error) {
	resp := C.write_without_alignment(registeredProof, C.int32_t(srcFd), C.uint64_t(srcSize), C.int32_t(dstFd))
	defer resp.Destroy()
	if err := CheckErr(resp); err != nil {
		return 0, nil, err
	}

	return uint64(resp.value.total_write_unpadded), resp.value.comm_p.Copy(), nil
}

func SealPreCommitPhase1(registeredProof RegisteredSealProof, cacheDirPath SliceRefUint8, stagedSectorPath SliceRefUint8, sealedSectorPath SliceRefUint8, sectorId uint64, proverId *ByteArray32, ticket *ByteArray32, pieces SliceRefPublicPieceInfo) ([]byte, error) {
	resp := C.seal_pre_commit_phase1(registeredProof, cacheDirPath, stagedSectorPath, sealedSectorPath, C.uint64_t(sectorId), proverId, ticket, pieces)
	defer resp.Destroy()
	if err := CheckErr(resp); err != nil {
		return nil, err
	}
	return resp.value.Copy(), nil
}

func SealPreCommitPhase2(sealPreCommitPhase1Output SliceRefUint8, cacheDirPath SliceRefUint8, sealedSectorPath SliceRefUint8) ([]byte, []byte, error) {
	resp := C.seal_pre_commit_phase2(sealPreCommitPhase1Output, cacheDirPath, sealedSectorPath)
	defer resp.Destroy()
	if err := CheckErr(resp); err != nil {
		return nil, nil, err
	}
	return resp.value.comm_r.Copy(), resp.value.comm_d.Copy(), nil
}

func SealCommitPhase1(registeredProof RegisteredSealProof, commR *ByteArray32, commD *ByteArray32, cacheDirPath SliceRefUint8, replicaPath SliceRefUint8, sectorId uint64, proverId *ByteArray32, ticket *ByteArray32, seed *ByteArray32, pieces SliceRefPublicPieceInfo) ([]byte, error) {
	resp := C.seal_commit_phase1(registeredProof, commR, commD, cacheDirPath, replicaPath, C.uint64_t(sectorId), proverId, ticket, seed, pieces)
	defer resp.Destroy()
	if err := CheckErr(resp); err != nil {
		return nil, err
	}

	return resp.value.Copy(), nil
}

func SealCommitPhase2(sealCommitPhase1Output SliceRefUint8, sectorId uint64, proverId *ByteArray32) ([]byte, error) {
	resp := C.seal_commit_phase2(sealCommitPhase1Output, C.uint64_t(sectorId), proverId)
	defer resp.Destroy()
	if err := CheckErr(resp); err != nil {
		return nil, err
	}
	return resp.value.Copy(), nil
}

func AggregateSealProofs(registeredProof RegisteredSealProof, registeredAggregation RegisteredAggregationProof, commRs SliceRefByteArray32, seeds SliceRefByteArray32, sealCommitResponses SliceRefSliceBoxedUint8) ([]byte, error) {
	resp := C.aggregate_seal_proofs(registeredProof, registeredAggregation, commRs, seeds, sealCommitResponses)
	defer resp.Destroy()
	if err := CheckErr(resp); err != nil {
		return nil, err
	}
	return resp.value.Copy(), nil
}

func UnsealRange(registeredProof RegisteredSealProof, cacheDirPath SliceRefUint8, sealedSectorFdRaw int32, unsealOutputFdRaw int32, sectorId uint64, proverId *ByteArray32, ticket *ByteArray32, commD *ByteArray32, unpaddedByteIndex uint64, unpaddedBytesAmount uint64) error {
	resp := C.unseal_range(registeredProof, cacheDirPath, C.int32_t(sealedSectorFdRaw), C.int32_t(unsealOutputFdRaw), C.uint64_t(sectorId), proverId, ticket, commD, C.uint64_t(unpaddedByteIndex), C.uint64_t(unpaddedBytesAmount))
	defer resp.Destroy()
	if err := CheckErr(resp); err != nil {
		return err
	}
	return nil
}

func GenerateWinningPoStSectorChallenge(registeredProof RegisteredPoStProof, randomness *ByteArray32, sectorSetLen uint64, proverId *ByteArray32) ([]uint64, error) {
	resp := C.generate_winning_post_sector_challenge(registeredProof, randomness, C.uint64_t(sectorSetLen), proverId)
	defer resp.Destroy()
	if err := CheckErr(resp); err != nil {
		return nil, err
	}
	return resp.value.Copy(), nil
}

func GenerateWinningPoSt(randomness *ByteArray32, replicas SliceRefPrivateReplicaInfo, proverId *ByteArray32) ([]PoStProof, error) {
	resp := C.generate_winning_post(randomness, replicas, proverId)
	defer resp.Destroy()
	if err := CheckErr(resp); err != nil {
		return nil, err
	}
	return resp.value.Copy(), nil
}

func GenerateWindowPoSt(randomness *ByteArray32, replicas SliceRefPrivateReplicaInfo, proverId *ByteArray32) ([]PoStProof, []uint64, error) {
	resp := C.generate_window_post(randomness, replicas, proverId)
	defer resp.Destroy()
	faults := resp.value.faulty_sectors.Copy()

	if err := CheckErr(resp); err != nil {
		return nil, faults, err
	}
	proofs := resp.value.proofs.Copy()
	return proofs, faults, nil
}

func GetGpuDevices() ([]string, error) {
	resp := C.get_gpu_devices()
	defer resp.Destroy()
	if err := CheckErr(resp); err != nil {
		return nil, err
	}

	return resp.value.CopyAsStrings(), nil
}

func GetSealVersion(registeredProof RegisteredSealProof) (string, error) {
	resp := C.get_seal_version(registeredProof)
	defer resp.Destroy()
	if err := CheckErr(resp); err != nil {
		return "", err
	}

	return string(resp.value.Copy()), nil
}

func GetPoStVersion(registeredProof RegisteredPoStProof) (string, error) {
	resp := C.get_post_version(registeredProof)
	defer resp.Destroy()
	if err := CheckErr(resp); err != nil {
		return "", err
	}

	return string(resp.value.Copy()), nil
}

func GetNumPartitionForFallbackPost(registeredProof RegisteredPoStProof, numSectors uint) (uint, error) {
	resp := C.get_num_partition_for_fallback_post(registeredProof, C.size_t(numSectors))
	defer resp.Destroy()
	if err := CheckErr(resp); err != nil {
		return 0, err
	}

	return uint(resp.value), nil
}

func ClearCache(sectorSize uint64, cacheDirPath SliceRefUint8) error {
	resp := C.clear_cache(C.uint64_t(sectorSize), cacheDirPath)
	defer resp.Destroy()
	return CheckErr(resp)
}

func Fauxrep(registeredProf RegisteredSealProof, cacheDirPath SliceRefUint8, sealedSectorPath SliceRefUint8) ([]byte, error) {
	resp := C.fauxrep(registeredProf, cacheDirPath, sealedSectorPath)
	defer resp.Destroy()
	if err := CheckErr(resp); err != nil {
		return nil, err
	}
	return resp.value.Copy(), nil
}

func Fauxrep2(registeredProf RegisteredSealProof, cacheDirPath SliceRefUint8, existingPAuxPath SliceRefUint8) ([]byte, error) {
	resp := C.fauxrep2(registeredProf, cacheDirPath, existingPAuxPath)
	defer resp.Destroy()
	if err := CheckErr(resp); err != nil {
		return nil, err
	}
	return resp.value.Copy(), nil
}