package gossip

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/ribencong/go-youPipe/pbs"
	"io"
	"net"
	"time"
)

const (
	GspMsgHeadSize     = 8
	ConditionalForward = 1
	HeartBeatTime      = time.Minute * 15
	RetrySubInterval   = time.Second * 15
	MsgCacheSize       = 1 << 15
	CommonTCPTimeOut   = time.Second * 4
	UpdateThreshold    = 10
	ForwardThreshold   = 10

	SubInit       = int(pbs.MsgType_SubInit)
	Forward       = int(pbs.MsgType_Forward)
	VoteContact   = int(pbs.MsgType_VoteContact)
	GotContact    = int(pbs.MsgType_GotContact)
	HeartBeat     = int(pbs.MsgType_HeartBeat)
	WelCome       = int(pbs.MsgType_WelCome)
	UpdateWeight  = int(pbs.MsgType_UpdateWeight)
	ReSubscribe   = int(pbs.MsgType_ReSubscribe)
	SubSuccess    = int(pbs.MsgType_SubSuccess)
	ReplaceView   = int(pbs.MsgType_ReplaceView)
	NewForReplace = int(pbs.MsgType_NewForReplace)
	AppPayload    = int(pbs.MsgType_AppPayload)

	BroadCastTarget = "-1"
	AppMsgBroadCast = 10000
)

var (
	ESelfReq       = fmt.Errorf("it's myself")
	EDuplicateConn = fmt.Errorf("duplicated connection")
	EInvalidMsg    = fmt.Errorf("unknown message type")
	ENotFound      = fmt.Errorf("not found")
	EOverForward   = fmt.Errorf("too much forwarded times")
	EInUsed        = fmt.Errorf("app watch port in used")
	ENotConnected  = fmt.Errorf("gossip connection is not connected, use SendTo instead")
)

func pack(typ int, nid string, param ...interface{}) (data []byte, err error) {
	t := pbs.MsgType(typ)
	var body []byte
	switch typ {
	case SubInit, WelCome, HeartBeat,
		ReSubscribe, SubSuccess, NewForReplace:
		body, err = proto.Marshal(&pbs.Gossip{
			ID: &pbs.ID{
				NodeId: nid,
			},
		})
	case GotContact:
		ip, _ := param[0].(string)
		body, err = proto.Marshal(&pbs.Gossip{
			IDWithIP: &pbs.IDWithIP{
				NodeId: nid,
				IP:     ip,
			},
		})

	case Forward:
		ip, _ := param[0].(string)
		msgId, _ := param[1].(string)
		body, err = proto.Marshal(&pbs.Gossip{
			Forward: &pbs.ForwardMsg{
				NodeId: nid,
				IP:     ip,
				MsgId:  msgId,
			},
		})
	case VoteContact:
		ttl, _ := param[1].(int32)
		ip, _ := param[0].(string)

		body, err = proto.Marshal(&pbs.Gossip{
			Vote: &pbs.Vote{
				NodeId: nid,
				IP:     ip,
				TTL:    ttl,
			},
		})
	case UpdateWeight:
		w := param[0].(float64)
		d := param[1].(VNodeDirect)

		body, err = proto.Marshal(&pbs.Gossip{
			UpdateWeight: &pbs.Weight{
				NodeId: nid,
				Weight: w,
				Direct: int32(d),
			},
		})
	case ReplaceView:
		r := param[0].(string)
		i := param[1].(string)

		body, err = proto.Marshal(&pbs.Gossip{
			RplView: &pbs.Replace{
				NodeId:  nid,
				AlterId: r,
				IP:      i,
			},
		})

	case AppPayload:
		from := param[0].(string)
		to := param[1].(string)
		ttl := param[2].(int32)
		pl := param[3].([]byte)

		body, err = proto.Marshal(&pbs.Gossip{
			AppMsg: &pbs.AppMsg{
				MsgId:   nid,
				LAddr:   from,
				RAddr:   to,
				TTL:     ttl,
				PayLoad: pl,
			},
		})

	default:
		return nil, errors.New("unknown msg type")
	}

	//logger.Debugf("pack (%d) data:%v", len(body), body)
	b := make([]byte, GspMsgHeadSize)
	binary.BigEndian.PutUint32(b[:4], uint32(t))
	binary.BigEndian.PutUint32(b[4:GspMsgHeadSize], uint32(len(body)))
	b = append(b, body...)
	return b, err
}

func pullMsg(conn net.Conn) (pbs.MsgType, *pbs.Gossip, error) {

	buf := make([]byte, GspMsgHeadSize)
	if _, err := conn.Read(buf); err != nil {
		logger.Warning("read head err:->", err)
		return 0, nil, err
	}

	typ := pbs.MsgType(binary.BigEndian.Uint32(buf[:4]))
	l := int(binary.BigEndian.Uint32(buf[4:GspMsgHeadSize]))
	//logger.Debugf("gossip head typ=%s len=%d", typ, l)

	data := make([]byte, l)
	if _, err := io.ReadFull(conn, data); err != nil {
		logger.Warning("read data err:->", err)
		return 0, nil, err
	}

	body := &pbs.Gossip{}
	if err := proto.Unmarshal(data, body); err != nil {
		logger.Warning("unpack data err:->", err)
		return 0, nil, err
	}

	return typ, body, nil
}
