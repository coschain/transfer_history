package types

const (
	LibTableName = "libinfo"
	TxDirectionSend = 1
	TxDirectionReceive = 2

	StatusSuccess = 200
	StatusIntervalError = 500
	StatusGetLibError = 501
	StatusGetTransferRecordError = 502
	StatusLackParamError = 503
	StatusParamInvalidError = 504
	StatusParamVerificationCodeInvalidError = 505
	StatusParamTransferDirectionInvalidError = 506
)


type BaseResponse struct {
	Status int
	Msg    string
}

type TransferRecord struct {
	OperationId string
	From   string
	To     string
	Memo   string
	Amount string
	BlockHeight  string
}

type TransferHistoryResponse struct {
	BaseResponse
	HeadBlockHeight string
	MaxBlockHeight string
    List []*TransferRecord
}

type QueryTransferRecordModel struct {
	List []*TransferRecord
	Lib  uint64
	MaxQueryBlkNum uint64
	Err  error
	ErrCode int
}

type SingleBlockTransferHistoryResponse struct {
	BaseResponse
	List []*TransferRecord
}
