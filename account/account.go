package account

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/op/go-logging"
	"github.com/youpipe/go-youPipe/utils"
	"gx/ipfs/QmW7VUmSvhvSGbYbdsh7uRjhGmsYkc9fL8aJ5CorxxrU5N/go-crypto/ed25519"
	"io/ioutil"
	"os"
	"sync"
)

type Account struct {
	sync.RWMutex
	Address ID
	Key     *Key
}

type SafeAccount struct {
	Version string `json:"version"`
	Address ID     `json:"address"`
	Cipher  string `json:"cipher"`
}

var (
	logger, _          = logging.GetLogger(utils.LMAccount)
	instance  *Account = nil
	once      sync.Once
)

func GetAccount() *Account {
	once.Do(func() {
		instance = newNode()
	})

	return instance
}

func CreateAccount(password string) SafeAccount {

	key, err := GenerateKey(password)
	if err != nil {
		panic(err)
	}
	address := key.ToNodeId()
	w := SafeAccount{
		Version: utils.CurrentVersion,
		Address: address,
		Cipher:  base58.Encode(key.LockedKey),
	}
	return w
}

func SaveToDisk(w SafeAccount) {
	data, err := json.Marshal(w)
	if err != nil {
		panic(err)
	}

	path := utils.SysConf.AccDataPath
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		panic(err)
	}
}

func (acc *Account) IsEmpty() bool {
	return len(acc.Address) == 0
}

func (acc *Account) FormatShow() string {
	ret := fmt.Sprintf("\n**********************************************************************\n"+
		"\tNodeID:\t%s"+
		"\n**********************************************************************\n",
		acc.Address)

	return ret
}

func newNode() *Account {
	obj := &Account{}

	path := utils.SysConf.AccDataPath
	fil, err := os.Open(path)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		return obj
	}
	defer fil.Close()

	acc := &SafeAccount{}
	parser := json.NewDecoder(fil)
	if err = parser.Decode(acc); err != nil {
		panic(err)
	}

	obj.Address = acc.Address
	obj.Key = &Key{
		LockedKey: base58.Decode(acc.Cipher),
	}

	return obj
}

//func (acc *Account) LockAcc() {
//	acc.Key.Lock()
//}

func (acc *Account) UnlockAcc(password string) bool {
	pk := acc.Address.ToPubKey()

	aesKey, err := AESKey(pk[:KP.S], password) //scrypt.Key([]byte(password), k.PubKey[:KP.S], KP.N, KP.R, KP.P, KP.L)
	if err != nil {
		fmt.Println("error to generate aes key:->", err)
		return false
	}

	cpTxt := make([]byte, len(acc.Key.LockedKey))
	copy(cpTxt, acc.Key.LockedKey)

	raw, err := Decrypt(aesKey, cpTxt)
	if err != nil {
		fmt.Println("Unlock raw private key:->", err)
		return false
	}
	tmpPub, tmpPri := populateKey(raw)
	if !bytes.Equal(pk, tmpPub[:]) {
		fmt.Println("Unlock public failed")
		return false
	}

	acc.Key.PubKey = tmpPub
	acc.Key.PriKey = tmpPri
	return true
}

func (acc *Account) CreateAesKey(key *[32]byte, peerAddr string) error {

	id, err := ConvertToID(peerAddr)
	if err != nil {
		return err
	}

	peerPub := id.ToPubKey()

	return acc.Key.GenerateAesKey(key, peerPub)
}

func CheckID(address string) bool {
	if len(address) <= len(AccPrefix) {
		return false
	}

	id := ID(address)
	pk := id.ToPubKey()
	if len(pk) != ed25519.PublicKeySize {
		return false
	}

	return true
}
