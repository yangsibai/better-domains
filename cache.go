package main

import (
	"encoding/json"
	"gopkg.in/redis.v3"
	"time"
)

const availableDomainsCacheKey string = "all_available_domains"
const availableDomainsCacheExpireation time.Duration = 10 * time.Minute

// set available domains cache
func setAvailableDomainsCache(domains []string) error {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Addr,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})
	val, err := json.Marshal(domains)
	if err != nil {
		return err
	}
	err = client.Set(availableDomainsCacheKey, val, availableDomainsCacheExpireation).Err()
	return err
}

// get available domains cache
func getAvailableDomainsCache() (domains []string, err error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Addr,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})
	bytes, err := client.Get(availableDomainsCacheKey).Bytes()
	if err != nil {
		return
	}
	err = json.Unmarshal(bytes, &domains)
	return
}
