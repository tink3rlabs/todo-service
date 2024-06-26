package storage

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

var leaderElectionLock = &sync.Mutex{}
var leaderElectionInstance *LeaderElection

const DEFAULT_HEARTBEAT_INTERVAL = 60 * time.Second

// LeaderElection provides methods for electing a leader out of eligible cluster members
type LeaderElection struct {
	Id                string
	Leader            Member
	storageType       string
	storageProvider   string
	storage           StorageAdapter
	heartbeatInterval time.Duration
}

// Member represents a leadership eligible cluster node
type Member struct {
	Id           string
	Registration int64
	Heartbeat    int64
}

// NewLeaderElection creates an instance of a LeaderElection struct
func NewLeaderElection() *LeaderElection {
	if leaderElectionInstance == nil {
		leaderElectionLock.Lock()
		defer leaderElectionLock.Unlock()
		if leaderElectionInstance == nil {
			s := StorageAdapterFactory{}
			storageAdapter, err := s.GetInstance(DEFAULT)
			if err != nil {
				log.Fatalf("failed to create LeaderElection instance: %s", err.Error())
				return nil
			}
			heartbeatInterval := viper.GetDuration("leadership.heartbeat_interval")
			if heartbeatInterval == 0 {
				heartbeatInterval = DEFAULT_HEARTBEAT_INTERVAL
			}
			leaderElectionInstance = &LeaderElection{
				Id:                uuid.NewString(),
				storage:           storageAdapter,
				storageType:       viper.GetString("storage.type"),
				storageProvider:   viper.GetString("storage.provider"),
				heartbeatInterval: heartbeatInterval,
			}
		}
	}
	return leaderElectionInstance
}

// createLeadershipTable creates the database table used for leader election
func (l *LeaderElection) createLeadershipTable() error {
	var statement string
	switch l.storageProvider {
	case "postgresql":
		statement = "CREATE TABLE IF NOT EXISTS members (id TEXT PRIMARY KEY, registration NUMERIC, heartbeat NUMERIC)"
	case "mysql":
		statement = "CREATE TABLE IF NOT EXISTS members (id VARCHAR(50) PRIMARY KEY, registration BIGINT, heartbeat BIGINT)"
	case "sqlite":
		statement = "CREATE TABLE IF NOT EXISTS members (id TEXT PRIMARY KEY, registration INTEGER, heartbeat INTEGER)"
	}
	return l.storage.Execute(statement)
}

// updateMembershipTable updates the database table used for leader election
func (l *LeaderElection) updateMembershipTable() error {
	now := time.Now().UnixMilli()
	statement := fmt.Sprintf(`INSERT INTO members VALUES('%v', %v, %v)`, l.Id, now, now)
	return l.storage.Execute(statement)
}

// removeMember removes a cluster node from the database table used for leader election
func (l *LeaderElection) removeMember(memberId string) error {
	statement := fmt.Sprintf(`DELETE FROM members WHERE id='%v'`, memberId)
	return l.storage.Execute(statement)
}

// heartbeat is used by cluster members to indicate they are still alive
func (l *LeaderElection) heartbeat() {
	for {
		time.Sleep(l.heartbeatInterval)
		now := time.Now().UnixMilli()
		log.Printf("updating heartbeat to: %v", now)
		statement := fmt.Sprintf(`UPDATE members SET heartbeat='%v' WHERE id='%s'`, now, l.Id)
		err := l.storage.Execute(statement)
		if err != nil {
			log.Printf("failed to update heartbeat: %v", err)
		}
	}
}

// monitorLeader is a go routine that is used by cluster members to ensure the current leader is still active or trigger a re-election
func (l *LeaderElection) monitorLeader() {
	for {
		time.Sleep(l.heartbeatInterval / 2)
		acceptableInterval := -2 * l.heartbeatInterval

		leader, err := l.getLeader()
		if err != nil {
			log.Printf("error monitoring leader: %v", err)
		} else {
			diff := time.Until(time.UnixMilli(leader.Heartbeat))
			if diff >= acceptableInterval {
				log.Printf("leader %s is healthy", l.Leader.Id)
			} else {
				log.Printf("leader %s hasn't updated its heartbeat in %v starting reelection", l.Leader.Id, diff)
				err = l.electLeader(true)

				if err != nil {
					log.Printf("failed to elect new leader: %v", err)
				}

				if l.Id == l.Leader.Id {
					log.Println("I am the new leader")
					break
				} else {
					log.Printf("detected a change in leadership, new leader is %v - monitoring it", l.Leader.Id)
				}
			}
		}
	}
}

// electLeader is used to elect a leader from the list of eligible cluster members. It elects the active member with the earliest registration date as leader
func (l *LeaderElection) electLeader(reElection bool) error {
	log.Println("starting election process")
	leader := l.Leader

	if reElection {
		log.Println("this is a reelection removing existing leader")
		err := l.removeMember(l.Leader.Id)
		if err != nil {
			return fmt.Errorf("failed to remove leader from membership table: %v", err)
		}
		leader = Member{}
	}

	members, err := l.Members()
	if err != nil {
		return fmt.Errorf("failed to list leader eligible members: %v", err)
	}

	for _, m := range members {
		if leader == (Member{}) {
			// We don't have a leader set pick the current member for now
			leader = m
		}
		if m.Registration <= leader.Registration {
			leader = m
		}
	}
	l.Leader = leader
	return nil
}

// getLeader return the current active leader's record from the database
func (l *LeaderElection) getLeader() (Member, error) {
	var member Member
	var err error
	switch l.storageType {
	case "sql":
		statement := fmt.Sprintf(`SELECT * FROM members WHERE id='%s'`, l.Leader.Id)
		a := GetSQLAdapterInstance()
		result := a.DB.Raw(statement).Scan(&member)
		if result.Error != nil {
			err = fmt.Errorf("failed to get leader: %v", result.Error)
		}
	}
	return member, err
}

// Members returns a list of cluster members
func (l *LeaderElection) Members() ([]Member, error) {
	var members []Member
	var err error
	switch l.storageType {
	case "sql":
		statement := "SELECT * FROM members"
		a := GetSQLAdapterInstance()
		result := a.DB.Raw(statement).Scan(&members)
		if result.Error != nil {
			err = fmt.Errorf("failed to list cluster members: %v", result.Error)
		}
	}
	return members, err
}

// Start triggers a new leader election
func (l *LeaderElection) Start() {
	if l.storageType == "memory" {
		log.Println("using memory storage adapter, leader election is only supported with persistent storage")
	} else {
		log.Println("using a persistent storage adapter, starting leader election")
		log.Println("creating membership table")
		err := l.createLeadershipTable()
		if err != nil {
			log.Fatalf("failed to create membership table: %v", err)
		}
		log.Printf("registering node: %s", l.Id)
		err = l.updateMembershipTable()
		if err != nil {
			log.Fatalf("failed to register node: %v", err)
		}
		go l.heartbeat()
		err = l.electLeader(false)
		if err != nil {
			log.Fatalf("failed to elect leader: %v", err)
		}
		if l.Id == l.Leader.Id {
			log.Println("I was elected leader")
		} else {
			log.Printf("leader is %s - monitoring it", l.Leader.Id)
			go l.monitorLeader()
		}
	}
}
