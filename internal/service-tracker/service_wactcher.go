package servicetracker

import (
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
)

type UpdateAction int

const (
	Add UpdateAction = iota
	Remove
)

type ServiceUpdate struct {
	Action  UpdateAction
	Address string
}

type ServiceTracker struct {
	mu                  sync.RWMutex
	serviceList         map[string]*service
	expirationInSeconds int
}

type service struct {
	addresses []string
	clients   map[string]chan ServiceUpdate
	info      map[string]time.Time
}

func NewServiceTracker(expiration int) *ServiceTracker {
	return &ServiceTracker{
		expirationInSeconds: expiration,
		serviceList:         make(map[string]*service),
	}
}

func (st *ServiceTracker) UpdateNode(serviceId, address string) {
	st.mu.Lock()
	defer st.mu.Unlock()

	serv, serviceExists := st.serviceList[serviceId]

	if !serviceExists {
		// Case where no client / serviceExists yet.
		st.serviceList[serviceId] = &service{
			addresses: []string{address},
			clients:   make(map[string]chan ServiceUpdate),
			info:      map[string]time.Time{address: time.Now()},
		}
		return
	}

	// The service is exists.

	_, nodeExits := serv.info[address]

	if !nodeExits {
		for _, channels := range serv.clients {
			channels <- ServiceUpdate{
				Action:  Add,
				Address: address,
			}
		}
	}

	// Update timestamp
	serv.info[address] = time.Now()

	// Only add to address list if it's not already there
	found := slices.Contains(serv.addresses, address)
	if !found {
		serv.addresses = append(serv.addresses, address)
	}
}

func (st *ServiceTracker) GetNode(serviceId string) (string, bool) {
	st.mu.Lock()
	defer st.mu.Unlock()

	serv, exists := st.serviceList[serviceId]

	if !exists || len(serv.addresses) == 0 {
		return "", false
	}

	expiration := st.expirationInSeconds

	// Calculate the "Dead" threshold
	expirationLimit := time.Now().Add(-time.Duration(expiration) * time.Second)

	// Attempt to find a valid node in the round-robin cycle
	for i := 0; i < len(serv.addresses); i++ {
		address := serv.addresses[0]
		serv.addresses = serv.addresses[1:] // Shift

		lastUpdated := serv.info[address]

		// If the node has sent a heartbeat within the last 30 seconds
		if lastUpdated.After(expirationLimit) {
			serv.addresses = append(serv.addresses, address) // Put back at the end
			return address, true
		}

		// If it's dead, we don't append it back, effectively removing it from rotation
		delete(serv.info, address)
		for _, channels := range serv.clients {
			channels <- ServiceUpdate{
				Action:  Remove,
				Address: address,
			}
		}
	}

	return "", false
}

func (st *ServiceTracker) AddClient(serviceId string) (string, chan ServiceUpdate) {
	st.mu.Lock()
	defer st.mu.Unlock()

	serv, serviceExists := st.serviceList[serviceId]

	channel := make(chan ServiceUpdate)

	clientId := uuid.New().String()

	if !serviceExists {
		clientMap := make(map[string]chan ServiceUpdate)
		clientMap[clientId] = channel

		st.serviceList[serviceId] = &service{
			addresses: make([]string, 0),
			clients:   clientMap,
			info:      make(map[string]time.Time),
		}
	} else {
		serv.clients[clientId] = channel
	}

	return clientId, channel
}

func (st *ServiceTracker) RemoveClient(clientId, serviceId string) {
	st.mu.Lock()
	defer st.mu.Unlock()

	serv := st.serviceList[serviceId]

	delete(serv.clients, clientId)
}

func (st *ServiceTracker) GetNodes(serviceId string) []string {

	st.mu.RLock()
	defer st.mu.Unlock()

	serv, exists := st.serviceList[serviceId]

	if !exists {
		return []string{}
	}

	var activeAddress []string

	expirationLimit := time.Now().Add(-time.Duration(st.expirationInSeconds) * time.Second)

	for key, lastUpdated := range serv.info {
		if lastUpdated.After(expirationLimit) {
			activeAddress = append(activeAddress, key)
		}
	}

	return activeAddress
}
