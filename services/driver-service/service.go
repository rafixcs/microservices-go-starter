package main

import (
	"fmt"
	math "math/rand/v2"
	pb "ride-sharing/shared/proto/driver"
	"slices"
	"sync"

	"github.com/mmcloughlin/geohash"
)

type Service struct {
	drivers []*driverInMap
	mu      sync.RWMutex
}

type driverInMap struct {
	Driver *pb.Driver
	//Index  int
	// TODO: route
}

func NewService() *Service {
	return &Service{
		drivers: make([]*driverInMap, 0),
	}
}

func (s *Service) FindAvailableDrivers(packageType string) []string {
	var matchingDrivers []string

	for _, driver := range s.drivers {
		if driver.Driver.PackageSlug == packageType {
			matchingDrivers = append(matchingDrivers, driver.Driver.ID)
		}
	}

	if len(matchingDrivers) == 0 {
		return []string{}
	}

	return matchingDrivers
}

// TODO: Register and unregister methods
func (s *Service) RegisterDriver(id string, packageSlug string) (*pb.Driver, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.drivers == nil {
		s.drivers = make([]*driverInMap, 0)
	}

	randomIndx := math.IntN(len(PredefinedRoutes))
	randomRoute := PredefinedRoutes[randomIndx]
	geohashs := geohash.Encode(randomRoute[0][0], randomRoute[0][1])

	var driver driverInMap

	idx := slices.IndexFunc(s.drivers, func(d *driverInMap) bool {
		return d.Driver.ID == id
	})

	if idx != -1 {
		return nil, fmt.Errorf("driver id %s already registered", id)
	} else {
		driver = driverInMap{
			Driver: &pb.Driver{
				Geohash: geohashs,
				Location: &pb.Location{
					Latitude:  randomRoute[0][0],
					Longitude: randomRoute[0][1],
				},
				Name:        "Luke Skywalker",
				ID:          id,
				PackageSlug: packageSlug,
			},
		}
		s.drivers = append(s.drivers, &driver)
	}

	return driver.Driver, nil
}

func (s *Service) UnRegisterDriver(id string, packageSlug string) (*pb.Driver, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.drivers == nil {
		return nil, fmt.Errorf("[driver.Service.UnRegisterDriver] empty drivers database")
	}

	idx := slices.IndexFunc(s.drivers, func(d *driverInMap) bool {
		return d.Driver.ID == id
	})

	copyDriver := *s.drivers[idx]

	if idx == -1 {
		return nil, fmt.Errorf("[driver.Service.UnRegisterDriver] driver id %s not found", id)
	}

	s.drivers = slices.Delete(s.drivers, idx, idx+1)

	return copyDriver.Driver, nil
}
