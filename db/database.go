package db

import (
	"errors"
	"fmt"
	"github.com/coschain/contentos-go/app/plugins"
	"github.com/exchange-service/transfer_history/config"
	"github.com/exchange-service/transfer_history/logs"
	"github.com/exchange-service/transfer_history/types"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"strconv"
	"time"
)

var (
	cosNodeDb *gorm.DB
	cosNodeDbHost string
	curBlockHeight uint64
	checkInterval = 2 * time.Minute
	stop  chan bool
	isChecking bool
)

func StartDbService() error {
	logger := logs.GetLogger()
	logger.Debugln("Start db service")
	nodeDb,err := getCosFullNodeDb()
	if err != nil {
		logger.Errorf("StartDbService: fail to get cos observe node db,the error is %v", err)
		return err
	}
	cosNodeDb = nodeDb
	checkCosNodeDbValid()
    return nil
}

func getCosFullNodeDb() (*gorm.DB, error) {
	if cosNodeDb != nil {
		return cosNodeDb,nil
	}
	logger := logs.GetLogger()
	list,err := config.GetCosFullNodeDbConfigList()
	if err != nil {
		logger.Errorf("GetCosObserveNodeDb: fail to get cos observe node db config, the error is %v", err)
		return nil, errors.New("open db: fail to get observe node db config")
	}
	var dbErr error
	for _,cf := range list {
		db,err := openDb(cf)
		if err != nil {
			logger.Errorf("GetCosObserveNodeDb: fail to open db, the error is %v", err)
			dbErr = err
		} else if db != nil {
			cosNodeDbHost = cf.Host
			cosNodeDb = db
			return db,nil
		}
	}
	return nil, dbErr
}

func openDb(dbCfg *config.DbConfig) (*gorm.DB, error) {
	log := logs.GetLogger()
	source := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbCfg.User, dbCfg.Password, dbCfg.Host, dbCfg.Port,dbCfg.DbName)
	db,err := gorm.Open(dbCfg.Driver, source)
	if err != nil {
		log.Errorf("openDb: fail to open db: %v, the error is %v ", dbCfg, err)
		return nil,errors.New("fail to open db")
	}
	return db,nil
}

// Timing check the database status regularly(check block height change)
func checkCosNodeDbValid()  {
	ticker := time.NewTicker(checkInterval)
	go func() {
		for {
			select {
			case <- ticker.C:
				checkBlockStatus()
			case <- stop:
				ticker.Stop()
			}

		}
	}()
}

func checkBlockStatus()  {
	logger := logs.GetLogger()
	if isChecking {
		logger.Infoln("last round check block status not finish")
		return
	}
	isChecking = true
	defer func() {
		isChecking = false
	}()
	logger.Infoln("start check block status")
	if cosNodeDb != nil {
		var process plugins.BlockLogProcess
		err := cosNodeDb.Take(&process).Error
		if err != nil {
			logger.Errorf("checkBlockStatus: fail to get cos chain block process")
		} else {
			if  process.BlockHeight <= curBlockHeight {
				logger.Infof("new block height is %v, cache block height is %v", process.BlockHeight, curBlockHeight)
				logger.Infof("checkBlockStatus: Need to switch to another cos observe node db")
				//need to switch full node db
				list,err := config.GetCosFullNodeDbConfigList()
				if err != nil {
					logger.Errorf("checkBlockStatus: fail to get db config list, the error is %v", err)
				} else {
					for _,cf := range list {
						if cf.Host != cosNodeDbHost {
							db,err := openDb(cf)
							if err == nil {
								logger.Infof("checkBlockStatus: success to switch origin cos node db:%v to new db:%v", cosNodeDbHost, cf.Host)
								cosNodeDb = db
								cosNodeDbHost = cf.Host
								break
							} else {
								logger.Errorf("checkBlockStatus: fail to switch new db, the error is %v", err)
							}
						}
					}
				}
			}
			curBlockHeight = process.BlockHeight
		}

	}
	logger.Infoln("finish this round check block status")

}

func CloseDbService() {
	logger := logs.GetLogger()
	logger.Infoln("Close my sql database")

	if cosNodeDb != nil {
		if err := cosNodeDb.Close(); err != nil {
			logger.Errorf("Fail to close cos observe node db, the error is %v", err)
		}
	}
}

func getLib(db *gorm.DB) (uint64,error) {
	logger := logs.GetLogger()
	if db == nil {
		return 0, errors.New("GetLib: db instance is empty")
	}
	var libInfo types.LibInfo
	err := db.Table(types.LibTableName).Find(&libInfo).Error
	if err != nil {
		logger.Errorf("GetLib: fail to get lib, the error is %v", err)
		return 0, errors.New("GetLib: fail to get current lib")
	}
	return libInfo.Lib,nil
}


