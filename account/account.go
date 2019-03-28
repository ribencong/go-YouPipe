package account

import (
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
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
	logger            = utils.NewLog(utils.LMAccount)
	instance *Account = nil
	once     sync.Once
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
	defer key.Lock()
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
		if err := ioutil.WriteFile(path, []byte{}, 0644); err != nil {
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

//
//func (acc *Account) ToPubKey() []byte {
//	if len(acc.NodeId) <= len(AccPrefix) {
//		return nil
//	}
//	return base58.Decode(acc.NodeId[len(AccPrefix):])
//}
