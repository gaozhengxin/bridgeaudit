package params

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/log"
)

var (
	configFile string
	scanConfig = &ScanConfig{}
)

// ScanConfig scan config
type ScanConfig struct {
	Tokens []*TokenConfig
}

// TokenConfig token config
type TokenConfig struct {
	PairID         string
	SwapServer     string
	TokenAddress   string
	DepositAddress string   `toml:",omitempty" json:",omitempty"`
	LogTopics      []string `toml:",omitempty" json:",omitempty"`
}

// IsNativeToken is native token
func (c *TokenConfig) IsNativeToken() bool {
	return c.TokenAddress == "native"
}

// GetScanConfig get scan config
func GetScanConfig() *ScanConfig {
	return scanConfig
}

// LoadConfig load config
func LoadConfig(filePath string) *ScanConfig {
	log.Println("LoadConfig Config file is", filePath)
	if !common.FileExist(filePath) {
		log.Fatalf("LoadConfig error: config file '%v' not exist", filePath)
	}

	config := &ScanConfig{}
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		log.Fatalf("LoadConfig error (toml DecodeFile): %v", err)
	}

	var bs []byte
	if log.JSONFormat {
		bs, _ = json.Marshal(config)
	} else {
		bs, _ = json.MarshalIndent(config, "", "  ")
	}
	log.Println("LoadConfig finished.", string(bs))

	if err := config.CheckConfig(); err != nil {
		log.Fatalf("LoadConfig Check config failed. %v", err)
	}

	configFile = filePath // init config file path
	scanConfig = config   // init scan config
	return scanConfig
}

// ReloadConfig reload config
func ReloadConfig() {
	log.Println("ReloadConfig Config file is", configFile)
	if !common.FileExist(configFile) {
		log.Errorf("ReloadConfig error: config file '%v' not exist", configFile)
		return
	}

	config := &ScanConfig{}
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		log.Errorf("ReloadConfig error (toml DecodeFile): %v", err)
		return
	}

	if err := config.CheckConfig(); err != nil {
		log.Errorf("ReloadConfig Check config failed. %v", err)
		return
	}
	log.Println("ReloadConfig success.")
	scanConfig = config // reassign scan config
}

// CheckConfig check scan config
func (c *ScanConfig) CheckConfig() (err error) {
	if len(c.Tokens) == 0 {
		return errors.New("no token config exist")
	}
	pairIDMap := make(map[string]struct{})
	tokensMap := make(map[string]struct{})
	exist := false
	for _, tokenCfg := range c.Tokens {
		err = tokenCfg.CheckConfig()
		if err != nil {
			return err
		}
		pairID := strings.ToLower(tokenCfg.PairID)
		if _, exist = pairIDMap[pairID]; exist {
			return errors.New("duplicate pairID " + pairID)
		}
		pairIDMap[pairID] = struct{}{}
		if !tokenCfg.IsNativeToken() {
			tokenAddr := strings.ToLower(tokenCfg.TokenAddress)
			if _, exist = tokensMap[tokenAddr]; exist {
				return errors.New("duplicate token address " + tokenAddr)
			}
			tokensMap[tokenAddr] = struct{}{}
		}
	}
	return nil
}

// CheckConfig check token config
func (c *TokenConfig) CheckConfig() error {
	if c.PairID == "" {
		return errors.New("empty 'PairID'")
	}
	if c.SwapServer == "" {
		return errors.New("empty 'SwapServer'")
	}
	if !c.IsNativeToken() && !common.IsHexAddress(c.TokenAddress) {
		return errors.New("wrong 'TokenAddress' " + c.TokenAddress)
	}
	if c.DepositAddress != "" && !common.IsHexAddress(c.DepositAddress) {
		return errors.New("wrong 'DepositAddress' " + c.DepositAddress)
	}
	for _, topic := range c.LogTopics {
		if common.HexToHash(topic).Hex() != topic {
			return errors.New("wrong 'LogTopics' " + topic)
		}
	}
	return nil
}
