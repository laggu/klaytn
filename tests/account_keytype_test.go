package tests

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/blockchain/types/accountkey"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/math"
	"github.com/ground-x/klaytn/common/profile"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/kerrors"
	"github.com/stretchr/testify/assert"
	"math/big"
	"strconv"
	"testing"
	"time"
)

// TestAccountCreationMaxMultiSigKey tests multiSig account creation with maximum private keys.
// A multiSig account supports maximum 10 different private keys.
// Create a multiSig account with 11 different private keys (more than 10 -> failed)
func TestAccountCreationMaxMultiSigKey(t *testing.T) {
	if testing.Verbose() {
		enableLog()
	}
	prof := profile.NewProfiler()

	// Initialize blockchain
	start := time.Now()
	bcdata, err := NewBCData(6, 4)
	if err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_blockchain", time.Now().Sub(start))
	defer bcdata.Shutdown()

	// Initialize address-balance map for verification
	start = time.Now()
	accountMap := NewAccountMap()
	if err := accountMap.Initialize(bcdata); err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_accountMap", time.Now().Sub(start))

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	// multisig setting
	threshold := uint(10)
	weights := []uint{1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 0, 1}
	multisigAddr := common.HexToAddress("0xbbfa38050bf3167c887c086758f448ce067ea8ea")
	prvKeys := []string{
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380000",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380001",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380002",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380003",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380004",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380005",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380006",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594300007",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594300008",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594300009",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594300010",
	}

	// multi-sig account
	multisig, err := createMultisigAccount(threshold, weights, prvKeys, multisigAddr)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("multisigAddr = ", multisig.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// Create a multiSig account with 11 different private keys (more than 10 -> failed)
	{
		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            multisig.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    multisig.AccKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, kerrors.ErrMaxKeysExceed, err)
		assert.Equal(t, (*types.Receipt)(nil), receipt)

		reservoir.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountCreationMultiSigKeyBigThreshold tests multiSig account creation with abnormal threshold.
// When a multisig account is created, a threshold value should be less or equal to the total weight of private keys.
// If not, the account cannot creates any valid signatures.
// The test creates a multisig account with a threshold (10) and the total weight (6). (failed case)
func TestAccountCreationMultiSigKeyBigThreshold(t *testing.T) {
	if testing.Verbose() {
		enableLog()
	}
	prof := profile.NewProfiler()

	// Initialize blockchain
	start := time.Now()
	bcdata, err := NewBCData(6, 4)
	if err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_blockchain", time.Now().Sub(start))
	defer bcdata.Shutdown()

	// Initialize address-balance map for verification
	start = time.Now()
	accountMap := NewAccountMap()
	if err := accountMap.Initialize(bcdata); err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_accountMap", time.Now().Sub(start))

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	// multisig setting
	threshold := uint(10)
	weights := []uint{1, 2, 3}
	multisigAddr := common.HexToAddress("0xbbfa38050bf3167c887c086758f448ce067ea8ea")
	prvKeys := []string{
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380000",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380001",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380002",
	}

	// multi-sig account
	multisig, err := createMultisigAccount(threshold, weights, prvKeys, multisigAddr)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("multisigAddr = ", multisig.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// creates a multisig account with a threshold (10) and the total weight (6). (failed case)
	{
		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            multisig.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    multisig.AccKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrUnsatisfiableThreshold, receipt.Status)

		reservoir.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountCreationRoleBasedKeyNested tests account creation with nested RoleBasedKey.
// Nested RoleBasedKey is not allowed in Klaytn.
// The test should fail to the account creation
// 1. A key for the first role, RoleTransaction, is nested
// 2. A key for the second role, RoleAccountUpdate, is nested.
// 3. A key for the third role, RoleFeePayer, is nested.
func TestAccountCreationRoleBasedKeyNested(t *testing.T) {
	if testing.Verbose() {
		enableLog()
	}
	prof := profile.NewProfiler()

	// Initialize blockchain
	start := time.Now()
	bcdata, err := NewBCData(6, 4)
	if err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_blockchain", time.Now().Sub(start))
	defer bcdata.Shutdown()

	// Initialize address-balance map for verification
	start = time.Now()
	accountMap := NewAccountMap()
	if err := accountMap.Initialize(bcdata); err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_accountMap", time.Now().Sub(start))

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	anon, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)
	anon2, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78990001")
	assert.Equal(t, nil, err)
	anon3, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78990002")
	assert.Equal(t, nil, err)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("anonAddr = ", anon.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// 1. A key for the first role, RoleTransaction, is nested
	{
		keys := genTestKeys(3)
		roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[2].PublicKey),
		})

		keys2 := genTestKeys(2)
		nestedKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			roleKey,
			accountkey.NewAccountKeyPublicWithValue(&keys2[0].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys2[1].PublicKey),
		})

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            anon.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    nestedKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrNestedRoleBasedKey, receipt.Status)
	}

	// 2. A key for the second role, RoleAccountUpdate, is nested.
	{
		keys := genTestKeys(3)
		roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[2].PublicKey),
		})

		keys2 := genTestKeys(2)
		nestedKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyPublicWithValue(&keys2[0].PublicKey),
			roleKey,
			accountkey.NewAccountKeyPublicWithValue(&keys2[1].PublicKey),
		})

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            anon2.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    nestedKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrNestedRoleBasedKey, receipt.Status)

	}

	// 3. A key for the third role, RoleFeePayer, is nested.
	{
		keys := genTestKeys(3)
		roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[2].PublicKey),
		})

		keys2 := genTestKeys(2)
		nestedKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyPublicWithValue(&keys2[0].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys2[1].PublicKey),
			roleKey,
		})

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            anon3.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    nestedKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrNestedRoleBasedKey, receipt.Status)
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountCreationRoleBasedKeyInvalidNumKey tests account creation with a RoleBased key which contains invalid number of sub-keys.
// A RoleBased key can contain 1 ~ 3 sub-keys, otherwise it will fail to the account creation.
// 1. try to create an account with a RoleBased key which contains 4 sub-keys.
// 2. try to create an account with a RoleBased key which contains 0 sub-key.
func TestAccountCreationRoleBasedKeyInvalidNumKey(t *testing.T) {
	if testing.Verbose() {
		enableLog()
	}
	prof := profile.NewProfiler()

	// Initialize blockchain
	start := time.Now()
	bcdata, err := NewBCData(6, 4)
	if err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_blockchain", time.Now().Sub(start))
	defer bcdata.Shutdown()

	// Initialize address-balance map for verification
	start = time.Now()
	accountMap := NewAccountMap()
	if err := accountMap.Initialize(bcdata); err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_accountMap", time.Now().Sub(start))

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	anon, err := createAnonymousAccount("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("anonAddr = ", anon.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// 1. try to create an account with a RoleBased key which contains 4 sub-keys.
	{
		keys := genTestKeys(4)
		roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{
			accountkey.NewAccountKeyPublicWithValue(&keys[0].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[1].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[2].PublicKey),
			accountkey.NewAccountKeyPublicWithValue(&keys[3].PublicKey),
		})

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            anon.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    roleKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrLengthTooLong, receipt.Status)
	}

	// 2. try to create an account with a RoleBased key which contains 0 sub-key.
	{
		roleKey := accountkey.NewAccountKeyRoleBasedWithValues(accountkey.AccountKeyRoleBased{})

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            anon.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    roleKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrZeroLength, receipt.Status)
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountCreationMultiSigKeyDupPrvKeys tests multiSig account creation with duplicated private keys.
// A multisig account has all different private keys, therefore account creation with duplicated private keys should be failed.
// The case when two same private keys are used in creation processes.
func TestAccountCreationMultiSigKeyDupPrvKeys(t *testing.T) {
	if testing.Verbose() {
		enableLog()
	}
	prof := profile.NewProfiler()

	// Initialize blockchain
	start := time.Now()
	bcdata, err := NewBCData(6, 4)
	if err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_blockchain", time.Now().Sub(start))
	defer bcdata.Shutdown()

	// Initialize address-balance map for verification
	start = time.Now()
	accountMap := NewAccountMap()
	if err := accountMap.Initialize(bcdata); err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_accountMap", time.Now().Sub(start))

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	// the case when two same private keys are used in creation processes.
	threshold := uint(2)
	weights := []uint{1, 1, 2}
	multisigAddr := common.HexToAddress("0xbbfa38050bf3167c887c086758f448ce067ea8ea")
	prvKeys := []string{
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380000",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380001",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380001",
	}

	// multi-sig account
	multisig, err := createMultisigAccount(threshold, weights, prvKeys, multisigAddr)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("multisigAddr = ", multisig.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// the case when two same private keys are used in creation processes.
	{
		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            multisig.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    multisig.AccKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrDuplicatedKey, receipt.Status)

		reservoir.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountCreationMultiSigKeyWeightOverflow tests multiSig account creation with weight overflow.
// If the sum of weights is overflowed, the test should fail.
func TestAccountCreationMultiSigKeyWeightOverflow(t *testing.T) {
	if testing.Verbose() {
		enableLog()
	}
	prof := profile.NewProfiler()

	// Initialize blockchain
	start := time.Now()
	bcdata, err := NewBCData(6, 4)
	if err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_blockchain", time.Now().Sub(start))
	defer bcdata.Shutdown()

	// Initialize address-balance map for verification
	start = time.Now()
	accountMap := NewAccountMap()
	if err := accountMap.Initialize(bcdata); err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_accountMap", time.Now().Sub(start))

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	// Simply check & set the maximum value of uint
	MAX := uint(math.MaxUint32)
	if strconv.IntSize == 64 {
		MAX = math.MaxUint64
	}

	// multisig setting
	threshold := uint(MAX)
	weights := []uint{MAX / 2, MAX / 2, MAX / 2}
	multisigAddr := common.HexToAddress("0xbbfa38050bf3167c887c086758f448ce067ea8ea")
	prvKeys := []string{
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380000",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380001",
		"a5c9a50938a089618167c9d67dbebc0deaffc3c76ddc6b40c2777ae594380002",
	}

	// multi-sig account
	multisig, err := createMultisigAccount(threshold, weights, prvKeys, multisigAddr)

	if testing.Verbose() {
		fmt.Println("reservoirAddr = ", reservoir.Addr.String())
		fmt.Println("multisigAddr = ", multisig.Addr.String())
	}

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// creates a multisig account with a threshold, uint(MAX), and the total weight, uint(MAX/2)*3. (failed case)
	{
		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            multisig.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: false,
			types.TxValueKeyAccountKey:    multisig.AccKey,
		}
		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrWeightedSumOverflow, receipt.Status)

		reservoir.Nonce += 1
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}

// TestAccountCreationHumanReadableFail tests account creation with an invalid address.
// For a human-readable account, only alphanumeric characters are allowed in the address.
// In addition, the first character should be an alphabet.
// 1. Non-alphanumeric characters in the address.
// 2. The first character of the address is a number.
// 3. A valid address, "humanReadable"
func TestAccountCreationHumanReadableFail(t *testing.T) {
	if testing.Verbose() {
		enableLog()
	}
	prof := profile.NewProfiler()

	// Initialize blockchain
	start := time.Now()
	bcdata, err := NewBCData(6, 4)
	if err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_blockchain", time.Now().Sub(start))
	defer bcdata.Shutdown()

	// Initialize address-balance map for verification
	start = time.Now()
	accountMap := NewAccountMap()
	if err := accountMap.Initialize(bcdata); err != nil {
		t.Fatal(err)
	}
	prof.Profile("main_init_accountMap", time.Now().Sub(start))

	// reservoir account
	reservoir := &TestAccountType{
		Addr:  *bcdata.addrs[0],
		Keys:  []*ecdsa.PrivateKey{bcdata.privKeys[0]},
		Nonce: uint64(0),
	}

	key, err := crypto.HexToECDSA("98275a145bc1726eb0445433088f5f882f8a4a9499135239cfb4040e78991dab")
	assert.Equal(t, nil, err)

	signer := types.NewEIP155Signer(bcdata.bc.Config().ChainID)

	// 1. Non-alphanumeric characters in the address.
	{
		readable := &TestAccountType{
			Addr:   common.HexToAddress("68756d616e5265616461626c655f000000000000"), // humanReadable_
			Keys:   []*ecdsa.PrivateKey{key},
			Nonce:  uint64(0),
			AccKey: accountkey.NewAccountKeyPublicWithValue(&key.PublicKey),
		}

		if testing.Verbose() {
			fmt.Println("reservoirAddr = ", reservoir.Addr.String())
			fmt.Println("anonAddr = ", readable.Addr.String())
		}

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            readable.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyAccountKey:    readable.AccKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrNotHumanReadableAddress, receipt.Status)
	}

	// 2. The first character of the address is a number.
	{
		readable := &TestAccountType{
			Addr:   common.HexToAddress("3068756d616e5265616461626c65000000000000"), // 0humanReadable
			Keys:   []*ecdsa.PrivateKey{key},
			Nonce:  uint64(0),
			AccKey: accountkey.NewAccountKeyPublicWithValue(&key.PublicKey),
		}

		if testing.Verbose() {
			fmt.Println("reservoirAddr = ", reservoir.Addr.String())
			fmt.Println("anonAddr = ", readable.Addr.String())
		}

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            readable.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyAccountKey:    readable.AccKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusErrNotHumanReadableAddress, receipt.Status)
	}

	// 3. A valid address, "humanReadable"
	{
		readable := &TestAccountType{
			Addr:   common.HexToAddress("68756d616e5265616461626c6500000000000000"), // humanReadable
			Keys:   []*ecdsa.PrivateKey{key},
			Nonce:  uint64(0),
			AccKey: accountkey.NewAccountKeyPublicWithValue(&key.PublicKey),
		}

		if testing.Verbose() {
			fmt.Println("reservoirAddr = ", reservoir.Addr.String())
			fmt.Println("anonAddr = ", readable.Addr.String())
		}

		amount := new(big.Int).SetUint64(1000000000000)
		values := map[types.TxValueKeyType]interface{}{
			types.TxValueKeyNonce:         reservoir.Nonce,
			types.TxValueKeyFrom:          reservoir.Addr,
			types.TxValueKeyTo:            readable.Addr,
			types.TxValueKeyAmount:        amount,
			types.TxValueKeyGasLimit:      gasLimit,
			types.TxValueKeyGasPrice:      gasPrice,
			types.TxValueKeyHumanReadable: true,
			types.TxValueKeyAccountKey:    readable.AccKey,
		}

		tx, err := types.NewTransactionWithMap(types.TxTypeAccountCreation, values)
		assert.Equal(t, nil, err)

		err = tx.SignWithKeys(signer, reservoir.Keys)
		assert.Equal(t, nil, err)

		receipt, _, err := applyTransaction(t, bcdata, tx)
		assert.Equal(t, nil, err)
		assert.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}
