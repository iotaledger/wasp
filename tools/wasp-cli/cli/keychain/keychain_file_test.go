package keychain

import (
	"os"
	"testing"

	"github.com/awnumar/memguard"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

func TestGetSetSeed(t *testing.T) {
	wd, _ := os.Getwd()

	keyChainFile := NewKeyChainFile(wd, func() *memguard.Enclave {
		return memguard.NewEnclave([]byte("HI"))
	})

	testSeed := cryptolib.NewSeed()
	err := keyChainFile.SetSeed(testSeed)
	require.NoError(t, err)

	pulledSeed, err := keyChainFile.GetSeed()
	require.NoError(t, err)

	require.EqualValues(t, pulledSeed, &testSeed)
}

func TestGetSetJWT(t *testing.T) {
	wd, _ := os.Getwd()

	keyChainFile := NewKeyChainFile(wd, func() *memguard.Enclave {
		return memguard.NewEnclave([]byte("HI"))
	})

	testJWT := "ey....."
	err := keyChainFile.SetJWTAuthToken("wasp", testJWT)
	require.NoError(t, err)

	pulledSeed, err := keyChainFile.GetJWTAuthToken("wasp")
	require.NoError(t, err)

	require.EqualValues(t, pulledSeed, testJWT)
}

func TestGetSetStrongholdPassword(t *testing.T) {
	wd, _ := os.Getwd()

	keyChainFile := NewKeyChainFile(wd, func() *memguard.Enclave {
		return memguard.NewEnclave([]byte("HI"))
	})

	testStrongholdPassword := memguard.NewEnclaveRandom(32)
	err := keyChainFile.SetStrongholdPassword(testStrongholdPassword)
	require.NoError(t, err)

	pulledStrongholdPassword, err := keyChainFile.GetStrongholdPassword()
	require.NoError(t, err)

	bufferA, err := testStrongholdPassword.Open()
	require.NoError(t, err)
	defer bufferA.Destroy()

	bufferB, err := pulledStrongholdPassword.Open()
	require.NoError(t, err)
	defer bufferB.Destroy()

	require.EqualValues(t, bufferA.Bytes(), bufferB.Bytes())
}

func TestInvalidPassword(t *testing.T) {
	wd, _ := os.Getwd()

	keyChainFile := NewKeyChainFile(wd, func() *memguard.Enclave {
		return memguard.NewEnclave([]byte("HI"))
	})

	err := keyChainFile.SetJWTAuthToken("wasp", "asd")
	require.NoError(t, err)

	keyChainFile2 := NewKeyChainFile(wd, func() *memguard.Enclave {
		return memguard.NewEnclave([]byte("WRONG_PW"))
	})

	_, err = keyChainFile2.GetJWTAuthToken("wasp")
	require.ErrorIs(t, err, ErrInvalidPassword)
}
