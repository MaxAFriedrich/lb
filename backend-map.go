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

func (backendMap *BackendMap) proxyPathToHost() map[string]string {
	proxyPathToHost := make(map[string]string)
	for _, backend := range backendMap.Backends {
		for _, service := range backend.Services {
			proxyPath := fmt.Sprintf("/%s/%s/%s/", backend.ID, service.BoxID, service.ServiceID)
			proxyPathToHost[proxyPath] = service.Host
		}
	}
	return proxyPathToHost
}

func (backendMap *BackendMap) proxyPathToType() map[string]string {
	proxyPathToType := make(map[string]string)
	for _, backend := range backendMap.Backends {
		for _, service := range backend.Services {
			proxyPath := fmt.Sprintf("/%s/%s/%s/", backend.ID, service.BoxID, service.ServiceID)
			proxyType, err := backendMap.getProxyType(service.BoxID, service.ServiceID)
			if err != nil {
				panic(err)
			}
			proxyPathToType[proxyPath] = proxyType
		}
	}
	return proxyPathToType
}

func (backendMap *BackendMap) getProxyType(boxId string, serviceId string) (string, error) {
	if backendMap == nil {
		fmt.Println("backendMap is nil")
		return "", nil
	}
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
