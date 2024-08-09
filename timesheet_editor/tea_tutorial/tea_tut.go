package main

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

func main() {
	// open db. If not exists, create
	db, err := bolt.Open("./my.db", 0600, nil) // 0600 - read-write permission
	if err != nil {
		panic(err)
}


// create bucket, keys, values
err = db.Update(func(tx *bolt.Tx) error {
	b, err := tx.CreateBucketIfNotExists([]byte("2024-August-06")) // create date bucket
	if err != nil {
		return err
	}
	err = b.Put([]byte("Did a thing"), []byte("7")) // create line entry for date
	if err != nil {
		return err
	}
	err = b.Put([]byte("Worked some overtime"), []byte("2.5")) // create line entry for date
	if err != nil {
		return err
	}

	return nil
})
if err != nil {
	panic(err)
}

// View bucket. Iterate over entries. Count entries
db.View(func(tx *bolt.Tx) error {
	b := tx.Bucket([]byte("2024-August-06"))
	
	c := b.Cursor()
	i := 0
	for k,v := c.First(); k != nil; k,v = c.Next() {
		i ++
		fmt.Printf("key=%s, value=%s\n", k, v)
	}

	fmt.Printf("Found %d entries\n", i)
	return nil
})

defer db.Close()
}