package travis

import (
	"bytes"
	"encoding/gob"
	"log"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

const repoSubscriptionKey = "repo_subscriptions"

type Updater struct {
	GitHubAccessToken string
	RedisURL          string
}

func (u Updater) Run() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	c, err := redis.DialURL(u.RedisURL)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	for {
		cutoff := int32(time.Now().Add(time.Minute * -1).Unix())
		n, err := redis.Strings(c.Do("ZRANGEBYSCORE", repoSubscriptionKey, cutoff, "+inf"))
		if err != nil {
			log.Fatal(err)
		}

		for _, repoSlug := range n {
			log.Println(string(repoSlug))
			_, err := u.updateFromGitHub(repoSlug)
			if err != nil {
				log.Println(err)
				continue
			}
		}

		// delete old subscriptions
		_, err = c.Do("ZREMRANGEBYSCORE", repoSubscriptionKey, 0, cutoff)
		if err != nil {
			log.Fatal(err)
		}

		time.Sleep(time.Minute)
	}
}

func (u Updater) updateFromGitHub(repoSlug string) (Repo, error) {
	repo := Repo{}
	c, err := redis.DialURL(u.RedisURL)
	if err != nil {
		return repo, err
	}
	defer c.Close()

	splitted := strings.Split(repoSlug, "/")
	tc := Client{Org: splitted[0], Repo: splitted[1], GitHubAccessToken: u.GitHubAccessToken}
	repo, err = tc.Repository()
	if err != nil {
		return repo, err
	}

	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(repo)
	if err != nil {
		return repo, err
	}

	key := "github-repo-" + repoSlug
	_, err = c.Do("SET", key, buf.Bytes(), "EX", 3600) // last is expiration seconds
	if err != nil {
		return repo, err
	}

	return repo, nil
}

func (u Updater) SubscribeRepo(repoSlug string) {
	c, err := redis.DialURL(u.RedisURL)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	timestamp := int32(time.Now().Unix())
	_, err = c.Do("ZADD", repoSubscriptionKey, timestamp, repoSlug)
	if err != nil {
		log.Fatal(err)
	}
}

func (u Updater) GetRepo(repoSlug string) (Repo, error) {
	repo := Repo{}
	c, err := redis.DialURL(u.RedisURL)
	if err != nil {
		return repo, err
	}
	defer c.Close()

	key := "github-repo-" + repoSlug
	b, err := redis.Bytes(c.Do("GET", key))
	if err == redis.ErrNil {
		log.Printf("fetching %v on demand\n", repoSlug)
		repo, err = u.updateFromGitHub(repoSlug)
		if err != nil {
			return repo, err
		}
		return repo, nil
	} else if err != nil {
		return repo, err
	}

	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&repo)
	return repo, err
}
