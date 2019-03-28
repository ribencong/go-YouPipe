package account

import (
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/golang/protobuf/proto"
	"github.com/youpipe/go-youPipe/pbs"
	"github.com/youpipe/go-youPipe/utils"
	"io/ioutil"
	"sync"
)

type Account struct {
	NodeId string
	Key    *Key
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

	acc := &pbs.Account{
		NodeId: key.ToNodeId(),
		Key: &pbs.Key{
			PrivateKey: key.PriKey,
			PublicKey:  key.PubKey[:],
		},
	}

	data, err := proto.Marshal(acc)
	if err != nil {
		panic(err)
	}

	path := utils.SysConf.AccDataPath
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		panic(err)
	}
	return acc.NodeId
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
	if _, ok := utils.FileExists(path); !ok {

		err := ioutil.WriteFile(path, []byte{}, 0644)
		if err != nil {
			panic(err)
		}

		return obj
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	if len(data) == 0 {
		return obj
	}

	pbAcc := pbs.Account{}
	if err := proto.Unmarshal(data, &pbAcc); err != nil {
		logger.Panicf("unknown account data :->%v", err)
	}
	obj.NodeId = pbAcc.NodeId
	obj.Key = &Key{
		PriKey: pbAcc.Key.PrivateKey,
	}
	copy(obj.Key.PubKey[:], pbAcc.Key.PublicKey)

	return obj
}

func (acc *Account) ToPubKey() PublicKey {

	if len(acc.NodeId) <= len(AccPrefix) {
		return PublicKey{}
	}

	realStr := acc.NodeId[len(AccPrefix):]
	pubData := base58.Decode(realStr)

	var key PublicKey
	copy(key[:], pubData)
	return key
}
