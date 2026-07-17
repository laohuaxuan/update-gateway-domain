package mse

import (
	"sync"

	"update-gateway-domain/internal/store"
)

type ClientManager struct {
	clients sync.Map
	store   *store.Store
}

func NewClientManager(store *store.Store) *ClientManager {
	return &ClientManager{store: store}
}

func (cm *ClientManager) GetClient(accountID string) (*Client, error) {
	if cached, ok := cm.clients.Load(accountID); ok {
		return cached.(*Client), nil
	}

	account, err := cm.store.GetAccountByID(accountID)
	if err != nil {
		return nil, err
	}

	client, err := NewClient(account.RegionID, account.AccessKeyID, account.AccessKeySecret, account.AcceptLanguage)
	if err != nil {
		return nil, err
	}

	cm.clients.Store(accountID, client)
	return client, nil
}

func (cm *ClientManager) RemoveClient(accountID string) {
	cm.clients.Delete(accountID)
}

func (cm *ClientManager) Clear() {
	cm.clients.Range(func(key, value interface{}) bool {
		cm.clients.Delete(key)
		return true
	})
}