//get transfer record of account, if isSender = true, get send out record or get deposit record
func GetTransferRecord(sBlkNum uint64, acct string, isSender bool) *types.QueryTransferRecordModel {
	logger := logs.GetLogger()
	cosDb, err := getCosFullNodeDb()
	model := &types.QueryTransferRecordModel{
		List: make([]*types.TransferRecord, 0),
	}
	if err != nil {
		logger.Errorf("GetTransferRecord: fail to get cos full node db,the error is %v", err)
		model.Err = errors.New("system error,fail to open full node db")
		model.ErrCode = types.StatusIntervalError
		return model
	}
	//1. get current lib
	lib,err := getLib(cosDb)
	if err != nil {
		logger.Errorf("GetTransferRecord: fail to lib,the error is %v", err)
		model.Err = errors.New("fail to get lib")
		model.ErrCode = types.StatusGetLibError
        return model
	}
	model.Lib = lib
	var (
		tList []*plugins.TransferRecord
		recList []*types.TransferRecord
		maxBlkNum uint64
	)
	//2. get max block height in transfer record db
    err = cosDb.Model(plugins.TransferRecord{}).Select("max(block_height)").Row().Scan(&maxBlkNum)
    if err != nil {
    	if err != gorm.ErrRecordNotFound {
			logger.Errorf("GetTransferRecord: fail to get max block height in transfer record,the error is %v", err)
			model.Err = errors.New("fail to get head block height of transfer record")
			model.ErrCode = types.StatusGetTransferRecordError
			return model
		}
	}
	if lib < maxBlkNum {
		lib = maxBlkNum
	}
	model.MaxQueryBlkNum = maxBlkNum
	if lib < sBlkNum  || maxBlkNum < sBlkNum{
		return model
	}
	//3. get transfer record
	filter := fmt.Sprintf("block_height >= %v AND block_height <= %v AND `from` = \"%v\"", sBlkNum, maxBlkNum, acct)
	if !isSender {
		//get deposit record
		filter = fmt.Sprintf("block_height >= %v AND block_height <= %v AND `to` = \"%v\"", sBlkNum, maxBlkNum, acct)
	}
	err = cosNodeDb.Model(plugins.TransferRecord{}).Where(filter).Order("block_height ASC").Find(&tList).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			// not found record
			return model
		}
		model.Err = errors.New("fail to get transfer record")
		model.ErrCode = types.StatusGetTransferRecordError
		return model
	}
	listLen := len(tList)
	if listLen > 0 {
		for _,rec := range tList {
			rec := &types.TransferRecord{
				OperationId: rec.OperationId,
				From: rec.From,
				To: rec.To,
				Memo: rec.Memo,
				Amount: strconv.FormatUint(rec.Amount, 10),
				BlockHeight: strconv.FormatUint(rec.BlockHeight, 10),
			}
			recList = append(recList, rec)
		}
		model.List = recList
	}
	model.Lib = lib
	return model
}

// get transfer record
func GetUserTransferRecordByBlock(blkNum uint64, acct string, isSender bool) *types.QueryTransferRecordModel {
	logger := logs.GetLogger()
	cosDb, err := getCosFullNodeDb()
	model := &types.QueryTransferRecordModel{
		List: make([]*types.TransferRecord, 0),
	}
	if err != nil {
		logger.Errorf("GetUserTransferRecordByBlock: fail to get cos full node db,the error is %v", err)
		model.Err = errors.New("system error,fail to open full node db")
		model.ErrCode = types.StatusIntervalError
		return model
	}
	var (
		tList []*plugins.TransferRecord
		recList []*types.TransferRecord
	)
	filter := fmt.Sprintf("block_height = %v AND `from` = \"%v\" ", blkNum, acct)
	if !isSender {
		filter = fmt.Sprintf("block_height = %v AND `to` = \"%v\" ", blkNum, acct)
	}
	err = cosDb.Model(plugins.TransferRecord{}).Where(filter).Find(&tList).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			logger.Errorf("GetTransferRecord: fail to get max block height in transfer record,the error is %v", err)
			model.Err = errors.New("fail to get head block height of transfer record")
			model.ErrCode = types.StatusGetTransferRecordError
			return model
		}
	}
	if len(tList) > 0 {
		for _,rec := range tList {
			rec := &types.TransferRecord{
				OperationId: rec.OperationId,
				From: rec.From,
				To: rec.To,
				Memo: rec.Memo,
				Amount: strconv.FormatUint(rec.Amount, 10),
				BlockHeight: strconv.FormatUint(rec.BlockHeight, 10),
			}
			recList = append(recList, rec)
		}
		model.List = recList
	}
	return model
}