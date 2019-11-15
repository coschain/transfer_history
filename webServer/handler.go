package webServer

import (
	"encoding/json"
	"errors"
	"fmt"
	"transfer_history/config"
	"transfer_history/db"
	"transfer_history/logs"
	"transfer_history/types"
	"transfer_history/utils"
	"net/http"
	"net/url"
	"strconv"
)

const (
	startBlockNumKey = "start"
	singleBlockKey = "block"
	accountNameKey = "account"
	txDirectionKey = "direction"
	verificationCodeKey = "code"
)

type historyParamsModel struct {
	verificationCode string
	txDirection  int
	account  string
	err error
	errCode int
}

//
// get a list of user's transfer history
//
func getTransferHistory(w http.ResponseWriter, r *http.Request)  {
	w.Header().Add("Access-Control-Allow-Origin", "*")
    logger := logs.GetLogger()
	res := types.TransferHistoryResponse{
		List: make([]*types.TransferRecord,0),
	}

	paramsInfo := parseHistoryParams(r)
	if paramsInfo.err != nil {
		res.Status = paramsInfo.errCode
		res.Msg  = paramsInfo.err.Error()
		writeResponse(w, res)
		return
	}
	dir := paramsInfo.txDirection
	acctName := paramsInfo.account

	// get start block height
	startBlkNum,err,code := parseBlockNumberByKey(r, startBlockNumKey)
	if err != nil {
		res.Status = code
		res.Msg = err.Error()
		writeResponse(w, res)
		return
	}

	logger.Infof("getTransferHistory: start is:%v, transfer direction is:%v, account is:%v, verification code is:%v", startBlkNum, dir, acctName, paramsInfo.verificationCode)
	isSender := false
	if dir == types.TxDirectionSend {
		isSender = true
	}
	model := db.GetTransferRecord(startBlkNum, acctName, isSender)
	if model.Err != nil {
		res.Status = model.ErrCode
		res.Msg = model.Err.Error()
	} else {
		res.Status = types.StatusSuccess
		res.HeadBlockHeight = strconv.FormatUint(model.Lib, 10)
		res.MaxBlockHeight = strconv.FormatUint(model.MaxQueryBlkNum, 10)
		if len(model.List) > 0 {
			res.List = model.List
		}
	}
	writeResponse(w, res)
}

func getTransferHistoryOfBlock(w http.ResponseWriter, r *http.Request)  {
	logger := logs.GetLogger()
	res := types.SingleBlockTransferHistoryResponse{
		List: make([]*types.TransferRecord,0),
	}
	paramsInfo := parseHistoryParams(r)
	if paramsInfo.err != nil {
		res.Status = paramsInfo.errCode
		res.Msg = paramsInfo.err.Error()
		writeResponse(w, res)
		return
	}
	// parse block number param
	blkNum,err,code := parseBlockNumberByKey(r, singleBlockKey)
	if err != nil {
		res.Status = code
		res.Msg = err.Error()
		writeResponse(w, res)
		return
	}

	logger.Infof("getTransferHistoryByBlock: block is:%v, account is:%v, verification code is:%v", blkNum, paramsInfo.account, paramsInfo.verificationCode)
	isSender := false
	if paramsInfo.txDirection != types.TxDirectionSend {
		isSender = false
	}
	model := db.GetUserTransferRecordByBlock(blkNum, paramsInfo.account, isSender)
	if model.Err != nil {
		res.Status = model.ErrCode
		res.Msg = model.Err.Error()
	} else {
		res.Status = types.StatusSuccess
		if len(model.List) > 0 {
			res.List = model.List
		}
	}
	writeResponse(w, res)
}

func writeResponse(w http.ResponseWriter, data interface{}) {
	js, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Fail to marshal json", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _,err := w.Write(js); err != nil {
		log := logs.GetLogger()
		log.Errorf("w.Write fail, json is %v, error is %v \n", string(js), err)
		http.Error(w, "Fail to write json", types.StatusIntervalError)
	}

}

func parseHistoryParams(r *http.Request) historyParamsModel {
	model := historyParamsModel{}
	logger := logs.GetLogger()
	//Get Verification Code
	vCode,err,code := parseParameterFromRequest(r, verificationCodeKey)
	if err != nil {
		model.err = err
		model.errCode = code
		return model
	}
	if !config.CheckIsValidVerificationCode(vCode) {
		msg := fmt.Sprintf("verification code %v is invalid", vCode)
		model.err = errors.New(msg)
		model.errCode = types.StatusParamVerificationCodeInvalidError
		return model
	}

	//Get transfer direction
	dirStr,err,code := parseParameterFromRequest(r, txDirectionKey)
	if err != nil {
		model.err = err
		model.errCode = code
		return model
	}

	dir,err := strconv.Atoi(dirStr)
	if err != nil {
		msg := fmt.Sprintf("fail to parse transfer direction %v", dirStr)
		model.err = errors.New(msg)
		model.errCode = types.StatusGetTransferRecordError
		return model
	}
	if dir != types.TxDirectionSend && dir != types.TxDirectionReceive {
		// invalid transfer direction
		logger.Errorf("getTransferHistory: transfer direction %v is invalid", dir)
		msg := fmt.Sprintf("transfer direction %v is invalid", dir)
		model.err = errors.New(msg)
		model.errCode = types.StatusParamTransferDirectionInvalidError
		return model
	}

	// get account param
	acctName,err,code := parseParameterFromRequest(r, accountNameKey)
	if err != nil {
		model.err = err
		model.errCode = code
		return model
	}
	model.account = acctName
	model.txDirection = dir
	model.verificationCode = vCode
	return model
}

func parseBlockNumberByKey(r *http.Request, pKey string) (uint64,error,int) {
	blkStr,err,code := parseParameterFromRequest(r, pKey)
	if err != nil {
		return 0,err,code
	}
	blkNum,err := strconv.ParseUint(blkStr, 10, 64)
	if err != nil {
		msg := fmt.Sprintf("fail to parse block param,%v", err)
		err := errors.New(msg)
		return 0, err, types.StatusParamInvalidError
	}
	return blkNum,nil,0
}

func parseParameterFromRequest(r *http.Request, parameter string) (string,error,int) {
	var (
		err error
		errCode int
	)

	if r == nil {
		return "", errors.New("empty http request"), types.StatusParamInvalidError
	}
	reqMethod := r.Method
	//just handle POST and Get Method
	if reqMethod == http.MethodPost || reqMethod == http.MethodGet {
		if reqMethod == http.MethodGet {
			queryForm, err := url.ParseQuery(r.URL.RawQuery)
			if err == nil && len(queryForm[parameter]) > 0  && utils.CheckIsNotEmptyStr(queryForm[parameter][0]){
				return queryForm[parameter][0], err, http.StatusOK
			} else {
				return "", errors.New(fmt.Sprintf("lack parameter %v", parameter)), types.StatusLackParamError
			}
		} else {
			err = r.ParseForm()
			if err != nil {
				return "", err, types.StatusParamInvalidError
			}
			val := r.PostFormValue(parameter)
			if len(val) < 1 {
				return "", errors.New(fmt.Sprintf("lack parameter %v", parameter)), types.StatusLackParamError
			}
			return val, nil, http.StatusOK
		}

	} else {
		err = errors.New(fmt.Sprintf("Not support %v method", reqMethod))
		errCode = http.StatusMethodNotAllowed
	}
	return "", err, errCode
}