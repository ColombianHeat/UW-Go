package main

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

func encrypt(data []byte) ([]byte, error) {
	var outBlob windows.DataBlob

	inBlob := windows.DataBlob{
		Size: uint32(len(data)),
		Data: &data[0],
	}
	err := windows.CryptProtectData(&inBlob, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &outBlob)
	if err != nil {
		return nil, err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(outBlob.Data)))

	out := make([]byte, outBlob.Size)
	copy(out, (*[1 << 30]byte)(unsafe.Pointer(outBlob.Data))[:outBlob.Size:outBlob.Size])
	return out, nil
}

func decrypt(data []byte) ([]byte, error) {
	var outBlob windows.DataBlob

	inBlob := windows.DataBlob{
		Size: uint32(len(data)),
		Data: &data[0],
	}
	err := windows.CryptUnprotectData(&inBlob, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &outBlob)
	if err != nil {
		return nil, err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(outBlob.Data)))

	out := make([]byte, outBlob.Size)
	copy(out, (*[1 << 30]byte)(unsafe.Pointer(outBlob.Data))[:outBlob.Size:outBlob.Size])
	return out, nil
}