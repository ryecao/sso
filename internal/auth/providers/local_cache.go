package providers

import (
	"github.com/buzzfeed/sso/internal/pkg/groups"
	log "github.com/buzzfeed/sso/internal/pkg/logging"
	"github.com/datadog/datadog-go/statsd"
)

type Cache interface {
	Get(keys string)
	Set(entries []string)
	Purge(keys []string)
}

type GroupCache struct {
	StatsdClient *statsd.Client
	provider     *Provider
	Cache        *groups.LocalCache
}

//func (p *SingleFlightProvider) Data() *ProviderData {
//	return p.provider.Data()
//}

// ValidateGroupMembership wraps the provider's ValidateGroupMembership around calls to check local cache.
func (p *GroupCache) ValidateGroupMembership(email string, allowedGroups []string, accessToken string) ([]string, error) {
	logger := log.NewLogEntry()

	// Create new local cache
	p.Cache = groups.NewLocalCache(p.ProviderData.GroupsCacheTTL)

	// Try to pull group membership from cache.
	groupMembership, err := p.Cache.Get([]string{email})
	// If the user's group membership is not in cache, call p.Provider.ValidateGroupMembership.
	if len(groupMembership) == 0 {
		validGroups, groupMembership, err := p.Provider.ValidateGroupMembership(email, allowedGroups, accessToken, groupMembership)
		if err != nil {
			return nil, err
		}
		// Create the cache 'Entry' for the returned groupMembership object.
		entries := make([]groups.Entry, 0, 1)
		entries = append(entries, groups.Entry{
			Key:  email,
			Data: []byte(groupMembership),
		})
		// Cache the created entry and return the valid groups.
		_, err = p.Cache.Set(entries)
		if err != nil {
			logger.Error("unable to save valid emails to cache")
		}
		return validGroups, nil

	}
	// The user's group membership is in the cache, so pass it in to p.Provider.ValidateGroupMembership and return the valid groups.
	validGroups, _, err := p.Provider.ValidateGroupMembership(email, allowedGroups, accessToken, groupMembership)
	if err != nil {
		return nil, err
	}
	return validGroups, nil
}
