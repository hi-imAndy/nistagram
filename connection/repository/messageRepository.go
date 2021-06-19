package repository

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j/dbtype"
	"nistagram/connection/model"
)

func (repo *Repository) CreateOrUpdateMessageRelationship(id1, id2 uint, connected bool) (*model.Message, bool) {
	session := (*repo.DatabaseDriver).NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()
	message := model.Message{
		PrimaryProfile:		id1,
		SecondaryProfile:	id2,
		Approved: 			connected,
	}
	resultingBlock, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
			"MATCH (a:Profile), (b:Profile) \n" +
				"WHERE a.profileID = $primary AND b.profileID = $secondary \n" +
				"MERGE (a)-[e:MESSAGE]->(b) " +
				"	ON CREATE SET e += { approved: $approved} \n" +
				"	ON MATCH SET e += { approved: $approved} \n" +
				"RETURN e",
			message.ToMap())
		var record *neo4j.Record
		if err != nil {
			return nil, err
		} else {
			record, err = result.Single()
			if err != nil {
				return nil, err
			}
		}
		res := record.Values[0].(dbtype.Relationship).Props
		fmt.Println(res)
		var ret = model.Message{
			PrimaryProfile:		id1,
			SecondaryProfile:	id2,
			Approved:			res["approved"].(bool),
		}
		return ret, err
	})
	if err != nil {
		fmt.Println(err.Error())
		return nil, false
	}
	var ret = resultingBlock.(model.Message)
	return &ret, true
}

func (repo *Repository) SelectMessage(id1, id2 uint) (*model.Message, bool) {
	session := (*repo.DatabaseDriver).NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()
	block := model.Message{
		PrimaryProfile:    id1,
		SecondaryProfile:  id2,
	}
	resultingBlock, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
			"MATCH (a:Profile)-[e:MESSAGE]->(b:Profile) \n"+
				"WHERE a.profileID = $primary AND b.profileID = $secondary \n"+
				"RETURN e",
			block.ToMap())
		var record *neo4j.Record
		if err != nil {
			return nil, err
		} else {
			record, err = result.Single()
			if err != nil {
				return nil, err
			}
		}
		res := record.Values[0].(dbtype.Relationship).Props
		fmt.Println(res)
		var ret = model.Message{
			PrimaryProfile:		id1,
			SecondaryProfile:	id2,
			Approved: 			res["approved"].(bool),
		}
		return ret, err
	})
	if err != nil {
		fmt.Println(err.Error())
		return nil, false
	}
	var ret = resultingBlock.(model.Message)
	return &ret, true
}

func (repo *Repository) DeleteMessage(followerId, profileId uint) (*model.Message, bool) {
	message, ok := repo.SelectMessage(followerId, profileId)
	if !ok {
		return nil, false
	}
	session := (*repo.DatabaseDriver).NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()
	_, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		return transaction.Run(
			"MATCH (a:Profile)-[e:MESSAGE]->(b:Profile) \n"+
				"WHERE a.profileID = $primary AND b.profileID = $secondary \n"+
				"DELETE e",
			message.ToMap())
	})
	if err != nil {
		return nil, false
	}
	return message, true
}