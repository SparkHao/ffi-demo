package cgo

/*
#cgo LDFLAGS: -L${SRCDIR}/..
#cgo pkg-config: ${SRCDIR}/../filcrypto.pc
#include "../filcrypto.h"
#include <stdlib.h>
*/
import "C"

func Hash(message SliceRefUint8) *[96]byte {
	resp := C.hash(message)
	defer resp.Destroy()
	return resp.CopyAsArray()
}

func Aggregate(flattenedSignatures SliceRefUint8) *[96]byte {
	resp := C.aggregate(flattenedSignatures)
	defer resp.Destroy()
	return resp.CopyAsArray()
}

func Verify(signature SliceRefUint8, flattenedDigests SliceRefUint8, flattenedPublicKeys SliceRefUint8) bool {
	resp := C.verify(signature, flattenedDigests, flattenedPublicKeys)
	return bool(resp)
}

func HashVerify(signature SliceRefUint8, flattenedMessages SliceRefUint8, messageSizes SliceRefUint, flattenedPublicKeys SliceRefUint8) bool {
	resp := C.hash_verify(signature, flattenedMessages, messageSizes, flattenedPublicKeys)
	return bool(resp)
}

func PrivateKeyGenerate() *[32]byte {
	resp := C.private_key_generate()
	defer resp.Destroy()
	return resp.CopyAsArray()
}

func PrivateKeyGenerateWithSeed(rawSeed *ByteArray32) *[32]byte {
	resp := C.private_key_generate_with_seed(rawSeed)
	defer resp.Destroy()
	return resp.CopyAsArray()
}

func PrivateKeySign(rawPrivateKey SliceRefUint8, message SliceRefUint8) *[96]byte {
	resp := C.private_key_sign(rawPrivateKey, message)
	defer resp.Destroy()
	return resp.CopyAsArray()
}

func PrivateKeyPublicKey(rawPrivateKey SliceRefUint8) *[48]byte {
	resp := C.private_key_public_key(rawPrivateKey)
	defer resp.Destroy()
	return resp.CopyAsArray()
}
func CreateZeroSignature() *[96]byte {
	resp := C.create_zero_signature()
	defer resp.Destroy()
	return resp.CopyAsArray()
}
