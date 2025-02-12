package confs

import (
	"encoding/json"
	"os"

	"github.com/NoelM/minigo"
)

type NotelConf struct {
	CommuneDbPath  string          `json:"communeDbPath"`
	MessagesDbPath string          `json:"messagesDbPath"`
	UsersDbPath    string          `json:"usersDbPath"`
	BlogDbPath     string          `json:"blogDbPath"`
	AnnuaireDbPath string          `json:"annuaireDbPath"`
	Connectors     []ConnectorConf `json:"connectors"`
}

type ConnectorConf struct {
	Active bool               `json:"active"`
	Kind   string             `json:"kind"`
	Tag    string             `json:"tag"`
	Path   string             `json:"path"`
	Config []minigo.ATCommand `json:"config,omitempty"`
}

func LoadConfig(path string) (*NotelConf, error) {
	configData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &NotelConf{
		Connectors: make([]ConnectorConf, 0),
	}

	if err = json.Unmarshal(configData, config); err != nil {
		return nil, err
	}

	return config, nil
}
