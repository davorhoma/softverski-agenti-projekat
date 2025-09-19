package aktorski_framework

import (
	"encoding/json"
	"fmt"

	"go.etcd.io/bbolt"
)

type PersistenceActor struct {
	db *bbolt.DB
}

func NewPersistenceActor(dbPath string) (*PersistenceActor, error) {
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	return &PersistenceActor{db: db}, nil
}

func (a *PersistenceActor) Stopped() {
	if a.db != nil {
		err := a.db.Close()
		if err != nil {
			fmt.Printf("[PersistenceActor] Failed to close database: %v\n", err)
		} else {
			fmt.Println("[PersistenceActor] Database closed successfully.")
		}
	}
}

func (a *PersistenceActor) Receive(ctx *Context, msg Message) {
	switch m := msg.(type) {
	case LoadStateRequest:
		var data []byte
		err := a.db.View(func(tx *bbolt.Tx) error {
			bucket := tx.Bucket([]byte("states"))
			if bucket == nil {
				return nil
			}
			data = bucket.Get([]byte(m.Key))
			return nil
		})

		if err != nil {
			return
		}
		m.ReplyChan <- data

	case SaveStateMessage:
		data, err := json.Marshal(m.State)
		if err != nil {
			return
		}

		err = a.db.Update(func(tx *bbolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists([]byte("states"))
			if err != nil {
				return err
			}
			return bucket.Put([]byte(m.Key), data)
		})

		if err != nil {
		}
	}
}
