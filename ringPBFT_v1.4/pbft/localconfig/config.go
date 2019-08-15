package localconfig

import (
	"github.com/spf13/viper"
	"fmt"
)

const cmdRoot = "conf"

type TopLevel struct {
	Pbft Pbft
	Peer Peer
	Ledger Ledger
}
type Pbft struct {
	TotalNum   int
	PrimaryNum int
	TimeOut	   int64
	ViewChangeTimeOut int64
}

type Peer struct {
	NodeID int
	Port   int
}
type Ledger struct {
	Path string
}

func Load() (*TopLevel, error) {
	viper.SetConfigName(cmdRoot)
	viper.AddConfigPath("D:/GOPATH/src/github.com/ringPBFT/pbft/localconfig")
	err := viper.ReadInConfig() // Find and read the config file //读取配置文件
	if err != nil {             // Handle errors reading the config file
         fmt.Println(fmt.Errorf("Fatal error when reading %s config file: %s\n", cmdRoot, err))
    }
    var uconf TopLevel
    uconf.Peer.NodeID = viper.GetInt("Peer.NodeID")
    uconf.Peer.Port = viper.GetInt("Peer.Port")
    uconf.Pbft.TotalNum = viper.GetInt("Pbft.TotalNum")
    uconf.Pbft.PrimaryNum = viper.GetInt("Pbft.PrimaryNum")
    uconf.Pbft.TimeOut = viper.GetInt64("Pbft.TimeOut")
    uconf.Pbft.ViewChangeTimeOut = viper.GetInt64("Pbft.ViewChangeTimeOut")
    uconf.Ledger.Path = viper.GetString("Ledger.Path")
    return &uconf, nil
}
