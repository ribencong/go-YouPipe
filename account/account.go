package account

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/op/go-logging"
	"github.com/youpipe/go-youPipe/utils"
	"io/ioutil"
	"os"
	"sync"
)

type Account struct {
	sync.RWMutex
	NodeId string
	Key    *Key
}

type accData struct {
	Version string `json:"version"`
	Address string `json:"address"`
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

func CreateAccount(password string) string {

	key, err := GenerateKey(password)
	if err != nil {
		panic(err)
	}
	address := key.ToNodeId()
	w := accData{
		Version: utils.CurrentVersion,
		Address: address,
		Cipher:  base58.Encode(key.LockedKey),
	}
	data, err := json.Marshal(w)
	if err != nil {
		panic(err)
	}

	path := utils.SysConf.AccDataPath
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		panic(err)
	}
	return address
}

func (acc *Account) IsEmpty() bool {
	return len(acc.NodeId) == 0
}

func (acc *Account) FormatShow() string {
	ret := fmt.Sprintf("\n**********************************************************************\n"+
		"\tNodeID:\t%s"+
		"\n**********************************************************************\n",
		acc.NodeId)

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

	acc := &accData{}
	parser := json.NewDecoder(fil)
	if err = parser.Decode(acc); err != nil {
		panic(err)
	}

	obj.NodeId = acc.Address
	obj.Key = &Key{
		LockedKey: base58.Decode(acc.Cipher),
	}

	return obj
}

//func (acc *Account) LockAcc() {
//	acc.Key.Lock()
//}

func (acc *Account) UnlockAcc(password string) bool {
	fmt.Println("\n should remove this:->", password)

	pk := ToPubKey(acc.NodeId)

	aesKey, err := getAESKey(pk[:kp.S], password) //scrypt.Key([]byte(password), k.PubKey[:kp.S], kp.N, kp.R, kp.P, kp.L)
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
	acc.Key.priKey = tmpPri
	return true
}

func ToPubKey(nid string) []byte {
	if len(nid) <= len(AccPrefix) {
		return nil
	}
	ss := nid[len(AccPrefix):]
	return base58.Decode(ss)
}
