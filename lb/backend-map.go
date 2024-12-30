package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type BackendMap struct {
	Backends []struct {
		ID       string `yaml:"id"`
		Services []struct {
			BoxID     string `yaml:"box_id"`
			Host      string `yaml:"host"`
			ServiceID string `yaml:"service_id"`
		} `yaml:"services"`
	} `yaml:"backends"`
	Layout []struct {
		ID       string `yaml:"id"`
		Name     string `yaml:"name"`
		Services []struct {
			Description string `yaml:"description"`
			ID          string `yaml:"id"`
			Name        string `yaml:"name"`
			Proxy       string `yaml:"proxy"`
		} `yaml:"services"`
	} `yaml:"layout"`
	LbEndpoint string `yaml:"lb_endpoint"`
}

func parseBackendMap(file string, backendMap **BackendMap) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, &backendMap)
	if err != nil {
		return err
	}
	return nil
}

//	func getHostPort(instanceId string, boxId string, serviceId string) (string, error) {
//		if backendMap == nil {
//			fmt.Println("backendMap is nil")
//			return "", nil
//		}
//		// TODO optimise this using a map
//		for _, backend := range backendMap.Backends {
//			if backend.ID == instanceId {
//				for _, service := range backend.Services {
//					if service.BoxID == boxId && service.ServiceID == serviceId {
//						return service.Host, nil
//					}
//				}
//			}
//		}
//		return "", nil
//	}
func (backendMap *BackendMap) getProxyType(boxId string, serviceId string) (string, error) {
	if backendMap == nil {
		fmt.Println("backendMap is nil")
		return "", nil
	}
	// TODO optimise this using a map
	for _, layout := range backendMap.Layout {
		if layout.ID == boxId {
			for _, service := range layout.Services {
				if service.ID == serviceId {
					return service.Proxy, nil
				}
			}
		}
	}
	return "", nil
}
