package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/txn2/provision"
	yaml "gopkg.in/yaml.v2"
)

// PvObj generic object
type PvObj struct {
	Kind string      `json:"kind",yaml:"kind"`
	Spec interface{} `json:"spec",yaml:"spec"`
}

// PvObjAsset spec contains an Asset type
type PvObjAsset struct {
	PvObj
	Spec provision.Asset
}

// PvObjAccount spec contains an Account type
type PvObjAccount struct {
	PvObj
	Spec provision.Account
}

// PvObjUser spec contains a User type
type PvObjUser struct {
	PvObj
	Spec provision.User
}

// NewPvObjectStore
func NewPvObjectStore(backend string) *PvObjectStore {

	// Http Client Configuration for outbound connections
	netTransport := &http.Transport{
		MaxIdleConnsPerHost: 10,
		DialContext: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	httpClient := &http.Client{
		Timeout:   time.Second * 60,
		Transport: netTransport,
	}

	jsonTypeMap := map[string]map[string][]byte{}

	return &PvObjectStore{
		Backend:     backend,
		httpClient:  httpClient,
		jsonKindMap: jsonTypeMap,
	}
}

// PvObjectManager
type PvObjectStore struct {
	Backend    string
	httpClient *http.Client
	// object kind -> object id -> byte slice of json
	jsonKindMap map[string]map[string][]byte
	mux         sync.Mutex
}

// SendObjects
func (m *PvObjectStore) SendObjects() error {
	m.mux.Lock()
	defer m.mux.Unlock()

	for tp, jsBytesArr := range m.jsonKindMap {
		for id, jsBytesArr := range jsBytesArr {
			req, err := http.NewRequest("POST", m.Backend+"/"+tp, bytes.NewBuffer(jsBytesArr))
			if err != nil {
				return fmt.Errorf("cannot build resuest for %s object %s to %s: %s",
					tp,
					id,
					m.Backend+"/"+tp,
					err.Error(),
				)
			}

			req.Header.Set("Content-Kind", "application/json")
			resp, err := m.httpClient.Do(req)
			if err != nil {
				return fmt.Errorf("cannot post resuest for %s object %s to %s: %s",
					tp,
					id,
					m.Backend+"/"+tp,
					err.Error(),
				)
			}

			if resp.StatusCode != 200 {
				return fmt.Errorf("got status code %d for %s object %s posting to %s: %s",
					resp.StatusCode,
					tp,
					id,
					m.Backend+"/"+tp,
				)
			}

			fmt.Printf("Sent %s object %s.\n", tp, id)
		}

	}

	return nil
}

// ObjectsInStore
func (m *PvObjectStore) ObjectStore() map[string]map[string][]byte {
	return m.jsonKindMap
}

// ObjectsInStore
func (m *PvObjectStore) ObjectsInStore() int {
	count := 0
	for _, jsons := range m.jsonKindMap {
		count += len(jsons)
	}
	return count
}

// LoadObjects
func (m *PvObjectStore) LoadObjectsFromPath(path string) error {

	// is file or directory
	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("file %s stat error: %s", path, err.Error())
	}

	mode := fi.Mode()

	if mode.IsDir() {
		// load all yaml files in directory
		fmt.Printf("Reading directory: %s\n", path)
		dirname := path + string(filepath.Separator)
		d, err := os.Open(dirname)
		if err != nil {
			return fmt.Errorf("error ropening path %s: %s", path, err.Error())
		}

		files, err := d.Readdir(-1)
		if err != nil {
			return fmt.Errorf("error reading directory %s: %s", dirname, err.Error())
		}

		for _, file := range files {
			if file.Mode().IsRegular() {
				if filepath.Ext(file.Name()) == ".yml" || filepath.Ext(file.Name()) == ".yaml" {
					fmt.Printf("Reading file: %s%s\n", dirname, file.Name())
					err := m.LoadObjectsFromPath(dirname + file.Name())
					if err != nil {
						return err
					}
				}
			}
		}

		return nil
	}

	ymlData, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("file %s read error: %s", path, err.Error())
	}

	pvObj := &PvObj{}

	err = yaml.Unmarshal([]byte(ymlData), &pvObj)
	if err != nil {
		return fmt.Errorf("file %s unmarshal yaml error: %s", path, err.Error())
	}

	// check for type Asset, Account or User
	if pvObj.Kind == "Asset" {
		pvObjAsset := &PvObjAsset{}
		err = yaml.Unmarshal([]byte(ymlData), &pvObjAsset)
		if err != nil {
			return fmt.Errorf("file %s unmarshal Asset yaml error: %s", path, err.Error())
		}

		jsBytes, err := json.Marshal(pvObjAsset.Spec)
		if err != nil {
			return fmt.Errorf("file %s marshal Asset json error: %s", path, err.Error())
		}

		m.mux.Lock()
		if m.jsonKindMap["asset"] == nil {
			m.jsonKindMap["asset"] = map[string][]byte{}
		}

		m.jsonKindMap["asset"][pvObjAsset.Spec.Id] = jsBytes
		m.mux.Unlock()
	}

	return nil
}
