package apiChat

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	"Open_IM/pkg/proto/chat"
	"Open_IM/pkg/proto/sdk_ws"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type paramsUserPullMsg struct {
	ReqIdentifier *int   `json:"reqIdentifier" binding:"required"`
	SendID        string `json:"sendID" binding:"required"`
	OperationID   string `json:"operationID" binding:"required"`
	Data          struct {
		SeqBegin *int64 `json:"seqBegin" binding:"required"`
		SeqEnd   *int64 `json:"seqEnd" binding:"required"`
	}
}

func PullMsg(c *gin.Context) {
	params := paramsUserPullMsg{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	token := c.Request.Header.Get("token")
	if ok, err := token_verify.VerifyToken(token, params.SendID); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "token validate err" + err.Error()})
		return
	}
	pbData := open_im_sdk.PullMessageReq{}
	pbData.UserID = params.SendID
	pbData.OperationID = params.OperationID
	pbData.SeqBegin = *params.Data.SeqBegin
	pbData.SeqEnd = *params.Data.SeqEnd
	grpcConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOfflineMessageName)
	msgClient := pbChat.NewChatClient(grpcConn)
	reply, err := msgClient.PullMessage(context.Background(), &pbData)
	if err != nil {
		log.NewError(params.OperationID, "UserPullMsg rpc failed, ", params, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 401, "errMsg": "UserPullMsg rpc failed, " + err.Error()})
		return
	}
	log.InfoByKv("rpc call success to pullMsgRep", pbData.OperationID, "ReplyArgs", reply.String(), "maxSeq", reply.GetMaxSeq(),
		"MinSeq", reply.GetMinSeq(), "singLen", len(reply.GetSingleUserMsg()), "groupLen", len(reply.GetGroupUserMsg()))

	msg := make(map[string]interface{})
	if v := reply.GetSingleUserMsg(); v != nil {
		msg["single"] = v
	} else {
		msg["single"] = []open_im_sdk.GatherFormat{}
	}
	if v := reply.GetGroupUserMsg(); v != nil {
		msg["group"] = v
	} else {
		msg["group"] = []open_im_sdk.GatherFormat{}
	}
	msg["maxSeq"] = reply.GetMaxSeq()
	msg["minSeq"] = reply.GetMinSeq()
	c.JSON(http.StatusOK, gin.H{
		"errCode":       reply.ErrCode,
		"errMsg":        reply.ErrMsg,
		"reqIdentifier": *params.ReqIdentifier,
		"data":          msg,
	})

}

type paramsUserPullMsgBySeqList struct {
	ReqIdentifier int     `json:"reqIdentifier" binding:"required"`
	SendID        string  `json:"sendID" binding:"required"`
	OperationID   string  `json:"operationID" binding:"required"`
	SeqList       []int64 `json:"seqList"`
}

func PullMsgBySeqList(c *gin.Context) {
	params := paramsUserPullMsgBySeqList{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	token := c.Request.Header.Get("token")
	if ok, err := token_verify.VerifyToken(token, params.SendID); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "token validate err" + err.Error()})
		return
	}
	pbData := open_im_sdk.PullMessageBySeqListReq{}
	pbData.UserID = params.SendID
	pbData.OperationID = params.OperationID
	pbData.SeqList = params.SeqList

	grpcConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOfflineMessageName)
	msgClient := pbChat.NewChatClient(grpcConn)
	reply, err := msgClient.PullMessageBySeqList(context.Background(), &pbData)
	if err != nil {
		log.ErrorByKv("PullMessageBySeqList error", pbData.OperationID, "err", err.Error())
		return
	}
	log.InfoByKv("rpc call success to PullMessageBySeqList", pbData.OperationID, "ReplyArgs", reply.String(), "maxSeq", reply.GetMaxSeq(),
		"MinSeq", reply.GetMinSeq(), "singLen", len(reply.GetSingleUserMsg()), "groupLen", len(reply.GetGroupUserMsg()))

	msg := make(map[string]interface{})
	if v := reply.GetSingleUserMsg(); v != nil {
		msg["single"] = v
	} else {
		msg["single"] = []open_im_sdk.GatherFormat{}
	}
	if v := reply.GetGroupUserMsg(); v != nil {
		msg["group"] = v
	} else {
		msg["group"] = []open_im_sdk.GatherFormat{}
	}
	msg["maxSeq"] = reply.GetMaxSeq()
	msg["minSeq"] = reply.GetMinSeq()
	c.JSON(http.StatusOK, gin.H{
		"errCode":       reply.ErrCode,
		"errMsg":        reply.ErrMsg,
		"reqIdentifier": params.ReqIdentifier,
		"data":          msg,
	})
}
