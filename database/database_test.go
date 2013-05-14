package database

import (
	"runtime"
	"strconv"
	"sync"
	"testing"
)

import "fmt"

type TestSession struct {
}

func (ms TestSession) DB(dbName string) Database {
	return &TestDatabase{}
}

type TestDatabase struct {
}

func (md TestDatabase) C(collectionName string) Collection {
	return &TestCollection{}
}

type TestCollection struct {
}

func (mc TestCollection) Find(selector interface{}) Query {
	return &TestQuery{}
}

func (mc TestCollection) RemoveId(id interface{}) error {
	return nil
}

func (mc TestCollection) Remove(selector interface{}) error {
	return nil
}

func (mc TestCollection) DropCollection() error {
	return nil
}

func (mc TestCollection) UpdateId(id interface{}, change interface{}) error {
	return nil
}

func (mc TestCollection) UpsertId(id interface{}, change interface{}) error {
	return nil
}

type TestQuery struct {
}

func (mq TestQuery) Count() (int, error) {
	return 0, nil
}

func (mq TestQuery) One(result interface{}) error {
	return nil
}

func (mq TestQuery) Iter() Iterator {
	return &TestIterator{}
}

type TestIterator struct {
}

func (mi TestIterator) All(result interface{}) error {
	return nil
}

func Test_ThreadSafety(t *testing.T) {
	runtime.GOMAXPROCS(2)
	Init(&TestSession{})

	char := NewCharacter("test", "", "")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		for i := 0; i < 100; i++ {
			char.SetName(strconv.Itoa(i))
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < 100; i++ {
			fmt.Println(char.GetName())
		}
		wg.Done()
	}()

	wg.Wait()
}

// vim: nocindent
