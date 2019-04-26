package service

import (
	"encoding/json"
	"github.com/ribencong/go-youPipe/utils"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type Receipt struct {
	proofs   chan *PipeProof
	database *leveldb.DB
}

func newReceipt() *Receipt {

	db, err := leveldb.OpenFile(utils.SysConf.ReceiptPath, &opt.Options{
		Filter: filter.NewBloomFilter(10),
	})
	if err != nil {
		logger.Fatalf("YouPipe miner database invalid:%s", err)
	}

	r := &Receipt{
		proofs:   make(chan *PipeProof, MaxCharger),
		database: db,
	}

	return r
}

//TODO::make it thread
func (r *Receipt) DBWork() {
	logger.Debugf("start database to save receipt ")

	for {
		proof := <-r.proofs
		data, err := json.Marshal(proof)
		if err != nil {
			logger.Errorf("marshal proof err:%v", err)
			continue
		}

		pid := proof.ToID()
		if err := r.database.Put([]byte(pid), data, nil); err != nil {
			logger.Errorf("save  proof err:%v", err)
			continue
		}
	}
}
