package utils

import (
	"fmt"

	"github.com/avast/retry-go"
	"go-micro.dev/v4"
)

func ServiceRetryGet(service micro.Service, svcName string, attempts uint) (string, error) {
	r := service.Options().Registry

	var (
		hostAndPort string
	)

	err := retry.Do(
		func() error {
			services, err := r.GetService(svcName)
			if err == nil {
				for _, s := range services {
					for _, n := range s.Nodes {
						hostAndPort = n.Address
						break
					}
					if hostAndPort != "" {
						break
					}
				}
			}

			if hostAndPort == "" {
				return fmt.Errorf("Service %v not found", svcName)
			}

			return nil
		},
		retry.Attempts(attempts),
	)

	if err != nil {
		return "", err
	}

	return hostAndPort, nil
}
