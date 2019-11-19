package internal

import (
  "os"
  "io/ioutil"

  "gopkg.in/yaml.v2"

	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/osfs"

  "gopkg.in/src-d/go-git.v4/utils/merkletrie/filesystem"
  "gopkg.in/src-d/go-git.v4/utils/merkletrie/noder"
)

const configPath string = "/.gimini/gimini.yaml"

type config struct {
  Paths []string
}

func defaultConfig () config {
  config := config{}
  return config
}

func getConfig () (config, error) {
  config := config{}

  // construct full path
	home, err := os.UserHomeDir()
	if err != nil {
		return config, err
	}
	fullConfigPath := home + configPath

  // read file
  data, err := ioutil.ReadFile(fullConfigPath)
  if os.IsNotExist(err) {
    // create default
    config, err = defaultConfig(), nil
  } else if err != nil {
    return config, err
  } else {
    // unmarshall
    err = yaml.Unmarshal([]byte(data), &config)
    if err != nil {
      return config, err
    }
  }

  return config, nil
}

func (c *config) save () error {
  // construct full path
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	fullConfigPath := home + configPath

  // marshal to yaml
  data, err := yaml.Marshal(c)
  if err != nil {
    return err
  }

  return ioutil.WriteFile(fullConfigPath, data, 0644)
}

func (c *config) add (path string) error {
  if isInSlice(c.Paths, path) {
    return nil
  }

  c.Paths = append(c.Paths, path)
  return c.save()
}

func isInSlice(ss []string, v string) bool {
  for _, s := range ss {
    if s == v {
      return true
    }
  }
  return false
}

func (c config) getFilesystemNodes(fs billy.Filesystem) ([]noder.Noder, error) {
  var nodes []noder.Noder

	to := filesystem.NewRootNode(osfs.New("/"), nil)
  nodes = append(nodes, to)

  return nodes, nil
}
