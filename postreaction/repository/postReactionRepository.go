package repository

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"nistagram/postreaction/model"
)

const reactionDbName = "postReactionDB"
const reactionsCollectionName = "reactions"
const reportsCollectionName = "reports"
const commentsCollectionName = "comments"

const profileIDColumn = "profileid"
const postIDColumn = "postid"
const reactionTypeColumn = "reactiontype"

var emptyContext = context.TODO()

type PostReactionRepository struct {
	Client *mongo.Client
}

func (repo *PostReactionRepository) ReactOnPost(reaction *model.Reaction) error {
	reactionsCollection := repo.getCollection(reactionsCollectionName)
	filter := bson.D{{profileIDColumn, reaction.ProfileID}, {postIDColumn, reaction.PostID}}
	var existingReaction model.Reaction
	exists := reactionsCollection.FindOne(emptyContext, filter).Decode(&existingReaction)
	if exists != nil {
		_, err := reactionsCollection.InsertOne(emptyContext, reaction)
		return err
	}
	update := bson.D{
		{"$set", bson.D{
			{reactionTypeColumn, reaction.ReactionType},
		}},
	}
	result, _ := reactionsCollection.UpdateOne(emptyContext, filter, update)
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (repo *PostReactionRepository) DeleteReaction(postID string, loggedUserID uint) error {
	reactionsCollection := repo.getCollection(reactionsCollectionName)
	filter := bson.D{{profileIDColumn, loggedUserID}, {postIDColumn, postID}}
	var existingReaction model.Reaction
	err := reactionsCollection.FindOne(emptyContext, filter).Decode(&existingReaction)
	if err != nil {
		return err
	}
	_, err = reactionsCollection.DeleteOne(emptyContext, existingReaction)
	return err
}

func (repo *PostReactionRepository) CommentPost(comment *model.Comment) error {
	commentsCollection := repo.getCollection(commentsCollectionName)
	_, err := commentsCollection.InsertOne(emptyContext, comment)
	return err
}

func (repo *PostReactionRepository) ReportPost(report *model.Report) error {
	reportsCollection := repo.getCollection(reportsCollectionName)
	_, err := reportsCollection.InsertOne(emptyContext, report)
	return err
}

func (repo *PostReactionRepository) GetProfileReactions(reactionType model.ReactionType, profileID uint) ([]model.Reaction, error) {
	reactionsCollection := repo.getCollection(reactionsCollectionName)
	filter := bson.D{{profileIDColumn, profileID}, {reactionTypeColumn, reactionType}}
	reactionCursor, err := reactionsCollection.Find(emptyContext, filter)
	if err != nil {
		return nil, err
	}
	var reactions []model.Reaction
	for reactionCursor.Next(emptyContext) {
		var result model.Reaction
		err = reactionCursor.Decode(&result)
		if err != nil {
			return nil, err
		}
		fmt.Println(result.ProfileID, result.ReactionType, result.PostID)
		reactions = append(reactions, result)
	}
	return reactions, nil
}

func (repo *PostReactionRepository) GetReactionType(profileID uint, postID string) string {
	reactionsCollection := repo.getCollection(reactionsCollectionName)
	filter := bson.D{{profileIDColumn, profileID}, {postIDColumn, postID}}
	var existingReaction model.Reaction
	err := reactionsCollection.FindOne(emptyContext, filter).Decode(&existingReaction)
	if err != nil {
		return "none"
	}
	return model.GetReactionTypeString(existingReaction.ReactionType)
}

func (repo *PostReactionRepository) GetAllReports() ([]model.Report, error) {
	var reports []model.Report
	reportsCollection := repo.getCollection(reportsCollectionName)
	filter := bson.D{{"isdeleted", false}}
	cursor, err := reportsCollection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	for cursor.Next(context.TODO()) {
		var result model.Report
		err = cursor.Decode(&result)
		if err != nil {
			return nil, err
		}
		reports = append(reports, result)
	}
	return reports, nil
}

func (repo *PostReactionRepository) GetReportsByPostId(postId string) ([]model.Report, error) {
	var reports []model.Report
	reportsCollection := repo.getCollection(reportsCollectionName)
	filter := bson.D{{"postid", postId}}
	cursor, err := reportsCollection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	for cursor.Next(context.TODO()) {
		var result model.Report
		err = cursor.Decode(&result)
		if err != nil {
			return nil, err
		}
		reports = append(reports, result)
	}
	return reports, nil
}

func (repo *PostReactionRepository) DeleteReport(reportId primitive.ObjectID) error {
	reportsCollection := repo.getCollection(reportsCollectionName)
	filter := bson.D{{"_id", reportId}}
	update := bson.D{
		{"$set", bson.D{
			{"isdeleted", true},
		}},
	}

	_, err := reportsCollection.UpdateOne(emptyContext, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (repo *PostReactionRepository) GetAllReactions(postID string) ([]uint, []uint, error) {
	reactionsCollection := repo.getCollection(reactionsCollectionName)
	filter := bson.D{{postIDColumn, postID}}
	reactionCursor, err := reactionsCollection.Find(emptyContext, filter)
	if err != nil {
		return nil, nil, err
	}
	likes := make([]uint, 0)
	dislikes := make([]uint, 0)
	for reactionCursor.Next(emptyContext) {
		var result model.Reaction
		err = reactionCursor.Decode(&result)
		if err != nil {
			return nil, nil, err
		}
		if result.ReactionType == model.LIKE {
			likes = append(likes, result.ProfileID)
		} else if result.ReactionType == model.DISLIKE {
			dislikes = append(dislikes, result.ProfileID)
		}
	}
	return likes, dislikes, nil
}

func (repo *PostReactionRepository) GetAllComments(postID string) ([]model.Comment, error) {
	commentsCollection := repo.getCollection(commentsCollectionName)
	filter := bson.D{{postIDColumn, postID}}
	commentsCursor, err := commentsCollection.Find(emptyContext, filter)
	if err != nil {
		return nil, err
	}
	comments := make([]model.Comment, 0)
	for commentsCursor.Next(emptyContext) {
		var result model.Comment
		err = commentsCursor.Decode(&result)
		if err != nil {
			return nil, err
		}
		comments = append(comments, result)
	}
	return comments, nil
}

func (repo *PostReactionRepository) getCollection(name string) *mongo.Collection {
	return repo.Client.Database(reactionDbName).Collection(name)
}
