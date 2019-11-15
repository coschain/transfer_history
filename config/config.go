package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
)

const (
	EnvDev  = "dev"
	EnvPro  = "pro"
	EnvTest = "test"
)


type FullNodeDbInfo struct {
	FullNodeDbDriver   string `json:"fullNodeDbDriver"`
	FullNodeDbUser     string `json:"fullNodeDbUser"`
	FullNodeDbPassword string `json:"fullNodeDbPassword"`
	FullNodeDbName     string `json:"fullNodeDbName"`
	FullNodeDbHost     string `json:"fullNodeDbHost"`
	FullNodeDbPort     string `json:"fullNodeDbPort"`
}

type EnvConfig struct {
	HttpPort      string      `json:"httpPort"`
	LogPath         string    `json:"logPath"`
	FullNodeDbList  []FullNodeDbInfo `json:"fullNodeDbList"`
	VerificationCodeList []string `json:"verificationCodeList"`
}

type serviceConfig struct {
	Dev    EnvConfig     `json:"dev"`
	Pro    EnvConfig     `json:"pro"`
	Test   EnvConfig     `json:"test"`
}

type DbConfig struct {
	Driver string
	User   string
	Password string
	Host     string
	Port     string
	DbName   string
}


var (
	svConfig *EnvConfig
	configOnce sync.Once
	env = EnvDev // default env is dev
	httpPort = "8000" //default http port of web server
)


// read config json file
func LoadExchangeTransferHistoryConfig(path string) error {
	if svConfig != nil {
		return nil
	}
	var config serviceConfig
	p,err := filepath.Abs(path)
	if err != nil {
		return err
	}
	fmt.Printf("config path is %v \n", p)
	cfgJson,err := ioutil.ReadFile(p)
	if err != nil {
		fmt.Printf("LoadExchangeTransferHistoryConfig:fail to read json file, the error is %v \n", err)
		return err
	} else {
		if err := json.Unmarshal(cfgJson, &config); err != nil {
			fmt.Printf("LoadExchangeTransferHistoryConfig: fail to  Unmarshal json, the error is %v \n", err)
		} else {
			if IsDevEnv() {
				svConfig = &config.Dev
			} else if IsTestEnv() {
				svConfig = &config.Test
			} else if IsProEnv(){
				svConfig = &config.Pro
			} else {
				return errors.New("fail to get reward config of unKnown env")
			}
		}
	}
	return nil
}

func SetConfigEnv(ev string) error{
	if ev != EnvPro && ev != EnvDev && ev != EnvTest {
		return errors.New(fmt.Sprintf("Fail to set unknown environment %v", ev))
	}
	configOnce.Do(func() {
		env = ev
	})
	return nil
}

func IsDevEnv() bool {
	return env == EnvDev
}

func IsTestEnv() bool {
	return env == EnvTest
}

func IsProEnv() bool {
	return env == EnvPro
}

func GetHttpPort() string {
	return svConfig.HttpPort
}

// get log output path 
func GetLogOutputPath() string {
	if svConfig != nil {
		return svConfig.LogPath
	}
	return ""
}


// get cos observe node database config list
func GetCosFullNodeDbConfigList() ([]*DbConfig, error) {
	var list []*DbConfig
	if svConfig != nil {
		for _,cf := range svConfig.FullNodeDbList {
			info := &DbConfig{}
			info.Driver = cf.FullNodeDbDriver
			info.User = cf.FullNodeDbUser
			info.Password = cf.FullNodeDbPassword
			info.Port= cf.FullNodeDbPort
			info.Host = cf.FullNodeDbHost
			info.DbName = cf.FullNodeDbName
			list = append(list, info)
		}

	} else {
		return nil, errors.New("can't get observe db config from empty reward config")
	}
	return list,nil
}

func GetVerificationCodeList() []string {
	if svConfig != nil {
		return svConfig.VerificationCodeList
	}
	return nil
}

func CheckIsValidVerificationCode(code string) bool {
	if svConfig != nil {
		for _,vCode := range svConfig.VerificationCodeList {
			if code == vCode {
				return true
			}
		}
	}
	return false
}